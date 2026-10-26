// Package mail sends the few messages ChatCMS has to send: an invitation, a
// password reset, and word that a password was changed. It speaks SMTP, and
// says plainly when it has nowhere to send to.
package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// ErrNowhereToSend says no mail server is configured, which is not a failure
// on a single-tenant install run by the person who owns it.
var ErrNowhereToSend = errors.New("no mail server is configured, so nothing can be sent")

// Message is one message, in both the forms a reader's program may want.
type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// Settings is where to send and as whom.
type Settings struct {
	Host     string
	Port     string
	Username string
	Password string
	// From is the address messages come from, which a receiving server checks
	// against the domain it was told to accept.
	From string
	// FromName is what a reader sees instead of the address.
	FromName string
	// StartTLS asks for the connection to be upgraded. It is on for a real
	// server and off for a development one, which has no certificate.
	StartTLS bool
	// Timeout bounds the whole exchange, so a server that accepts the
	// connection and then says nothing cannot hold a worker forever.
	Timeout time.Duration
}

// Configured says whether there is anywhere to send.
func (s Settings) Configured() bool {
	return strings.TrimSpace(s.Host) != ""
}

func (s Settings) address() string {
	port := strings.TrimSpace(s.Port)
	if port == "" {
		port = "25"
	}
	return net.JoinHostPort(strings.TrimSpace(s.Host), port)
}

// Sender sends a message. It is an interface so that what the rest of the
// application depends on is "something that sends", and the tests can watch
// what was sent without a server.
type Sender interface {
	Send(ctx context.Context, message Message) error
}

// SMTP sends over SMTP.
type SMTP struct {
	settings Settings
}

func NewSMTP(settings Settings) *SMTP {
	if settings.Timeout <= 0 {
		settings.Timeout = 30 * time.Second
	}
	return &SMTP{settings: settings}
}

// Settings is what this sender was built with.
func (s *SMTP) Settings() Settings { return s.settings }

func (s *SMTP) Send(ctx context.Context, message Message) error {
	if !s.settings.Configured() {
		return ErrNowhereToSend
	}
	if strings.TrimSpace(message.To) == "" {
		return errors.New("a message needs somewhere to go")
	}
	if strings.ContainsAny(message.To, "\r\n") {
		// Flattening it would send to a strange address rather than to
		// nobody, and quietly sending somewhere unintended is worse than
		// refusing outright.
		return errors.New("that address has a line break in it, so it is not an address")
	}

	dialer := net.Dialer{Timeout: s.settings.Timeout}
	connection, err := dialer.DialContext(ctx, "tcp", s.settings.address())
	if err != nil {
		return fmt.Errorf("the mail server could not be reached: %w", err)
	}
	defer connection.Close()

	// A deadline on the socket as well as on the dial: a server that accepts
	// and then goes quiet is the case a context alone does not cover.
	if deadline, ok := ctx.Deadline(); ok {
		_ = connection.SetDeadline(deadline)
	} else {
		_ = connection.SetDeadline(time.Now().Add(s.settings.Timeout))
	}

	client, err := smtp.NewClient(connection, s.settings.Host)
	if err != nil {
		return fmt.Errorf("the mail server did not answer as one: %w", err)
	}
	defer client.Close()

	if s.settings.StartTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return errors.New("the mail server does not offer an encrypted connection, and one was asked for")
		}
		if err := client.StartTLS(&tls.Config{ServerName: s.settings.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("securing the connection failed: %w", err)
		}
	}

	if s.settings.Username != "" {
		auth := smtp.PlainAuth("", s.settings.Username, s.settings.Password, s.settings.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("the mail server refused the credentials: %w", err)
		}
	}

	if err := client.Mail(s.settings.From); err != nil {
		return fmt.Errorf("the mail server refused the sender: %w", err)
	}
	if err := client.Rcpt(message.To); err != nil {
		return fmt.Errorf("the mail server refused the recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("the mail server would not take the message: %w", err)
	}
	if _, err := writer.Write(Compose(s.settings, message)); err != nil {
		return fmt.Errorf("writing the message failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("the mail server did not accept the message: %w", err)
	}

	return client.Quit()
}
