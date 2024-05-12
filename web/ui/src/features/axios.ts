import axios from 'axios';

const url = document.location.href;
const ctxRootExtractor = /^(?:https?:\/\/)?[^/]+(\/[^/]+)?/i;
const contextRoot = ctxRootExtractor.exec(url)?.[1] ?? '';
const localDev = process.env.NODE_ENV === 'development' && url.includes('dev=standalone');

export const ps_client = axios.create({
  baseURL: localDev ? 'http://localhost:3010' : `${contextRoot}/api`,
  headers: {
    'Content-Type': 'application/json',
    'X-Tenant': 'almawave.com',
    'X-Subscription': 'labrd'
  },
});
