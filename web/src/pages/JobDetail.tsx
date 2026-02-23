import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, type Job } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

export default function JobDetail() {
  const { jobId } = useParams<{ jobId: string }>();
  const [job, setJob] = useState<Job | null>(null);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);

  useEffect(() => {
    if (!jobId) return;
    api.jobs.get(jobId).then(setJob).finally(() => setLoading(false));
  }, [jobId]);

  const toggleStatus = async () => {
    if (!job || !jobId) return;
    setUpdating(true);
    try {
      const newStatus = job.status === 'open' ? 'closed' : job.status === 'closed' ? 'draft' : 'open';
      const updated = await api.jobs.update(jobId, {
        title: job.title,
        description: job.description,
        location: job.location || undefined,
        department: job.department || undefined,
        salary: job.salary || undefined,
        status: newStatus,
        requisitionId: job.requisition_id || undefined,
      });
      setJob(updated);
    } finally {
      setUpdating(false);
    }
  };

  if (loading) {
    return <DashboardLayout><p className="text-gray-500">Loading...</p></DashboardLayout>;
  }

  if (!job) {
    return <DashboardLayout><p className="text-gray-500">Job not found.</p></DashboardLayout>;
  }

  const statusLabel = job.status === 'draft' ? 'Publish' : job.status === 'open' ? 'Close' : 'Reopen as Draft';

  return (
    <DashboardLayout>
      <div className="mb-4">
        <Link to="/jobs" className="text-sm text-gray-500 hover:text-gray-700">&larr; Back to Jobs</Link>
      </div>

      <div className="flex items-start justify-between mb-6">
        <div>
          <h2 className="text-2xl font-semibold text-gray-900">{job.title}</h2>
          <div className="flex gap-4 mt-2 text-sm text-gray-600">
            {job.location && <span>{job.location}</span>}
            {job.department && <span>{job.department}</span>}
            {job.salary && <span>{job.salary}</span>}
            <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
              job.status === 'open' ? 'bg-green-100 text-green-800' :
              job.status === 'draft' ? 'bg-yellow-100 text-yellow-800' :
              'bg-gray-100 text-gray-800'
            }`}>
              {job.status}
            </span>
          </div>
        </div>
        <div className="flex gap-2">
          <button
            onClick={toggleStatus}
            disabled={updating}
            className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 text-sm disabled:opacity-50"
          >
            {statusLabel}
          </button>
          <Link
            to={`/jobs/${jobId}/pipeline`}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
          >
            View Pipeline
          </Link>
        </div>
      </div>

      <div className="bg-white rounded-lg border p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-3">Description</h3>
        <div className="prose prose-sm max-w-none text-gray-700 whitespace-pre-wrap">
          {job.description}
        </div>
      </div>
    </DashboardLayout>
  );
}
