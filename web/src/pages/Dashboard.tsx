import { useAuth } from '../hooks/use-auth';
import { useNavigate } from 'react-router-dom';

export default function Dashboard() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-xl font-bold text-gray-900">TempleATS</h1>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">
              {user?.name} ({user?.orgName})
            </span>
            <button
              onClick={handleLogout}
              className="text-sm text-gray-500 hover:text-gray-700"
            >
              Logout
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8">
        <h2 className="text-2xl font-semibold text-gray-900 mb-6">Dashboard</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <p className="text-sm text-gray-500">Open Jobs</p>
            <p className="text-3xl font-bold text-gray-900 mt-1">0</p>
          </div>
          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <p className="text-sm text-gray-500">Active Requisitions</p>
            <p className="text-3xl font-bold text-gray-900 mt-1">0</p>
          </div>
          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <p className="text-sm text-gray-500">Total Candidates</p>
            <p className="text-3xl font-bold text-gray-900 mt-1">0</p>
          </div>
        </div>
      </main>
    </div>
  );
}
