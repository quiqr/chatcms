package mail

import (
	"fmt"
	"html"
	"strings"
)

// These are the only messages ChatCMS sends. They are written for a content
// creator, in the language docs/design/DESIGN.md section 7 asks for: no git
// words, nothing about repositories, nothing a person would have to look up.

// Invitation is the message somebody gets when they are asked to join.
func Invitation(siteName, invitedBy, link string) Message {
	intro := "You have been invited to edit a website."
	if siteName != "" && invitedBy != "" {
		intro = fmt.Sprintf("%s invited you to help edit %s.", invitedBy, siteName)
	} else if siteName != "" {
		intro = fmt.Sprintf("You have been invited to help edit %s.", siteName)
	}

	text := join(
		intro,
		"",
		"To get started, choose a password:",
		link,
		"",
		"The link works for a week. If it has run out, ask whoever invited you to send another.",
		"",
		"If you were not expecting this, you can ignore it. Nothing happens until you use the link.",
	)

	body := paragraphs(
		escape(intro),
		button(link, "Choose a password"),
		"The link works for a week. If it has run out, ask whoever invited you to send another.",
		"If you were not expecting this, you can ignore it. Nothing happens until you use the link.",
	)

	subject := "You have been invited to edit a website"
	if siteName != "" {
		subject = "You have been invited to edit " + siteName
	}

	return Message{Subject: subject, Text: text, HTML: body}
}

// PasswordReset is the message somebody gets when they ask for a new password.
func PasswordReset(link string) Message {
	text := join(
		"You asked to set a new password.",
		"",
		"Choose one here:",
		link,
		"",
		"The link works for an hour, and only once.",
		"",
		"If you did not ask for this, you can ignore it. Your password has not changed.",
	)

	body := paragraphs(
		"You asked to set a new password.",
		button(link, "Choose a new password"),
		"The link works for an hour, and only once.",
		"If you did not ask for this, you can ignore it. Your password has not changed.",
	)

	return Message{Subject: "Set a new password", Text: text, HTML: body}
}

// PasswordChanged tells somebody their password changed, which is how they
// find out if it was not them.
func PasswordChanged(helpAddress string) Message {
	closing := "If this was not you, ask whoever looks after your website straight away."
	if helpAddress != "" {
		closing = fmt.Sprintf("If this was not you, write to %s straight away.", helpAddress)
	}

	text := join("Your password has been changed.", "", closing)
	return Message{
		Subject: "Your password has been changed",
		Text:    text,
		HTML:    paragraphs("Your password has been changed.", escape(closing)),
	}
}

func join(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

func escape(text string) string { return html.EscapeString(text) }

// button is a link that looks like one without needing a stylesheet, because
// a mail program will not fetch one.
func button(link, label string) string {
	safe := escape(link)
	return fmt.Sprintf(
		`<p><a href="%s" style="display:inline-block;padding:10px 16px;`+
			`background:#0b5f8a;color:#ffffff;text-decoration:none;border-radius:6px">%s</a></p>`+
			`<p style="font-size:13px;color:#555">Or copy this address: %s</p>`,
		safe, escape(label), safe)
}

func paragraphs(parts ...string) string {
	var out strings.Builder
	out.WriteString(`<div style="font-family:system-ui,sans-serif;font-size:15px;line-height:1.5;color:#1a1a1a">`)
	for _, part := range parts {
		if strings.HasPrefix(part, "<p") {
			out.WriteString(part)
			continue
		}
		out.WriteString("<p>" + part + "</p>")
	}
	out.WriteString("</div>")
	return out.String()
}
