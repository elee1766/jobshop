// Icon names a view manifest may use (JOB-86): lucide icons by name. An
// unknown name shows the puzzle piece. Each import adds to this plugin's
// bundle, so the list is short; a view that needs another icon can ship
// it in a nav of its own.
import Inbox from '@lucide/svelte/icons/inbox';
import MessageSquare from '@lucide/svelte/icons/message-square';
import ListChecks from '@lucide/svelte/icons/list-checks';
import Activity from '@lucide/svelte/icons/activity';
import Terminal from '@lucide/svelte/icons/terminal';
import Boxes from '@lucide/svelte/icons/boxes';
import Rocket from '@lucide/svelte/icons/rocket';
import Puzzle from '@lucide/svelte/icons/puzzle';
import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
import Settings from '@lucide/svelte/icons/settings';
import Bell from '@lucide/svelte/icons/bell';
import FileText from '@lucide/svelte/icons/file-text';
import GitPullRequest from '@lucide/svelte/icons/git-pull-request';
import ChartBar from '@lucide/svelte/icons/chart-bar';
import Search from '@lucide/svelte/icons/search';
import Server from '@lucide/svelte/icons/server';
import Users from '@lucide/svelte/icons/users';
import Database from '@lucide/svelte/icons/database';
import Shield from '@lucide/svelte/icons/shield';
import Clock from '@lucide/svelte/icons/clock';
import Zap from '@lucide/svelte/icons/zap';
import House from '@lucide/svelte/icons/house';

/** @type {Record<string, any>} */
const icons = {
  inbox: Inbox,
  'message-square': MessageSquare,
  'list-checks': ListChecks,
  activity: Activity,
  terminal: Terminal,
  boxes: Boxes,
  rocket: Rocket,
  puzzle: Puzzle,
  'layout-dashboard': LayoutDashboard,
  settings: Settings,
  bell: Bell,
  'file-text': FileText,
  'git-pull-request': GitPullRequest,
  'chart-bar': ChartBar,
  search: Search,
  server: Server,
  users: Users,
  database: Database,
  shield: Shield,
  clock: Clock,
  zap: Zap,
  house: House,
};

/** @param {string | undefined} name */
export const iconFor = (name) => icons[name ?? ''] ?? Puzzle;
