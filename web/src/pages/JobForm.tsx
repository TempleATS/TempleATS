import { useState, useEffect, type FormEvent } from 'react';
import { useNavigate, useSearchParams, useParams } from 'react-router-dom';
import { api, type Requisition } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

export default function JobForm() {
  const navigate = useNavigate();
  const { jobId } = useParams<{ jobId: string }>();
  const [searchParams] = useSearchParams();
  const preselectedReqId = searchParams.get('reqId');
  const isEdit = !!jobId;

  const [title, setTitle] = useState('');
  const [companyBlurb, setCompanyBlurb] = useState('');
  const [teamDetails, setTeamDetails] = useState('');
  const [responsibilities, setResponsibilities] = useState('');
  const [qualifications, setQualifications] = useState('');
  const [closingStatement, setClosingStatement] = useState('');
  const [location, setLocation] = useState('');
  const [department, setDepartment] = useState('');
  const [salary, setSalary] = useState('');
  const [status, setStatus] = useState('draft');
  const [requisitionId, setRequisitionId] = useState(preselectedReqId || '');
  const [reqs, setReqs] = useState<Requisition[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);

  useEffect(() => {
    api.reqs.list().then(r => setReqs(r.filter(req => req.status === 'open')));
  }, []);

  useEffect(() => {
    if (isEdit) {
      // Edit mode: load existing job
      api.jobs.get(jobId).then(job => {
        setTitle(job.title);
        setCompanyBlurb(job.company_blurb);
        setTeamDetails(job.team_details);
        setResponsibilities(job.responsibilities);
        setQualifications(job.qualifications);
        setClosingStatement(job.closing_statement);
        setLocation(job.location || '');
        setDepartment(job.department || '');
        setSalary(job.salary || '');
        setStatus(job.status);
        setRequisitionId(job.requisition_id || '');
      }).finally(() => setInitialLoading(false));
    } else {
      // Create mode: prepopulate from org defaults
      api.settings.getDefaults().then(defaults => {
        setCompanyBlurb(defaults.defaultCompanyBlurb);
        setClosingStatement(defaults.defaultClosingStatement);
      }).finally(() => setInitialLoading(false));
    }
  }, [jobId, isEdit]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      if (isEdit) {
        await api.jobs.update(jobId, {
          title,
          companyBlurb,
          teamDetails,
          responsibilities,
          qualifications,
          closingStatement,
          location: location || undefined,
          department: department || undefined,
          salary: salary || undefined,
          status,
          requisitionId: requisitionId || undefined,
        });
        navigate(`/jobs/${jobId}`);
      } else {
        const job = await api.jobs.create({
          title,
          companyBlurb,
          teamDetails,
          responsibilities,
          qualifications,
          closingStatement,
          location: location || undefined,
          department: department || undefined,
          salary: salary || undefined,
          requisitionId: requisitionId || undefined,
        });
        navigate(`/jobs/${job.id}`);
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  if (initialLoading) {
    return <DashboardLayout><p className="text-gray-500">Loading...</p></DashboardLayout>;
  }

  return (
    <DashboardLayout>
      <div className="max-w-2xl">
        <h2 className="text-2xl font-semibold text-gray-900 mb-6">
          {isEdit ? 'Edit Job' : 'Create Job'}
        </h2>

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
            <label className="block text-sm font-medium text-gray-700 mb-1">Company Blurb</label>
            <p className="text-xs text-gray-500 mb-1">Standard company overview — prepopulated from org defaults, editable per job.</p>
            <textarea
              value={companyBlurb}
              onChange={e => setCompanyBlurb(e.target.value)}
              rows={4}
              placeholder="About the company..."
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Team & Role Details *</label>
            <p className="text-xs text-gray-500 mb-1">Specific details about the team and role.</p>
            <textarea
              value={teamDetails}
              onChange={e => setTeamDetails(e.target.value)}
              required
              rows={6}
              placeholder="About the team and what this role involves..."
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Key Responsibilities</label>
            <textarea
              value={responsibilities}
              onChange={e => setResponsibilities(e.target.value)}
              rows={5}
              placeholder={"- Design and build scalable systems\n- Collaborate with cross-functional teams\n- Mentor junior engineers"}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Minimum Qualifications</label>
            <textarea
              value={qualifications}
              onChange={e => setQualifications(e.target.value)}
              rows={5}
              placeholder={"- 3+ years of experience in Go or similar\n- Strong understanding of distributed systems\n- BS in Computer Science or equivalent"}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Closing Statement</label>
            <p className="text-xs text-gray-500 mb-1">Standard closing — prepopulated from org defaults, editable per job.</p>
            <textarea
              value={closingStatement}
              onChange={e => setClosingStatement(e.target.value)}
              rows={3}
              placeholder="We are an equal opportunity employer..."
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
              {loading ? 'Saving...' : isEdit ? 'Save Changes' : 'Create Job (as Draft)'}
            </button>
            <button
              type="button"
              onClick={() => navigate(isEdit ? `/jobs/${jobId}` : '/jobs')}
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
