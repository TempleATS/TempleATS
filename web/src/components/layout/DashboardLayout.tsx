import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/use-auth';
import Sidebar from './Sidebar';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex-1 flex flex-col">
        <header className="bg-white border-b px-6 py-3 flex items-center justify-between">
          <div />
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">{user?.name}</span>
            <button onClick={handleLogout} className="text-sm text-gray-500 hover:text-gray-700">
              Logout
            </button>
          </div>
        </header>
        <main className="flex-1 p-6 bg-gray-50">
          {children}
        </main>
      </div>
    </div>
  );
}
