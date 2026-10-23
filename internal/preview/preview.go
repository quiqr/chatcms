// Package preview runs a Hugo server for one site so that somebody editing
// can see their work as a reader will. A preview is started when it is asked
// for and stopped when nobody has looked at it for a while, because an idle
// Hugo server is a process doing nothing but holding memory.
package preview

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var (
	// ErrTooMany says this machine is already running as many previews as it
	// will.
	ErrTooMany = errors.New("preview: too many sites are being previewed at once")
	// ErrNoHugo says no Hugo was named.
	ErrNoHugo = errors.New("preview: no Hugo was named")
	// ErrNotReady says the preview did not answer in time.
	ErrNotReady = errors.New("preview: the preview did not start")
)

// Settings are what previewing needs to know.
type Settings struct {
	// HugoBin is the pinned Hugo, the same one builds use.
	HugoBin string
	// Idle is how long a preview with nobody looking at it is kept.
	Idle time.Duration
	// Most is how many previews may run at once.
	Most int
	// Ready is how long to wait for a new preview to answer.
	Ready time.Duration
}

const (
	DefaultIdle  = 15 * time.Minute
	DefaultMost  = 8
	DefaultReady = 30 * time.Second
)

func (s Settings) withDefaults() Settings {
	if s.Idle <= 0 {
		s.Idle = DefaultIdle
	}
	if s.Most <= 0 {
		s.Most = DefaultMost
	}
	if s.Ready <= 0 {
		s.Ready = DefaultReady
	}
	return s
}

// running is one site's preview.
type running struct {
	id       string
	dir      string
	port     int
	command  *exec.Cmd
	proxy    *httputil.ReverseProxy
	lastSeen time.Time
}

// Servers keeps the previews. It is the one thing that knows a preview is a
// process at all; everything above it asks for a site and gets something that
// answers requests.
type Servers struct {
	settings Settings
	now      func() time.Time

	mu      sync.Mutex
	running map[string]*running
	stop    chan struct{}
	stopped bool
}

func NewServers(settings Settings) (*Servers, error) {
	settings = settings.withDefaults()
	if settings.HugoBin == "" {
		return nil, ErrNoHugo
	}

	servers := &Servers{
		settings: settings,
		now:      time.Now,
		running:  map[string]*running{},
		stop:     make(chan struct{}),
	}
	go servers.sweep()
	return servers, nil
}

// For gives something that answers requests for one site, starting a preview
// if there is not one already. The directory is the site's working copy.
func (s *Servers) For(ctx context.Context, id, dir string) (http.Handler, error) {
	s.mu.Lock()

	if held, found := s.running[id]; found && held.dir == dir {
		held.lastSeen = s.now()
		s.mu.Unlock()
		return held.proxy, nil
	}
	if held, found := s.running[id]; found {
		// The site moved, which means the working copy was made again. The
		// old preview is serving files that are no longer there.
		delete(s.running, id)
		s.mu.Unlock()
		stop(held)
		s.mu.Lock()
	}
	if len(s.running) >= s.settings.Most {
		s.mu.Unlock()
		return nil, ErrTooMany
	}
	s.mu.Unlock()

	started, err := s.start(ctx, id, dir)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		go stop(started)
		return nil, ErrNotReady
	}
	s.running[id] = started
	return started.proxy, nil
}

func (s *Servers) start(ctx context.Context, id, dir string) (*running, error) {
	port, err := freePort()
	if err != nil {
		return nil, err
	}

	// Hugo is told to bind only this machine: the preview is reached through
	// the app, which is what asks who somebody is. A preview open to the
	// network would be a customer's unpublished work with no sign-in at all.
	command := exec.Command(s.settings.HugoBin,
		"server",
		"--source", dir,
		"--bind", "127.0.0.1",
		"--port", strconv.Itoa(port),
		"--baseURL", "http://127.0.0.1:"+strconv.Itoa(port)+"/",
		"--appendPort=false",
		"--renderToMemory",
		"--buildDrafts",
		"--noBuildLock",
	)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Env = []string{"HOME=" + dir, "PATH=/usr/bin:/bin", "TZ=UTC"}

	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("preview: %w", err)
	}

	where, err := url.Parse("http://127.0.0.1:" + strconv.Itoa(port))
	if err != nil {
		kill(command)
		return nil, err
	}

	held := &running{
		id:       id,
		dir:      dir,
		port:     port,
		command:  command,
		proxy:    httputil.NewSingleHostReverseProxy(where),
		lastSeen: s.now(),
	}

	if err := waitFor(ctx, port, s.settings.Ready); err != nil {
		stop(held)
		return nil, err
	}
	return held, nil
}

// waitFor waits for the preview to answer. A new site takes a moment to
// render, and answering before it does would show a person an error rather
// than their site.
func waitFor(ctx context.Context, port int, within time.Duration) error {
	deadline := time.Now().Add(within)
	address := "127.0.0.1:" + strconv.Itoa(port)

	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		connection, err := net.DialTimeout("tcp", address, time.Second)
		if err == nil {
			connection.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return ErrNotReady
}

func freePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("preview: %w", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// Stop ends one site's preview, which is what happens when the site changes
// in a way the preview cannot follow.
func (s *Servers) Stop(id string) {
	s.mu.Lock()
	held, found := s.running[id]
	delete(s.running, id)
	s.mu.Unlock()

	if found {
		stop(held)
	}
}

// Running is how many previews are going, which is what the cap is about.
func (s *Servers) Running() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.running)
}

// Close stops every preview. A Hugo server outliving the app would hold a
// port and a working copy with nothing to end it.
func (s *Servers) Close() {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	close(s.stop)

	held := make([]*running, 0, len(s.running))
	for _, one := range s.running {
		held = append(held, one)
	}
	s.running = map[string]*running{}
	s.mu.Unlock()

	for _, one := range held {
		stop(one)
	}
}

// sweep stops previews nobody has looked at. It is a loop rather than a timer
// per preview so that there is one thing to stop when the app does.
func (s *Servers) sweep() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.sweepOnce()
		}
	}
}

func (s *Servers) sweepOnce() {
	s.mu.Lock()
	idle := []*running{}
	for id, held := range s.running {
		if s.now().Sub(held.lastSeen) > s.settings.Idle {
			idle = append(idle, held)
			delete(s.running, id)
		}
	}
	s.mu.Unlock()

	for _, held := range idle {
		stop(held)
	}
}

func stop(held *running) {
	if held == nil || held.command == nil {
		return
	}
	kill(held.command)
}

// kill ends Hugo and anything it started. Hugo watches files through
// children of its own, so signalling the group is what actually stops it.
func kill(command *exec.Cmd) {
	if command.Process == nil {
		return
	}
	syscall.Kill(-command.Process.Pid, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		command.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		<-done
	}
}

// SweepNow runs the idle check at once, which is how a test reaches it
// without waiting for the clock.
func (s *Servers) SweepNow() { s.sweepOnce() }

// PretendTimeIs makes the idle check believe a different moment, so that
// "nobody has looked at this for a quarter of an hour" can be tested in
// milliseconds.
func (s *Servers) PretendTimeIs(now func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.now = now
}
