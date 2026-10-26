package mail

import (
	"bytes"
	"fmt"
	"mime"
	"strings"
	"time"
)

// Compose writes the message as a mail server expects it: headers, then a
// plain text part and an HTML one, so that a reader's program can show
// whichever it prefers. A message with no HTML is sent as text alone.
//
// Lines are joined with CRLF because that is what the protocol says, and a
// message joined with bare newlines is one some servers quietly mangle.
func Compose(settings Settings, message Message) []byte {
	var out bytes.Buffer

	from := settings.From
	if settings.FromName != "" {
		from = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", settings.FromName), settings.From)
	}

	header(&out, "From", from)
	header(&out, "To", message.To)
	header(&out, "Subject", mime.QEncoding.Encode("utf-8", message.Subject))
	header(&out, "Date", at().Format(time.RFC1123Z))
	header(&out, "MIME-Version", "1.0")

	if strings.TrimSpace(message.HTML) == "" {
		header(&out, "Content-Type", `text/plain; charset="utf-8"`)
		out.WriteString("\r\n")
		out.WriteString(crlf(message.Text))
		return out.Bytes()
	}

	// A boundary that cannot appear in either part, because it is not
	// something a person would write.
	const boundary = "chatcms-part-boundary-8f2a1c"

	header(&out, "Content-Type", `multipart/alternative; boundary="`+boundary+`"`)
	out.WriteString("\r\n")

	out.WriteString("--" + boundary + "\r\n")
	out.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
	out.WriteString(crlf(message.Text))
	out.WriteString("\r\n")

	out.WriteString("--" + boundary + "\r\n")
	out.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
	out.WriteString(crlf(message.HTML))
	out.WriteString("\r\n")

	out.WriteString("--" + boundary + "--\r\n")
	return out.Bytes()
}

func header(out *bytes.Buffer, name, value string) {
	// A newline in a header value would let whatever produced it write
	// headers of its own, which is how a display name becomes a second
	// recipient nobody asked for.
	clean := strings.NewReplacer("\r", " ", "\n", " ").Replace(value)
	out.WriteString(name + ": " + clean + "\r\n")
}

func crlf(body string) string {
	normalised := strings.ReplaceAll(body, "\r\n", "\n")
	return strings.ReplaceAll(normalised, "\n", "\r\n")
}

// at is the clock, which the tests replace so that a composed message can be
// compared with a known one.
var at = time.Now
