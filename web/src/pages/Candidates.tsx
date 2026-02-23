import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api, type Candidate } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

function pgText(val: { String: string; Valid: boolean } | null): string | null {
  return val?.Valid ? val.String : null;
}

export default function Candidates() {
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');

  useEffect(() => {
    setLoading(true);
    const timeout = setTimeout(() => {
      api.candidates.list(search || undefined)
        .then(setCandidates)
        .finally(() => setLoading(false));
    }, search ? 300 : 0);
    return () => clearTimeout(timeout);
  }, [search]);

  return (
    <DashboardLayout>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-semibold text-gray-900">Candidates</h2>
      </div>

      <div className="mb-4">
        <input
          type="text"
          value={search}
          onChange={e => setSearch(e.target.value)}
          placeholder="Search by name or email..."
          className="w-full max-w-md px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      {loading ? (
        <p className="text-gray-500">Loading...</p>
      ) : candidates.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border">
          <p className="text-gray-500">
            {search ? 'No candidates match your search.' : 'No candidates yet.'}
          </p>
        </div>
      ) : (
        <div className="bg-white rounded-lg border overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50 border-b">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Name</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Email</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Phone</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Added</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {candidates.map(c => (
                <tr key={c.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <Link to={`/candidates/${c.id}`} className="text-blue-600 hover:underline font-medium">
                      {c.name}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">{c.email}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">{pgText(c.phone) || '-'}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">
                    {c.created_at?.Time ? new Date(c.created_at.Time).toLocaleDateString() : '-'}
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
