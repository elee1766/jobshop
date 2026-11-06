/** Desk's plugin api, set as `element.api` on every plugin element. */
export interface DeskApi {
  /** Plugin API version; plugin.json "api": ">=N" needs at least N. */
  readonly version: number;
  /** Desk's base path, "" or "/prefix". */
  readonly base: string;
  /** The proxy user Desk acts for. */
  readonly user: string;
  /** GET a Desk path (relative, e.g. "api/work"). Rejects on non-2xx. */
  get<T = unknown>(path: string): Promise<T>;
  /** POST JSON to a Desk path; the CSRF header is added. */
  post<T = unknown>(path: string, body?: unknown): Promise<T>;
  /** Admin RPC through Desk's desk.rpc policy (GET api/rpc lists what is allowed). */
  call<M extends keyof RpcMethods>(method: M, params?: RpcMethods[M]["params"]): Promise<RpcMethods[M]["result"]>;
  /** Run a desk_commands entry. Resolves to its stdout: parsed JSON, else text. */
  exec(name: string, input?: unknown, opts?: { ticket?: string }): Promise<unknown>;
  /** Every event on Desk's stream. Returns an unsubscribe function. */
  subscribe(fn: (ev: Event) => void): () => void;
  /** Events whose fields equal every filter value. Returns an unsubscribe function. */
  events(filter: Partial<Record<"kind" | "op" | "id" | "ticket_kind" | "label" | "pool", string>>, fn: (ev: Event) => void): () => void;
  /** Go to a view, with optional URL params (ticket, session, artifact). */
  navigate(view: string, params?: Record<string, string>): void;
  toast(msg: string): void;
  confirm(msg: string): Promise<boolean>;
  /** URL of a path under Desk's base. */
  url(path: string): string;
  /** Seq of the last event Desk received. */
  readonly lastSeq: number;
  /** Desk's URL state: view, ticket, session, item and any other key (api 3 keeps every key). */
  params(): Record<string, string>;
  /** Change URL state and push a history entry (api 3, #541). */
  go(patch: Record<string, string>): void;
  /** Called with params() after every URL change (api 3). Returns an unsubscribe function. */
  onParams(fn: (params: Record<string, string>) => void): () => void;
  /** Reactive admin call (api 3): refetches when a result id or opts.watch matches an event. */
  query<M extends keyof RpcMethods>(method: M, params?: RpcMethods[M]["params"], opts?: { watch?: Record<string, string> }): DeskQuery<RpcMethods[M]["result"]>;
  /** Whether an enabled plugin contributes to slot for target (api 3). */
  has(slot: string, target?: Record<string, unknown>): boolean;
  /** Mount the first contribution to slot matching target in host (api 3). Null when none matches. */
  mount(host: HTMLElement, slot: string, target: Record<string, unknown>, context?: unknown): { context(c: unknown): void; destroy(): void } | null;
  /**
   * Mount every contribution to a plugin-defined slot ("work:toolbar") in host, each in its own box (api 4).
   * A failing one shows an error card in its box only. Null when none matches target.
   */
  slot(host: HTMLElement, slot: `${string}:${string}`, context?: unknown, target?: Record<string, unknown>): { context(c: unknown): void; destroy(): void } | null;
}

/** State of an api.query. */
export interface DeskQuery<T> {
  readonly data: T | null;
  readonly error: string;
  readonly loading: boolean;
  refresh(): void;
  subscribe(fn: (s: { data: T | null; error: string; loading: boolean }) => void): () => void;
}

/** What Desk sets on a plugin element: api, config, and context per slot. */
export interface PluginElement<C = unknown> extends HTMLElement {
  api: DeskApi;
  /** desk.plugins.<name>.config from jobshop.yaml. */
  config: Record<string, unknown>;
  context: C;
}

/** context for drawer and action slots. */
export interface DrawerContext {
  ticket: Ticket;
  run?: Ticket;
}

/** context for the artifact slot. */
export interface ArtifactContext {
  artifact: ArtifactShowResult;
  file: ArtifactFile & { url: string };
}

/** One plugin.json contributes entry. A slot named "<plugin>:<name>" is one a plugin hosts with api.slot. */
export interface Contribution {
  slot: "view" | "drawer" | "action" | "artifact" | "header" | "session" | `${string}:${string}`;
  element: string;
  id?: string;
  title?: string;
  match?: Record<string, string[]>;
}

/** plugin.json. */
export interface PluginManifest {
  name: string;
  version?: string;
  entry: string;
  api?: string;
  contributes: Contribution[];
}
