import { useState, useEffect, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, type TeamMember } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

export default function ReqForm() {
  const navigate = useNavigate();
  const [title, setTitle] = useState('');
  const [jobCode, setJobCode] = useState('');
  const [level, setLevel] = useState('');
  const [department, setDepartment] = useState('');
  const [hiringManagerId, setHiringManagerId] = useState('');
  const [recruiterId, setRecruiterId] = useState('');
  const [managers, setManagers] = useState<TeamMember[]>([]);
  const [recruiterOptions, setRecruiterOptions] = useState<TeamMember[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    api.team.list().then(data => {
      const eligible = data.members.filter(m =>
        m.role === 'hiring_manager' || m.role === 'admin' || m.role === 'super_admin'
      );
      setManagers(eligible);
      const eligibleRecruiters = data.members.filter(m =>
        m.role === 'recruiter' || m.role === 'admin' || m.role === 'super_admin'
      );
      setRecruiterOptions(eligibleRecruiters);
    });
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const req = await api.reqs.create({
        title,
        jobCode: jobCode || undefined,
        level: level || undefined,
        department: department || undefined,
        targetHires: 1,
        hiringManagerId: hiringManagerId || undefined,
        recruiterId: recruiterId || undefined,
      });
      navigate(`/reqs/${req.id}`);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="max-w-2xl">
        <h2 className="text-2xl font-semibold text-gray-900 mb-6">Create Requisition</h2>

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
              placeholder="e.g., Senior Backend Engineer"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Job Code *</label>
            <input
              type="text"
              value={jobCode}
              onChange={e => setJobCode(e.target.value)}
              required
              placeholder="e.g., SWE, MLE, SYS"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Level</label>
              <input
                type="text"
                value={level}
                onChange={e => setLevel(e.target.value)}
                placeholder="e.g., L5, IC4, Senior"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
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
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Hiring Manager</label>
            <select
              value={hiringManagerId}
              onChange={e => setHiringManagerId(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">Assign to me</option>
              {managers.map(m => (
                <option key={m.id} value={m.id}>{m.name} ({m.role.replace('_', ' ')})</option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Recruiter</label>
            <select
              value={recruiterId}
              onChange={e => setRecruiterId(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">None</option>
              {recruiterOptions.map(m => (
                <option key={m.id} value={m.id}>{m.name} ({m.role.replace('_', ' ')})</option>
              ))}
            </select>
          </div>

          <div className="flex gap-3 pt-2">
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
            >
              {loading ? 'Creating...' : 'Create Requisition'}
            </button>
            <button
              type="button"
              onClick={() => navigate('/reqs')}
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
