import type {
  RpcMethods,
  Event,
  SessionEvent,
  SessionSubscribeResult,
  AttentionResult,
  RunResult,
  SystemStatus,
  PoolStatusResult,
  ResourceListResult,
  SessionInfo,
  Learning,
  MemoryEdit,
  ShowResult,
  TraceNode,
  GraphResult,
  HolderInfo,
  Ticket,
  IngestWorkersResult,
} from "./rpc.js";
export type * from "./rpc.js";
export interface RequestOptions {
  signal?: AbortSignal;
}
export interface StreamOptions extends RequestOptions {
  onState?: (
    state: "connected" | "reconnecting" | "closed" | "error",
    detail?: { error?: string },
  ) => void;
}
export interface EventsOptions extends StreamOptions {
  since?: number;
}
export interface SessionOptions extends StreamOptions {
  tail?: number;
  since?: number;
}
export interface ChatSession {
  id: string;
  key: string;
  title: string;
  state: string;
  alive: boolean;
  updated: string;
}
export interface LiveOptions extends StreamOptions {
  trace?: string;
  inspect?: string;
  audit?: string;
}
export interface TraceView extends ShowResult {
  inspected?: Ticket;
  inspection?: ShowResult;
  graph?: GraphResult;
  run?: RunResult;
  trace?: TraceNode;
  sessions?: Record<string, HolderInfo>;
}
export interface AutomationSummary {
  key: string;
  state: string;
  revision: number;
  attempts: number;
  created_at: string;
  updated_at: string;
  trigger?: string;
  run?: string;
  step?: string;
  entry: string;
  event?: { source: string; event: string; delivery: string };
  error?: string;
}
export interface AutomationRecord {
  key: string;
  state: string;
  revision: number;
  attempts: number;
  created_at: string;
  updated_at: string;
  value: Record<string, any>;
}
export interface AutomationList {
  items: AutomationSummary[];
  truncated: boolean;
}
export interface AutomationAudit {
  id: string;
  invocation: AutomationRecord;
  attempts: AutomationRecord[];
  calls: AutomationRecord[];
  truncated: boolean;
}
export interface SystemView {
  status: SystemStatus;
  pools: PoolStatusResult;
  resources: ResourceListResult;
  sessions: SessionInfo[];
  ingest: IngestWorkersResult;
  automation: AutomationList;
}
export interface LiveState {
  attention: AttentionResult;
  overview: { runs: RunResult[]; truncated: boolean };
  system: SystemView;
  automation: AutomationAudit | { id: string; error: string };
  chats: { pool: string; sessions: ChatSession[] };
  improvements: { learnings: Learning[]; edits: MemoryEdit[] };
  trace: TraceView | { error: string };
  desk: Event;
  scheduler: Event;
}
export type LiveEvent = {
  [K in keyof LiveState]: { kind: K; data: LiveState[K] };
}[keyof LiveState];
export interface Client {
  url(path: string): string;
  get<T = any>(path: string, options?: RequestOptions): Promise<T>;
  post<T = any>(
    path: string,
    body?: unknown,
    options?: RequestOptions,
  ): Promise<T>;
  call<M extends keyof RpcMethods>(
    method: M,
    params: RpcMethods[M]["params"],
    options?: RequestOptions,
  ): Promise<RpcMethods[M]["result"]>;
  events(
    fn: (event: Event, kind: string) => void,
    options?: EventsOptions,
  ): () => void;
  live(fn: (event: LiveEvent) => void, options?: LiveOptions): () => void;
  session(
    id: string,
    fn: (event: SessionEvent | SessionSubscribeResult, kind: string) => void,
    options?: SessionOptions,
  ): () => void;
}
export class ApiError extends Error {
  status: number;
  code: string;
}
export function createClient(options: {
  baseURL: string;
  fetch?: typeof fetch;
  EventSource?: typeof EventSource;
}): Client;
