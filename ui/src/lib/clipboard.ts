// Copy text to the clipboard and say whether it worked (JOB-136).
//
// navigator.clipboard needs a secure context and a user gesture; Desk is
// often served over plain http on a LAN, so fall back to a hidden textarea
// and execCommand. Focus goes back to where it was either way.
export async function copyText(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch {
    // Denied or unavailable: try the fallback below.
  }
  const active = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  const area = document.createElement('textarea');
  area.value = text;
  area.setAttribute('readonly', '');
  area.setAttribute('aria-hidden', 'true');
  area.style.cssText = 'position:fixed;top:0;left:0;opacity:0;pointer-events:none';
  document.body.append(area);
  area.select();
  let ok = false;
  try {
    ok = document.execCommand('copy');
  } catch {
    ok = false;
  }
  area.remove();
  if (active?.isConnected) active.focus({ preventScroll: true });
  return ok;
}
