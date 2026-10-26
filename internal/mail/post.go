package mail

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// JobKind is the kind of job that sends one message.
const JobKind = "send-mail"

// Posting is what the rest of the application depends on: somewhere to put a
// message so that it is sent soon, rather than sent now while somebody waits.
type Posting interface {
	Post(ctx context.Context, to string, message Message) error
}

// Queue is the piece of the job queue mail needs. It is named here rather
// than imported so that the queue does not have to know about mail.
type Queue interface {
	Enqueue(ctx context.Context, kind, subject string, payload any) (int64, error)
}

// Letter is one queued message.
type Letter struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Text    string `json:"text"`
	HTML    string `json:"html"`
}

// Post puts a message on the queue. Sending in the request would make
// somebody wait on a server that is not ours, and would lose the invitation
// if that server were slow or refusing.
type Post struct {
	queue  Queue
	logger *slog.Logger
	// configured says whether there is anywhere to send. When there is not,
	// the link is written to the log instead, which is what a single-tenant
	// install run by its owner needs.
	configured bool
	fallback   func(ctx context.Context, to string, message Message)
}

func NewPost(queue Queue, configured bool, logger *slog.Logger, fallback func(context.Context, string, Message)) *Post {
	return &Post{queue: queue, configured: configured, logger: logger, fallback: fallback}
}

func (p *Post) Post(ctx context.Context, to string, message Message) error {
	if !p.configured {
		if p.fallback != nil {
			p.fallback(ctx, to, message)
		}
		return nil
	}

	message.To = to
	if _, err := p.queue.Enqueue(ctx, JobKind, "", Letter{
		To:      message.To,
		Subject: message.Subject,
		Text:    message.Text,
		HTML:    message.HTML,
	}); err != nil {
		return fmt.Errorf("the message could not be queued: %w", err)
	}

	p.logger.InfoContext(ctx, "a message was queued", "to", to, "subject", message.Subject)
	return nil
}

// Deliver sends one queued message. It is given the sender rather than
// making one, so that a test watches what was sent instead of posting it.
//
// A failure is returned, so the queue tries again: a mail server that is down
// for a minute should not lose somebody's invitation. Having nowhere to send
// at all is the exception, because it will still be nowhere in five minutes
// and a job that can never succeed buries the ones that would.
func Deliver(ctx context.Context, sender Sender, letter Letter, logger *slog.Logger) error {
	err := sender.Send(ctx, Message{
		To:      letter.To,
		Subject: letter.Subject,
		Text:    letter.Text,
		HTML:    letter.HTML,
	})

	switch {
	case err == nil:
		logger.InfoContext(ctx, "a message was sent", "to", letter.To, "subject", letter.Subject)
		return nil
	case errors.Is(err, ErrNowhereToSend):
		logger.WarnContext(ctx, "a message was dropped because there is nowhere to send it",
			"to", letter.To, "subject", letter.Subject)
		return nil
	default:
		return err
	}
}
