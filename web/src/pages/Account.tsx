import { useState, useEffect } from 'react';
import { useAuth } from '../hooks/use-auth';
import { api, type CalendarConnection } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

export default function Account() {
  const { user } = useAuth();
  const [calendar, setCalendar] = useState<CalendarConnection | null>(null);
  const [loading, setLoading] = useState(true);
  const [connecting, setConnecting] = useState(false);
  const [disconnecting, setDisconnecting] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    api.account.getCalendar()
      .then(data => setCalendar(data))
      .catch(() => setCalendar(null))
      .finally(() => setLoading(false));

    // Check URL params for status
    const params = new URLSearchParams(window.location.search);
    if (params.get('calendar') === 'connected') {
      setSuccess('Google Calendar connected successfully!');
      window.history.replaceState({}, '', '/account');
    }
    if (params.get('error') === 'oauth_failed') {
      setError('Failed to connect Google Calendar. Please try again.');
      window.history.replaceState({}, '', '/account');
    }
  }, []);

  const handleConnect = async () => {
    setConnecting(true);
    setError('');
    try {
      const { url } = await api.account.getGoogleAuthUrl();
      window.location.href = url;
    } catch {
      setError('Google Calendar is not configured. Contact your administrator.');
      setConnecting(false);
    }
  };

  const handleDisconnect = async () => {
    setDisconnecting(true);
    setError('');
    try {
      await api.account.disconnectCalendar();
      setCalendar(null);
      setSuccess('Calendar disconnected.');
    } catch {
      setError('Failed to disconnect calendar.');
    } finally {
      setDisconnecting(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="max-w-2xl">
        <h1 className="text-2xl font-bold mb-6">My Account</h1>

        {error && (
          <div className="mb-4 p-3 bg-red-50 text-red-700 rounded text-sm">{error}</div>
        )}
        {success && (
          <div className="mb-4 p-3 bg-green-50 text-green-700 rounded text-sm">{success}</div>
        )}

        {/* Profile Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold mb-4">Profile</h2>
          <div className="space-y-3">
            <div>
              <label className="text-sm text-gray-500">Name</label>
              <p className="font-medium">{user?.name}</p>
            </div>
            <div>
              <label className="text-sm text-gray-500">Email</label>
              <p className="font-medium">{user?.email}</p>
            </div>
            <div>
              <label className="text-sm text-gray-500">Role</label>
              <p className="font-medium capitalize">{user?.role?.replace('_', ' ')}</p>
            </div>
          </div>
        </div>

        {/* Calendar Connection Section */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Calendar Connection</h2>
          <p className="text-sm text-gray-600 mb-4">
            Connect your Google Calendar to enable interview scheduling. Your calendar will be used
            to check your availability and create interview events.
          </p>

          {loading ? (
            <p className="text-sm text-gray-500">Loading...</p>
          ) : calendar ? (
            <div className="flex items-center justify-between p-4 bg-green-50 rounded-lg">
              <div>
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 bg-green-500 rounded-full" />
                  <span className="font-medium text-green-800">Google Calendar Connected</span>
                </div>
                <p className="text-sm text-green-700 mt-1">{calendar.calendar_email}</p>
              </div>
              <button
                onClick={handleDisconnect}
                disabled={disconnecting}
                className="px-3 py-1.5 text-sm text-red-600 border border-red-300 rounded hover:bg-red-50 disabled:opacity-50"
              >
                {disconnecting ? 'Disconnecting...' : 'Disconnect'}
              </button>
            </div>
          ) : (
            <button
              onClick={handleConnect}
              disabled={connecting}
              className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-300 rounded-lg shadow-sm hover:bg-gray-50 disabled:opacity-50"
            >
              <svg viewBox="0 0 24 24" className="w-5 h-5">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 01-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" />
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
              </svg>
              {connecting ? 'Redirecting...' : 'Connect Google Calendar'}
            </button>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
}
