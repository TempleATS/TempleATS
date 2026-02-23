import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { api, type User } from '../api/client';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  signup: (data: { email: string; name: string; password: string; orgName: string; orgSlug: string }) => Promise<void>;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType>({
  user: null,
  loading: true,
  login: async () => {},
  signup: async () => {},
  logout: async () => {},
});

export function useAuth() {
  return useContext(AuthContext);
}

export function useAuthProvider(): AuthContextType {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.auth.me()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const u = await api.auth.login({ email, password });
    setUser(u);
  }, []);

  const signup = useCallback(async (data: { email: string; name: string; password: string; orgName: string; orgSlug: string }) => {
    const u = await api.auth.signup(data);
    setUser(u);
  }, []);

  const logout = useCallback(async () => {
    await api.auth.logout();
    setUser(null);
  }, []);

  return { user, loading, login, signup, logout };
}
