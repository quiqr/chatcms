package gitsync

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/quiqr/chatcms/internal/redact"
)

// Config is what the engine needs to reach git and a remote.
type Config struct {
	// GitBin is the git binary to run. It is never looked up on PATH.
	GitBin string
	// SSHKeyPath is the private key used to reach the remote over SSH. An
	// empty value means the remote needs no key, which is how the tests run.
	SSHKeyPath string
	// System is recorded as the agent that wrote every version; the person
	// who made the edit is recorded as its author.
	System Identity
}

// Identity is a person or an agent a version is attributed to.
type Identity struct {
	Name  string
	Email string
}

func (i Identity) String() string {
	return fmt.Sprintf("%s <%s>", i.Name, i.Email)
}

func (i Identity) valid() bool {
	return i.Name != "" && i.Email != ""
}

// deterministicSettings make git behave the same on every machine, never ask
// a question nobody is there to answer, and never run code that came out of a
// customer's repository. See design.md.
func deterministicSettings() []string {
	return []string{
		"-c", "core.hooksPath=/dev/null",
		"-c", "credential.helper=",
		"-c", "core.quotePath=false",
		"-c", "advice.detachedHead=false",
		"-c", "gc.auto=0",
		"-c", "rerere.enabled=false",
		"-c", "core.autocrlf=false",
		"-c", "core.fsmonitor=false",
		"-c", "protocol.file.allow=always",
	}
}

type runner struct {
	bin  string
	dir  string
	env  []string
	home string
}

func newRunner(cfg Config, dir, home string) *runner {
	env := []string{
		"GIT_TERMINAL_PROMPT=0",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_ASKPASS=",
		"SSH_ASKPASS=",
		"HOME=" + home,
		"XDG_CONFIG_HOME=" + home,
		"LC_ALL=C",
		"TZ=UTC",
		// A shell no-op rather than a binary, because git starts the editor
		// through a shell and the environment here is deliberately bare.
		"GIT_EDITOR=:",
		// git itself is run by absolute path, but it starts ssh and its own
		// helpers by name, so they need somewhere to be found.
		"PATH=" + os.Getenv("PATH"),
		"GIT_SSH_COMMAND=" + sshCommand(cfg.SSHKeyPath),
		// Anything git writes is written by the system. A save overrides the
		// author so the version is attributed to the person who made the edit;
		// a replay keeps each version's original author by itself.
		"GIT_COMMITTER_NAME=" + cfg.System.Name,
		"GIT_COMMITTER_EMAIL=" + cfg.System.Email,
		"GIT_AUTHOR_NAME=" + cfg.System.Name,
		"GIT_AUTHOR_EMAIL=" + cfg.System.Email,
	}
	return &runner{bin: cfg.GitBin, dir: dir, env: env, home: home}
}

func sshCommand(keyPath string) string {
	command := "ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new"
	if keyPath != "" {
		command += " -o IdentitiesOnly=yes -i " + keyPath
	}
	return command
}

// run executes git and returns its standard output. A failure carries git's
// own message with anything credential-shaped removed.
func (r *runner) run(ctx context.Context, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, r.bin, append(deterministicSettings(), args...)...)
	cmd.Dir = r.dir
	cmd.Env = r.env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stdout.String(), &OperationError{
			Operation: args[0],
			Detail:    redact.Secrets(strings.TrimSpace(stderr.String() + " " + stdout.String())),
			Err:       err,
		}
	}
	return stdout.String(), nil
}

func (r *runner) nulFields(ctx context.Context, args ...string) ([]string, error) {
	out, err := r.run(ctx, args...)
	if err != nil {
		return nil, err
	}
	return splitNul(out), nil
}

func (r *runner) succeeds(ctx context.Context, args ...string) bool {
	_, err := r.run(ctx, args...)
	return err == nil
}

func splitLines(out string) []string {
	trimmed := strings.TrimRight(out, "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

func splitNul(out string) []string {
	trimmed := strings.TrimRight(out, "\x00")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\x00")
}
