import { useAtomValue } from 'jotai';
import type { EmbedView } from '../../../packages/jobshop-client/index.js';
import { useDeskRoute } from '../lib/route.ts';
import { goSection } from '../state/actions.ts';
import { embedsAtom } from '../state/live.ts';

// Embedded apps (JOB-151). An embed is an operator-configured iframe tab. The
// route names an id and the id is looked up in the configuration the
// controller sent: a URL in the address bar never becomes an iframe source.
//
// Only the selected embed is mounted. Leaving the view unmounts the iframe, so
// a hidden app keeps no timers, sockets or memory, and coming back reloads it.
//
// The iframe is a boundary, not a bridge: Desk passes no token, cookie, state
// or message to it, and listens for none from it.

// The origins the page's frame-src allowed when it loaded (a meta tag the
// server adds with the policy). A config reload can add an embed after that.
function allowedOrigins(): string[] {
  const meta = document.querySelector<HTMLMetaElement>('meta[name="jobshop-embed-origins"]');
  return meta?.content.split(' ').filter(Boolean) ?? [];
}

// Only http(s) reaches an iframe or an anchor, whatever the server sent.
function origin(url: string): string {
  try {
    const parsed = new URL(url);
    return parsed.protocol === 'https:' || parsed.protocol === 'http:' ? parsed.origin : '';
  } catch {
    return '';
  }
}

export function Embed() {
  const route = useDeskRoute();
  const embeds = useAtomValue(embedsAtom);
  const embed = embeds?.find((e) => e.id === route.embed);

  if (!embeds) {
    return (
      <>
        <div className="tag">
          <span className="mono">/embed</span>
        </div>
        <p className="section-body muted">loading</p>
      </>
    );
  }
  // A removed or mistyped id renders no frame at all.
  if (!embed || !origin(embed.url)) {
    return (
      <>
        <div className="tag">
          <span className="mono">/embed</span>
        </div>
        <div className="section-body">
          <p>{route.embed ? `No embedded app "${route.embed}" is configured.` : 'No embedded app is selected.'}</p>
          <p>
            <a
              href="./"
              onClick={(event) => {
                event.preventDefault();
                goSection('');
              }}
            >
              back to home
            </a>
          </p>
        </div>
      </>
    );
  }
  return <Frame embed={embed} />;
}

function Frame({ embed }: { embed: EmbedView }) {
  const where = origin(embed.url);
  const allowed = allowedOrigins().includes(where);
  return (
    <div className="embed-screen">
      <div className="tag">
        <span className="mono">/embed/{embed.id}</span>
        <a href={embed.url} target="_blank" rel="noopener noreferrer">
          open externally
        </a>
        <span className="muted wide">{where}</span>
      </div>
      {allowed ? (
        // The key remounts the frame when the configured address or
        // permissions change, so an old app never stays under new settings.
        <iframe
          key={[embed.id, embed.url, embed.sandbox, embed.allow].join('\n')}
          className="embed-frame"
          title={`${embed.label} (embedded app)`}
          src={embed.url}
          sandbox={embed.sandbox}
          allow={embed.allow || undefined}
          referrerPolicy="no-referrer"
        />
      ) : (
        <div className="section-body">
          <p>Desk loaded before this app was configured, so the browser would block its frame.</p>
          <p>
            <button onClick={() => location.reload()}>reload desk</button> to load it, or use open externally above.
          </p>
        </div>
      )}
      <p className="embed-note muted">
        A blank frame can mean the app refuses to be framed or needs its own sign-in. Desk cannot tell, so use open
        externally.
      </p>
    </div>
  );
}
