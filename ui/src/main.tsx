import { createRoot } from 'react-dom/client';
import './index.css';
import { App } from './App.tsx';

const target = document.getElementById('app');
if (!target) throw new Error('missing #app');
createRoot(target).render(<App />);
