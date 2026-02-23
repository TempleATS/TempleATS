import { useState, useEffect, type FormEvent } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { api, type Requisition } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

export default function JobForm() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const preselectedReqId = searchParams.get('reqId');

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [location, setLocation] = useState('');
  const [department, setDepartment] = useState('');
  const [salary, setSalary] = useState('');
  const [requisitionId, setRequisitionId] = useState(preselectedReqId || '');
  const [reqs, setReqs] = useState<Requisition[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    api.reqs.list().then(r => setReqs(r.filter(req => req.status === 'open')));
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const job = await api.jobs.create({
        title,
        description,
        location: location || undefined,
        department: department || undefined,
        salary: salary || undefined,
        requisitionId: requisitionId || undefined,
      });
      navigate(`/jobs/${job.id}`);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="max-w-2xl">
        <h2 className="text-2xl font-semibold text-gray-900 mb-6">Create Job</h2>

        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{error}</div>
        )}

        <form onSubmit={handleSubmit} className="bg-white p-6 rounded-lg border space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Title *</label>
            <input
              type="text"
              value={title}
              onChange={e => setTitle(e.target.value)}
              required
              placeholder="e.g., Senior Go Developer"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Description *</label>
            <textarea
              value={description}
              onChange={e => setDescription(e.target.value)}
              required
              rows={8}
              placeholder="Job description (supports markdown)"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Location</label>
              <input
                type="text"
                value={location}
                onChange={e => setLocation(e.target.value)}
                placeholder="e.g., Remote, NYC"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Salary</label>
              <input
                type="text"
                value={salary}
                onChange={e => setSalary(e.target.value)}
                placeholder="e.g., $120k-150k"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Department</label>
              <input
                type="text"
                value={department}
                onChange={e => setDepartment(e.target.value)}
                placeholder="e.g., Engineering"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Requisition</label>
              <select
                value={requisitionId}
                onChange={e => setRequisitionId(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">None</option>
                {reqs.map(req => (
                  <option key={req.id} value={req.id}>
                    {req.title} {req.level ? `(${req.level})` : ''}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="flex gap-3 pt-2">
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
            >
              {loading ? 'Creating...' : 'Create Job (as Draft)'}
            </button>
            <button
              type="button"
              onClick={() => navigate('/jobs')}
              className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 text-sm"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </DashboardLayout>
  );
}
