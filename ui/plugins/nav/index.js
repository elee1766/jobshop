// Entry of the built-in nav plugin (JOB-86): the links, user, live dot,
// theme toggle, header widgets and g+letter keys the desk app used to draw.
import { defineElement } from '../../kit/plugin.svelte.js';
import Nav from './Nav.svelte';

defineElement('desk-nav', Nav);
