// The chat panel registers the few imperative entry points other screens
// use (open a chat, start a draft, attach references). Registered on mount,
// removed on unmount, so there is exactly one panel to call.
import type { Chat } from './ui.ts';

export const chatBridge: {
  open?: (chat: Chat) => void;
  start?: (draft: string, refs: string[], name: string) => void;
  attach?: (refs: string[]) => void;
} = {};
