// Entry of the built-in sessions plugin (#541).
import { defineElement } from '../../kit/plugin.svelte.js';
import Sessions from './Sessions.svelte';

defineElement('desk-sessions-view', Sessions);
