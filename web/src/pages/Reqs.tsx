import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api, type Requisition } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

export default function Reqs() {
  const [reqs, setReqs] = useState<Requisition[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.reqs.list().then(setReqs).finally(() => setLoading(false));
  }, []);

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
        <Link
          to="/reqs/new"
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm font-medium"
        >
          Create Requisition
        </Link>
      </div>

      {loading ? (
        <p className="text-gray-500">Loading...</p>
      ) : reqs.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border">
          <p className="text-gray-500">No requisitions yet.</p>
          <Link to="/reqs/new" className="text-blue-600 hover:underline text-sm mt-2 inline-block">
            Create your first requisition
          </Link>
        </div>
      ) : (
        <div className="bg-white rounded-lg border overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50 border-b">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Title</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Level</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Department</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Target</th>
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
                  <td className="px-4 py-3 text-sm text-gray-600">{req.level || '-'}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">{req.department || '-'}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">{req.target_hires}</td>
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
