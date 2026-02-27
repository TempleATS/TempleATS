import { type ReactNode } from 'react';
import { render, type RenderOptions } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { AuthContext, type Role } from '../hooks/use-auth';
import type { User } from '../api/client';

interface TestUser extends User {
  role: Role;
}

const defaultUser: TestUser = {
  id: 'user-1',
  email: 'test@example.com',
  name: 'Test User',
  role: 'recruiter',
  orgId: 'org-1',
  orgSlug: 'test-org',
  orgName: 'Test Org',
};

interface WrapperOptions {
  user?: TestUser | null;
  route?: string;
}

function createWrapper({ user = defaultUser, route = '/' }: WrapperOptions = {}) {
  const ROLE_HIERARCHY: Role[] = ['interviewer', 'hiring_manager', 'recruiter', 'admin', 'super_admin'];

  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <MemoryRouter initialEntries={[route]}>
        <AuthContext.Provider
          value={{
            user,
            loading: false,
            login: async () => {},
            signup: async () => {},
            logout: async () => {},
            setUser: () => {},
            isAtLeast: (role: Role) => {
              if (!user) return false;
              return ROLE_HIERARCHY.indexOf(user.role) >= ROLE_HIERARCHY.indexOf(role);
            },
            hasRole: (...roles: Role[]) => {
              if (!user) return false;
              return roles.includes(user.role);
            },
          }}
        >
          {children}
        </AuthContext.Provider>
      </MemoryRouter>
    );
  };
}

export function renderWithProviders(
  ui: ReactNode,
  options?: WrapperOptions & Omit<RenderOptions, 'wrapper'>,
) {
  const { user, route, ...renderOptions } = options ?? {};
  return render(ui, {
    wrapper: createWrapper({ user, route }),
    ...renderOptions,
  });
}

export { defaultUser };
