package conversation

import (
	"bytes"
	"encoding/xml"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Messages reach the model as XML envelopes (#1073):
//
//	<messages>
//	<message from="admin" id="js-m1">title\nbody</message>
//	</messages>
//
// Sender, id, title, body and attachment lines are user input, so they go
// through encoding/xml's escaping: a closing tag, `&` or a quote in them
// cannot end the element or the attribute early.
const (
	messagesOpen  = "<messages>"
	messagesClose = "</messages>"
	// EnvelopePlaceholder stands in for the envelope in the session.prompt
	// preview, before any message exists.
	EnvelopePlaceholder = messagesOpen + "\n(next messages)\n" + messagesClose
)

// xmlEscape escapes s for element text or an attribute value.
func xmlEscape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

// xmlBody escapes s for element text. encoding/xml writes newline and tab
// as character references, which the model would read as noise; a literal
// one is valid XML text. A user's own "&#xA;" is already "&amp;#xA;" here,
// so only the references EscapeText made are restored.
func xmlBody(s string) string {
	return strings.NewReplacer("&#xA;", "\n", "&#x9;", "\t").Replace(xmlEscape(s))
}

// MessageElement is one <message> element.
func MessageElement(from, id, text string) string {
	return `<message from="` + xmlEscape(from) + `" id="` + xmlEscape(id) + `">` + xmlBody(text) + `</message>`
}

// Envelope wraps message elements in <messages>. With none (every
// message was context only) it is empty, so the turn carries no message part.
func Envelope(elements []string) string {
	if len(elements) == 0 {
		return ""
	}
	return messagesOpen + "\n" + strings.Join(elements, "\n") + "\n" + messagesClose
}

// Recipients are the addresses a session reads its messages at: its id,
// its key when it has one, and its pool as "pool:<name>".
func Recipients(id, pool, key string) map[string]bool {
	to := map[string]bool{id: true}
	if pool != "" {
		to["pool:"+pool] = true
	}
	if key != "" {
		to[key] = true
	}
	return to
}

// Addressed keeps the messages in msgs sent to one of to, in order.
// interruptOnly keeps only interrupt messages. A message whose fields do
// not decode is skipped.
func Addressed(msgs []domain.Ticket, to map[string]bool, interruptOnly bool) []domain.Ticket {
	var out []domain.Ticket
	for _, m := range msgs {
		f, err := domain.DecodeFields[domain.MessageFields](m)
		if err != nil || !to[f.To] || (interruptOnly && !f.Interrupt) {
			continue
		}
		out = append(out, m)
	}
	return out
}

// FitsBudget reports whether a message with size attachment bytes may join
// a batch that already holds claimed messages and used attachment bytes.
// The first message always fits, so one large message is still delivered.
func FitsBudget(claimed, used, size int) bool {
	return claimed == 0 || used+size <= rpc.MaxAttachmentBytes
}
