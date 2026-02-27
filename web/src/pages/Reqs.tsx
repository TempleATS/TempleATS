import { useState, useEffect, useRef, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { api, type Requisition, type TeamMember } from '../api/client';
import { useAuth } from '../hooks/use-auth';
import DashboardLayout from '../components/layout/DashboardLayout';

export default function Reqs() {
  const { isAtLeast } = useAuth();
  const [reqs, setReqs] = useState<Requisition[]>([]);
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const canCreate = isAtLeast('admin');
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const membersLoaded = useRef(false);

  const fetchReqs = useCallback((q?: string) => {
    setLoading(true);
    const reqPromise = api.reqs.list(q || undefined);
    if (!membersLoaded.current) {
      Promise.all([reqPromise, api.team.list()]).then(([reqData, teamData]) => {
        setReqs(reqData);
        setMembers(teamData.members);
        membersLoaded.current = true;
      }).finally(() => setLoading(false));
    } else {
      reqPromise.then(setReqs).finally(() => setLoading(false));
    }
  }, []);

  useEffect(() => {
    fetchReqs();
  }, [fetchReqs]);

  const handleSearch = (value: string) => {
    setSearch(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      fetchReqs(value);
    }, 300);
  };

  const hmName = (id: string | null) => {
    if (!id) return '-';
    return members.find(m => m.id === id)?.name || '-';
  };

  const statusColor = (status: string) => {
    switch (status) {
      case 'open': return 'bg-green-100 text-green-800';
      case 'filled': return 'bg-blue-100 text-blue-800';
      case 'cancelled': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <DashboardLayout>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-semibold text-gray-900">Requisitions</h2>
        {canCreate && (
          <Link
            to="/reqs/new"
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm font-medium"
          >
            Create Requisition
          </Link>
        )}
      </div>

      <div className="mb-4">
        <input
          type="text"
          placeholder="Search requisitions by title, job code, department..."
          value={search}
          onChange={e => handleSearch(e.target.value)}
          className="w-full max-w-md px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
      </div>

      {loading ? (
        <p className="text-gray-500">Loading...</p>
      ) : reqs.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border">
          <p className="text-gray-500">{search ? 'No requisitions match your search.' : 'No requisitions yet.'}</p>
          {!search && (
            <Link to="/reqs/new" className="text-blue-600 hover:underline text-sm mt-2 inline-block">
              Create your first requisition
            </Link>
          )}
        </div>
      ) : (
        <div className="bg-white rounded-lg border overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50 border-b">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Title</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Job Code</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Department</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Hiring Manager</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Opened</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {reqs.map(req => (
                <tr key={req.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <Link to={`/reqs/${req.id}`} className="text-blue-600 hover:underline font-medium">
                      {req.title}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">{req.job_code || '-'}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">{req.department || '-'}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">{hmName(req.hiring_manager_id)}</td>
                  <td className="px-4 py-3">
                    <span className={`text-xs px-2 py-1 rounded-full font-medium ${statusColor(req.status)}`}>
                      {req.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">
                    {new Date(req.opened_at).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </DashboardLayout>
  );
}
