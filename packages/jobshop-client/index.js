// No UI, framework, DOM, storage or global connection state lives here.
export class ApiError extends Error {
  constructor(message, status, code = "internal") {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

export function createClient({
  baseURL,
  fetch: request = globalThis.fetch,
  EventSource: Source = globalThis.EventSource,
} = {}) {
  if (!baseURL) throw new TypeError("baseURL is required");
  const base = new URL(baseURL);
  if (!["http:", "https:"].includes(base.protocol))
    throw new TypeError("baseURL must use HTTP or HTTPS");
  if (!base.pathname.endsWith("/")) base.pathname += "/";
  base.search = "";
  base.hash = "";
  const url = (path) => {
    const target = new URL(path.replace(/^\/+/, ""), base);
    if (
      target.origin !== base.origin ||
      !target.pathname.startsWith(base.pathname)
    )
      throw new TypeError("path must stay under baseURL");
    return target.href;
  };
  async function json(path, method, body, { signal } = {}) {
    const response = await request(url(path), {
      method,
      signal,
      credentials: "same-origin",
      headers: {
        Accept: "application/json",
        ...(method === "POST"
          ? { "Content-Type": "application/json", "X-Jobshop-Desk": "1" }
          : {}),
      },
      ...(method === "POST" ? { body: JSON.stringify(body ?? {}) } : {}),
    });
    let data;
    try {
      data = await response.json();
    } catch {
      throw new ApiError(
        "Invalid JSON response",
        response.status,
        "invalid_response",
      );
    }
    if (!response.ok)
      throw new ApiError(
        data?.error || `${response.status} ${response.statusText}`,
        response.status,
        data?.code,
      );
    return data;
  }
  function stream(path, kinds, onEvent, { signal, onState = () => {} } = {}) {
    if (!Source)
      throw new TypeError(
        "EventSource is unavailable, supply an implementation",
      );
    if (signal?.aborted) return () => {};
    const source = new Source(url(path), { withCredentials: true });
    let closed = false;
    const close = () => {
      if (closed) return;
      closed = true;
      source.close();
      signal?.removeEventListener("abort", close);
      onState("closed");
    };
    source.addEventListener("open", () => {
      if (!closed) onState("connected");
    });
    source.addEventListener("error", (event) => {
      if (closed) return;
      if (event.data) {
        try {
          onState("error", JSON.parse(event.data));
        } catch {
          onState("error");
        }
      } else onState("reconnecting");
    });
    for (const kind of kinds)
      source.addEventListener(kind, (event) => {
        if (closed) return;
        let data;
        try {
          data = JSON.parse(event.data);
        } catch {
          onState("error", { error: "Invalid event JSON" });
          return;
        }
        onEvent(data, kind);
        if (kind === "end") close();
      });
    signal?.addEventListener("abort", close, { once: true });
    return close;
  }
  return Object.freeze({
    url,
    get: (path, opts) => json(path, "GET", undefined, opts),
    post: (path, body, opts) => json(path, "POST", body, opts),
    call: (method, params, opts) =>
      json(`api/rpc/${encodeURIComponent(method)}`, "POST", params, opts),
    events: (fn, opts = {}) =>
      stream(
        "api/events" +
          (opts.since
            ? "?" + new URLSearchParams({ since: String(opts.since) })
            : ""),
        [
          "ticket",
          "run",
          "session",
          "pool",
          "activity",
          "ready",
          "lagged",
          "desk",
          "system",
          "state",
        ],
        fn,
        opts,
      ),
    live: (fn, opts = {}) =>
      stream(
        "api/live" +
          (opts.trace || opts.audit
            ? "?" +
              new URLSearchParams({
                ...(opts.trace ? { trace: opts.trace } : {}),
                ...(opts.inspect ? { inspect: opts.inspect } : {}),
                ...(opts.audit ? { audit: opts.audit } : {}),
              })
            : ""),
        [
          "attention",
          "overview",
          "system",
          "chats",
          "improvements",
          "embeds",
          "desk",
          "scheduler",
          "trace",
          "automation",
        ],
        (data, kind) => fn({ kind, data }),
        opts,
      ),
    session: (id, fn, opts = {}) => {
      const query = new URLSearchParams({ tail: String(opts.tail ?? 200) });
      if (opts.since) query.set("since", String(opts.since));
      return stream(
        `api/session/${encodeURIComponent(id)}/stream?${query}`,
        ["session", "end"],
        fn,
        opts,
      );
    },
  });
}
