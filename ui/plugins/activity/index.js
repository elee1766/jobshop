// Entry of the built-in activity plugin (#541).
import { defineElement } from '../../kit/plugin.svelte.js';
import Activity from './Activity.svelte';

defineElement('desk-activity-view', Activity);
