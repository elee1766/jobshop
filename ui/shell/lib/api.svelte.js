import { createClient } from '../../../packages/jobshop-client/index.js';

// The server's base tag keeps every endpoint under a reverse-proxy prefix.
export const client = createClient({ baseURL: new URL('.', document.baseURI).href });
export const get = client.get;
export const post = client.post;
export const api = client.url;
export const base = new URL('.', document.baseURI).pathname;
