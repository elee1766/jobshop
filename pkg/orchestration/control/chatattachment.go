package control

import (
	"github.com/gfx-labs/jobshop/pkg/conversation"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// The limits, layout and prompt blocks of message attachments are in
// pkg/conversation (JOB-171). The controller owns the root directory,
// redaction and reaping.

// attachmentDir keeps files outside agent-writable session mounts; they
// are reaped with the session.
func (c *Controller) attachmentDir(session string) string {
	return conversation.AttachmentDir(c.root.Dir, session)
}

// saveAttachments writes files for a message of session, masking secrets
// in text documents first.
func (c *Controller) saveAttachments(sc scope, session, message string, files []rpc.AttachmentContent) ([]domain.MessageAttachment, error) {
	return conversation.SaveAttachments(c.attachmentDir(session), message, files, func(text *string) {
		c.redactWrite(sc, "chat attachment", message, text)
	})
}
