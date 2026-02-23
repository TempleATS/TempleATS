import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/use-auth';

const navItems = [
  { path: '/dashboard', label: 'Dashboard' },
  { path: '/reqs', label: 'Requisitions' },
  { path: '/jobs', label: 'Jobs' },
  { path: '/candidates', label: 'Candidates' },
  { path: '/settings', label: 'Settings' },
];

export default function Sidebar() {
  const location = useLocation();
  const { user } = useAuth();

  return (
    <aside className="w-64 bg-gray-900 text-white min-h-screen flex flex-col">
      <div className="p-4 border-b border-gray-700">
        <h1 className="text-lg font-bold">TempleATS</h1>
        <p className="text-xs text-gray-400 mt-1">{user?.orgName}</p>
      </div>

      <nav className="flex-1 p-4 space-y-1">
        {navItems.map(item => {
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
        <p className="text-sm text-gray-400 truncate">{user?.email}</p>
      </div>
    </aside>
  );
}
