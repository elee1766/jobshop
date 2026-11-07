import { test, expect } from '@playwright/test';
import { randomBytes } from 'node:crypto';
import { writeFile, unlink } from 'node:fs/promises';
import { join } from 'node:path';

test.skip(!process.env.DESK_ATTACHMENTS, 'set DESK_ATTACHMENTS=1 with a real image-capable operator agent');

test('pasted image and document reach the real agent as native attachments', async ({ page, isMobile }) => {
  test.setTimeout(240_000);
  const name = 'attachments-' + randomBytes(5).toString('hex');
  const visual = randomBytes(4).toString('hex').toUpperCase();
  const secret = 'document-' + randomBytes(4).toString('hex');
  const document = 'Document verification\n' + 'Unrelated document content.\n'.repeat(180) + '\nDocument token: ' + secret;
  let id;
  const outputPath = process.env.DESK_IMAGE_OUTPUT_DIR ? join(process.env.DESK_IMAGE_OUTPUT_DIR, name + '.png') : null;
  try {
    await page.goto('./');
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'Chats', exact: true }).click();
    await page.getByRole('button', { name: 'New', exact: true }).click();
    await page.getByRole('textbox', { name: 'Name', exact: true }).fill(name);
    await page.getByRole('textbox', { name: 'Name', exact: true }).press('Enter');
    const composer = page.getByRole('textbox', { name: /^First message to new chat:/ });
    await composer.fill('Read the token printed in the attached image and the document token from the attached file. Use your file-reading tool for the resource link, decoding its file URL if necessary. Reply with both exact tokens.');
    const png = await composer.evaluate((element, { visual, document }) => {
      const canvas = window.document.createElement('canvas');
      canvas.width = 640;
      canvas.height = 160;
      const ctx = canvas.getContext('2d');
      ctx.fillStyle = 'white';
      ctx.fillRect(0, 0, 640, 160);
      ctx.fillStyle = 'black';
      ctx.font = 'bold 72px monospace';
      ctx.fillText(visual, 30, 110);
      const binary = atob(canvas.toDataURL('image/png').split(',')[1]);
      const data = Uint8Array.from(binary, c => c.charCodeAt(0));
      const imageClipboard = new DataTransfer();
      imageClipboard.items.add(new File([data], 'clipboard.png', { type: 'image/png' }));
      element.dispatchEvent(new ClipboardEvent('paste', { clipboardData: imageClipboard, bubbles: true, cancelable: true }));
      const textClipboard = new DataTransfer();
      textClipboard.setData('text/plain', document);
      element.dispatchEvent(new ClipboardEvent('paste', { clipboardData: textClipboard, bubbles: true, cancelable: true }));
      return canvas.toDataURL('image/png').split(',')[1];
    }, { visual, document });
    if (outputPath) {
      await writeFile(outputPath, Buffer.from(png, 'base64'));
      await composer.fill(await composer.inputValue() + '\nAlso use the read tool to inspect this image file: ' + outputPath);
    }
    await expect(page.getByRole('list', { name: 'Message attachments' }).locator('li')).toHaveCount(2);
    await expect(composer).not.toHaveValue(new RegExp(secret));
    const created = page.waitForResponse(response => response.url().endsWith('/api/chat/manage') && response.request().method() === 'POST');
    await composer.press('Enter');
    const response = await created;
    expect(response.ok(), await response.text()).toBeTruthy();
    id = (await response.json()).id;
    await expect(page.getByRole('list', { name: 'Message attachments' })).toHaveCount(0);
    await expect(page.locator('.transcript .markdown')).toContainText(visual, { timeout: 180_000 });
    await expect(page.locator('.transcript .markdown')).toContainText(secret, { timeout: 180_000 });
    const sent = page.getByRole('img', { name: 'User attachment', exact: true });
    await expect(sent).toBeVisible();
    expect(await sent.evaluate(img => img.naturalWidth)).toBeGreaterThan(0);
    if (outputPath) {
      await expect(page.getByRole('img', { name: 'Agent output', exact: true })).toBeVisible({ timeout: 30_000 });
    }
    await page.reload();
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'Chats', exact: true }).click();
    await page.getByRole('button', { name: 'Choose chat', exact: true }).click();
    await page.getByRole('searchbox', { name: 'Search chats', exact: true }).fill(name);
    await page.getByRole('table', { name: 'Recent chats', exact: true }).getByRole('button', { name, exact: true }).click();
    await expect(sent).toBeVisible();
    if (outputPath) await expect(page.getByRole('img', { name: 'Agent output', exact: true })).toBeVisible();
  } finally {
    if (outputPath) await unlink(outputPath).catch(() => {});
    if (id) {
      const response = await page.request.post('api/chat/manage', { headers: { 'X-Jobshop-Desk': '1' }, data: { action: 'archive', id } });
      expect(response.ok()).toBeTruthy();
    }
  }
});
