// Entry of the built-in desk plugin (JOB-85): the default `app`, the
// frame around every view.
import { defineElement } from '../../kit/plugin.svelte.js';
import Frame from './Frame.svelte';

defineElement('desk-app', Frame);
