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
