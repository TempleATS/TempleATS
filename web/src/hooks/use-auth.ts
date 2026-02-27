import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { api, type User } from '../api/client';

const ROLE_HIERARCHY = ['interviewer', 'hiring_manager', 'recruiter', 'admin', 'super_admin'] as const;
export type Role = typeof ROLE_HIERARCHY[number];

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  signup: (data: { email: string; name: string; password: string; orgName: string; orgSlug: string }) => Promise<void>;
  logout: () => Promise<void>;
  setUser: (user: User | null) => void;
  isAtLeast: (role: Role) => boolean;
  hasRole: (...roles: Role[]) => boolean;
}

export const AuthContext = createContext<AuthContextType>({
  user: null,
  loading: true,
  login: async () => {},
  signup: async () => {},
  logout: async () => {},
  setUser: () => {},
  isAtLeast: () => false,
  hasRole: () => false,
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

  const isAtLeast = useCallback((role: Role): boolean => {
    if (!user) return false;
    const userLevel = ROLE_HIERARCHY.indexOf(user.role as Role);
    const requiredLevel = ROLE_HIERARCHY.indexOf(role);
    return userLevel >= requiredLevel;
  }, [user]);

  const hasRole = useCallback((...roles: Role[]): boolean => {
    if (!user) return false;
    return roles.includes(user.role as Role);
  }, [user]);

  return { user, loading, login, signup, logout, setUser, isAtLeast, hasRole };
}
