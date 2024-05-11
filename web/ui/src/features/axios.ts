import axios from 'axios';

const url = document.location.href;
const contextRoot = url.match(/^(?:https?:\/\/)?[^/]+(\/[^/]+)?/i)?.[1] || "";
const localDev = process.env.NODE_ENV === 'development' && url.includes('dev=standalone');

export const ps_client = axios.create({
  baseURL: localDev ? 'http://localhost:3010' : `${contextRoot}/api`,
  headers: {
    'Content-Type': 'application/json',
  },
});
