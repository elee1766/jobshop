// Message attachments (chat composer): validation and the wire payload.
// The limits mirror the controller's: 8 files, 8 MiB in all, images as
// PNG, JPEG or GIF, other files UTF-8 text up to 2 MiB.
export interface Attachment {
  file: File;
  preview?: string;
}

const maxFiles = 8;
const maxTotal = 8 * 1024 * 1024;
const maxText = 2 * 1024 * 1024;

function checkTotals(existing: Attachment[], files: File[]) {
  const size = [...existing.map((item) => item.file), ...files].reduce((n, file) => n + file.size, 0);
  if (existing.length + files.length > maxFiles || size > maxTotal)
    throw new Error('Use at most 8 attachments totaling 8 MiB.');
}

// validate throws an Error with the message to show.
export async function validate(existing: Attachment[], files: File[]) {
  checkTotals(existing, files);
  for (const file of files) {
    if (!file.size) throw new Error('Empty files cannot be attached.');
    if (file.type.startsWith('image/')) {
      if (!['image/png', 'image/jpeg', 'image/gif'].includes(file.type))
        throw new Error('Use PNG, JPEG or GIF images.');
    } else {
      if (file.size > maxText) throw new Error('Text documents must be at most 2 MiB.');
      const content = new TextDecoder('utf-8', { fatal: true }).decode(await file.arrayBuffer());
      if (content.includes('\0')) throw new Error('Use UTF-8 text documents, not binary files.');
    }
  }
}

export function revoke(items: Attachment[]) {
  for (const item of items) if (item.preview) URL.revokeObjectURL(item.preview);
}

export function previewOf(file: File): Attachment {
  return { file, preview: file.type.startsWith('image/') ? URL.createObjectURL(file) : undefined };
}

export async function payload(items: Attachment[]) {
  return Promise.all(
    items.map(async ({ file }) => {
      const bytes = new Uint8Array(await file.arrayBuffer());
      let binary = '';
      for (let i = 0; i < bytes.length; i += 8192) binary += String.fromCharCode(...bytes.subarray(i, i + 8192));
      return {
        name: file.name,
        mime_type: file.type.startsWith('image/') ? file.type : 'text/plain',
        data: btoa(binary),
      };
    }),
  );
}

// Pasted text this large becomes a document, as before.
export const isLargePaste = (text: string) => text.length > 4000 || text.split('\n').length >= 40;
