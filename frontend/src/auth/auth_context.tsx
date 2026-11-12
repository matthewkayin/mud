import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import * as authApi from './auth_api.ts';
import type { User } from './auth_api.ts';

interface AuthContextValue {
  user: User | null;
  // True until the initial /api/auth/me check finishes.
  loading: boolean;
  // True when the server allows debug sign-in (a username with no password or session).
  debugAuth: boolean;
  signIn: (username: string, password: string) => Promise<void>;
  createAccount: (username: string, password: string) => Promise<void>;
  debugSignIn: (username: string) => Promise<void>;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [debugAuth, setDebugAuth] = useState(false);

  useEffect(() => {
    async function loadUser(): Promise<User | null> {
      const config = await authApi.getAuthConfig();
      setDebugAuth(config.debugAuth);

      const debugUsername = authApi.getDebugUsername();
      if (config.debugAuth && debugUsername !== null) {
        return { username: debugUsername };
      }
      return await authApi.getCurrentUser();
    }

    loadUser()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  const signIn = useCallback(async (username: string, password: string) => {
    setUser(await authApi.signIn(username, password));
  }, []);

  const createAccount = useCallback(async (username: string, password: string) => {
    setUser(await authApi.createAccount(username, password));
  }, []);

  const debugSignIn = useCallback(async (username: string) => {
    const debugUser = await authApi.debugSignIn(username);
    authApi.setDebugUsername(debugUser.username);
    setUser(debugUser);
  }, []);

  const signOut = useCallback(async () => {
    // A debug sign-in has no session. Calling logout would clear the cookie other tabs share.
    if (authApi.getDebugUsername() !== null) {
      authApi.clearDebugUsername();
    } else {
      await authApi.signOut();
    }
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({ user, loading, debugAuth, signIn, createAccount, debugSignIn, signOut }),
    [user, loading, debugAuth, signIn, createAccount, debugSignIn, signOut],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (context === null) {
    throw new Error('useAuth must be used inside an AuthProvider');
  }
  return context;
}
