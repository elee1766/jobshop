package control

import (
	"bytes"
	"encoding/xml"
	"strings"
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
	// messagesPlaceholder stands in for the envelope in the session.prompt
	// preview, before any message exists.
	messagesPlaceholder = messagesOpen + "\n(next messages)\n" + messagesClose
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

// messageElement is one <message> element.
func messageElement(from, id, text string) string {
	return `<message from="` + xmlEscape(from) + `" id="` + xmlEscape(id) + `">` + xmlBody(text) + `</message>`
}

// messagesEnvelope wraps message elements in <messages>. With none (every
// message was context only) it is empty, so the turn carries no message part.
func messagesEnvelope(elements []string) string {
	if len(elements) == 0 {
		return ""
	}
	return messagesOpen + "\n" + strings.Join(elements, "\n") + "\n" + messagesClose
}
