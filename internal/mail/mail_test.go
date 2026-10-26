package mail

import (
	"strings"
	"testing"
	"time"
)

func fixedClock(t *testing.T) {
	t.Helper()
	was := at
	at = func() time.Time { return time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { at = was })
}

func TestAMessageCarriesBothForms(t *testing.T) {
	fixedClock(t)

	written := string(Compose(Settings{From: "chatcms@example.org", FromName: "Quiqr ChatCMS"},
		Message{To: "marieke@korenaar.example", Subject: "Hello", Text: "Plain words", HTML: "<p>Rich words</p>"}))

	for _, want := range []string{
		"From: Quiqr ChatCMS <chatcms@example.org>",
		"To: marieke@korenaar.example",
		"Subject: Hello",
		"multipart/alternative",
		"Plain words",
		"<p>Rich words</p>",
	} {
		if !strings.Contains(written, want) {
			t.Errorf("the message does not carry %q", want)
		}
	}
}

func TestANameWithAccentsIsEncoded(t *testing.T) {
	fixedClock(t)

	written := string(Compose(Settings{From: "chatcms@example.org", FromName: "Bakkerij Körenaar"},
		Message{To: "somebody@example.org", Subject: "Vers brood", Text: "Words"}))

	if strings.Contains(written, "Körenaar") {
		t.Error("a name with accents was sent as raw bytes, which a mail server may mangle")
	}
	if !strings.Contains(written, "=?utf-8?q?") {
		t.Errorf("the name was not encoded:\n%s", written)
	}
}

func TestAMessageWithoutHTMLIsPlain(t *testing.T) {
	fixedClock(t)

	written := string(Compose(Settings{From: "chatcms@example.org"},
		Message{To: "somebody@example.org", Subject: "Hello", Text: "Plain words"}))

	if strings.Contains(written, "multipart") {
		t.Error("a message with no HTML was sent as two parts")
	}
	if !strings.Contains(written, `text/plain; charset="utf-8"`) {
		t.Error("the message does not say it is plain text")
	}
}

func TestEveryLineEndsTheWayTheProtocolSays(t *testing.T) {
	fixedClock(t)

	written := string(Compose(Settings{From: "chatcms@example.org"},
		Message{To: "somebody@example.org", Subject: "Hello", Text: "one\ntwo\nthree"}))

	if strings.Contains(strings.ReplaceAll(written, "\r\n", ""), "\n") {
		t.Error("a line ends with a bare newline, which some servers mangle")
	}
}

func TestAHeaderCannotCarryAnotherHeader(t *testing.T) {
	fixedClock(t)

	// A display name holding a newline would otherwise let whoever set it
	// write headers of their own, a second recipient among them.
	written := string(Compose(
		Settings{From: "chatcms@example.org", FromName: "Trouble"},
		Message{
			To:      "somebody@example.org\r\nBcc: quiet@example.org",
			Subject: "Hello\r\nBcc: quieter@example.org",
			Text:    "Words",
		}))

	for _, line := range strings.Split(written, "\r\n") {
		if strings.HasPrefix(line, "Bcc:") {
			t.Errorf("a header was smuggled in:\n%s", written)
		}
	}
}

func TestAnAddressWithALineBreakIsRefused(t *testing.T) {
	// Flattening it would send to a strange address rather than to nobody,
	// and quietly sending somewhere unintended is worse than refusing.
	sender := NewSMTP(Settings{Host: "localhost", Port: "1025", From: "chatcms@example.org"})

	err := sender.Send(t.Context(), Message{To: "somebody@example.org\r\nBcc: quiet@example.org"})
	if err == nil || !strings.Contains(err.Error(), "address") {
		t.Errorf("Send() to an address with a line break error = %v, want a refusal naming the address", err)
	}
}

func TestNowhereToSendIsSaidPlainly(t *testing.T) {
	sender := NewSMTP(Settings{})
	if sender.Settings().Configured() {
		t.Error("a sender with no host says it is configured")
	}

	err := sender.Send(t.Context(), Message{To: "somebody@example.org"})
	if err != ErrNowhereToSend {
		t.Errorf("Send() with nowhere to send error = %v, want ErrNowhereToSend", err)
	}
}

func TestTheInvitationReadsForSomebodyWhoIsNotTechnical(t *testing.T) {
	message := Invitation("Korenaar Bakery", "Joris Bakker", "https://chatcms.example/set-password?token=abc")

	if !strings.Contains(message.Subject, "Korenaar Bakery") {
		t.Errorf("subject = %q, want it to name the site", message.Subject)
	}
	if !strings.Contains(message.Text, "Joris Bakker invited you") {
		t.Errorf("the message does not say who invited them:\n%s", message.Text)
	}
	if !strings.Contains(message.Text, "https://chatcms.example/set-password?token=abc") {
		t.Error("the message does not carry the link in its plain text")
	}

	forbidden := []string{"commit", "branch", "repository", "deploy", "push", "merge", "SHA", "HEAD"}
	for _, word := range forbidden {
		for _, body := range []string{message.Subject, message.Text, message.HTML} {
			if strings.Contains(strings.ToLower(body), strings.ToLower(word)) {
				t.Errorf("the invitation says %q, which an editor should never be shown", word)
			}
		}
	}
}

func TestAnInvitationWithoutNamesStillReads(t *testing.T) {
	message := Invitation("", "", "https://chatcms.example/set-password?token=abc")

	if strings.Contains(message.Text, "  ") || strings.Contains(message.Text, " .") {
		t.Errorf("the message has a gap where a name would be:\n%s", message.Text)
	}
	if !strings.Contains(message.Text, "invited") {
		t.Error("the message does not say what it is")
	}
}

func TestTheResetSaysHowLongItLasts(t *testing.T) {
	message := PasswordReset("https://chatcms.example/set-password?token=abc")

	if !strings.Contains(message.Text, "an hour") {
		t.Error("the message does not say how long the link works")
	}
	if !strings.Contains(message.Text, "has not changed") {
		t.Error("the message does not reassure somebody who did not ask")
	}
}

func TestALinkInHTMLCannotCloseItsOwnTag(t *testing.T) {
	message := Invitation("Bakery", "Joris", `https://example.org/"><script>alert(1)</script>`)

	if strings.Contains(message.HTML, "<script>") {
		t.Errorf("a link escaped its attribute:\n%s", message.HTML)
	}
}

func TestPasswordChangedNamesWhereToTurn(t *testing.T) {
	with := PasswordChanged("help@chatcms.example")
	if !strings.Contains(with.Text, "help@chatcms.example") {
		t.Error("the message does not say where to write")
	}

	without := PasswordChanged("")
	if strings.Contains(without.Text, "write to ") {
		t.Errorf("the message offers an address it does not have:\n%s", without.Text)
	}
}
