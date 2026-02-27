import { Link, useLocation } from 'react-router-dom';
import { useAuth, type Role } from '../../hooks/use-auth';

interface NavItem {
  path: string;
  label: string;
  minRole?: Role;
}

const navItems: NavItem[] = [
  { path: '/dashboard', label: 'Dashboard', minRole: 'admin' },
  { path: '/reqs', label: 'Requisitions', minRole: 'recruiter' },
  { path: '/jobs', label: 'Jobs', minRole: 'hiring_manager' },
  { path: '/candidates', label: 'Candidates', minRole: 'hiring_manager' },
  { path: '/interviews', label: 'Interviews' },
  { path: '/referrals', label: 'Referrals' },
  { path: '/team', label: 'Team', minRole: 'recruiter' },
  { path: '/settings', label: 'Settings', minRole: 'super_admin' },
];

const ROLE_LABELS: Record<string, string> = {
  super_admin: 'Super Admin',
  admin: 'Admin',
  recruiter: 'Recruiter',
  hiring_manager: 'Hiring Manager',
  interviewer: 'Interviewer',
};

export default function Sidebar() {
  const location = useLocation();
  const { user, isAtLeast } = useAuth();

  return (
    <aside className="w-64 bg-gray-900 text-white min-h-screen flex flex-col">
      <div className="p-4 border-b border-gray-700">
        <h1 className="text-lg font-bold">TempleATS</h1>
        <p className="text-xs text-gray-400 mt-1">{user?.orgName}</p>
      </div>

      <nav className="flex-1 p-4 space-y-1">
        {navItems
          .filter(item => !item.minRole || isAtLeast(item.minRole))
          .map(item => {
            const active = location.pathname === item.path || location.pathname.startsWith(item.path + '/');
            return (
              <Link
                key={item.path}
                to={item.path}
                className={`block px-3 py-2 rounded text-sm ${
                  active
                    ? 'bg-gray-700 text-white font-medium'
                    : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                }`}
              >
                {item.label}
              </Link>
            );
          })}
      </nav>

      <div className="p-4 border-t border-gray-700">
        <Link
          to="/account"
          className={`block px-3 py-2 rounded text-sm mb-2 ${
            location.pathname === '/account'
              ? 'bg-gray-700 text-white font-medium'
              : 'text-gray-300 hover:bg-gray-800 hover:text-white'
          }`}
        >
          My Account
        </Link>
        <p className="text-sm text-gray-400 truncate">{user?.email}</p>
        <p className="text-xs text-gray-500 mt-0.5">{ROLE_LABELS[user?.role || ''] || user?.role}</p>
      </div>
    </aside>
  );
}
