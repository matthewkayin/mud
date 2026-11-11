// Wrappers around the server's /api/auth routes. The contract is described in docs/frontend.md.

export interface User {
  username: string;
}

async function request(method: string, path: string, body?: unknown): Promise<Response> {
  let response: Response;
  try {
    response = await fetch(path, {
      method,
      credentials: 'same-origin',
      headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new Error('Could not reach the server.');
  }

  if (!response.ok) {
    throw new Error(await errorMessageFromResponse(response));
  }
  return response;
}

async function errorMessageFromResponse(response: Response): Promise<string> {
  try {
    const json = await response.json();
    if (typeof json?.error === 'string') {
      return json.error;
    }
  } catch {
    // The body was not JSON (e.g. the dev proxy could not reach the backend).
  }
  if (response.status >= 500) {
    return 'Could not reach the server.';
  }
  return `The server refused the request (${response.status}).`;
}

export async function getCurrentUser(): Promise<User | null> {
  const response = await fetch('/api/auth/me', { credentials: 'same-origin' });
  if (!response.ok) {
    return null;
  }
  return await response.json() as User;
}

export async function signIn(username: string, password: string): Promise<User> {
  const response = await request('POST', '/api/auth/login', { username, password });
  return await response.json() as User;
}

export async function createAccount(username: string, password: string): Promise<User> {
  const response = await request('POST', '/api/auth/register', { username, password });
  return await response.json() as User;
}

export async function signOut(): Promise<void> {
  await request('POST', '/api/auth/logout');
}

export interface AuthConfig {
  debugAuth: boolean;
}

export async function getAuthConfig(): Promise<AuthConfig> {
  try {
    const response = await fetch('/api/auth/config', { credentials: 'same-origin' });
    if (!response.ok) {
      return { debugAuth: false };
    }
    const json = await response.json();
    return { debugAuth: json?.debugAuth === true };
  } catch {
    return { debugAuth: false };
  }
}

// Debug sign-in only validates the username. The server keeps no session for it,
// so the username is stored per tab and sent when the websocket opens.
export async function debugSignIn(username: string): Promise<User> {
  const response = await request('POST', '/api/auth/debug-login', { username });
  return await response.json() as User;
}

const DEBUG_USERNAME_STORAGE_KEY = 'mud_debug_username';

export function getDebugUsername(): string | null {
  try {
    return sessionStorage.getItem(DEBUG_USERNAME_STORAGE_KEY);
  } catch {
    return null;
  }
}

export function setDebugUsername(username: string): void {
  try {
    sessionStorage.setItem(DEBUG_USERNAME_STORAGE_KEY, username);
  } catch {
    // Without sessionStorage the debug sign-in only lasts until the page reloads.
  }
}

export function clearDebugUsername(): void {
  try {
    sessionStorage.removeItem(DEBUG_USERNAME_STORAGE_KEY);
  } catch {
    // Nothing to clear.
  }
}
