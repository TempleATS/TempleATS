import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, type Requisition, type Job } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

export default function ReqDetail() {
  const { reqId } = useParams<{ reqId: string }>();
  const [requisition, setRequisition] = useState<Requisition | null>(null);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!reqId) return;
    api.reqs.get(reqId)
      .then(data => {
        setRequisition(data.requisition);
        setJobs(data.jobs);
      })
      .finally(() => setLoading(false));
  }, [reqId]);

  if (loading) {
    return <DashboardLayout><p className="text-gray-500">Loading...</p></DashboardLayout>;
  }

  if (!requisition) {
    return <DashboardLayout><p className="text-gray-500">Requisition not found.</p></DashboardLayout>;
  }

  return (
    <DashboardLayout>
      <div className="mb-4">
        <Link to="/reqs" className="text-sm text-gray-500 hover:text-gray-700">&larr; Back to Requisitions</Link>
      </div>

      <div className="flex items-start justify-between mb-6">
        <div>
          <h2 className="text-2xl font-semibold text-gray-900">{requisition.title}</h2>
          <div className="flex gap-4 mt-2 text-sm text-gray-600">
            {requisition.level && <span>Level: {requisition.level}</span>}
            {requisition.department && <span>Dept: {requisition.department}</span>}
            <span>Target: {requisition.target_hires} hire(s)</span>
            <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
              requisition.status === 'open' ? 'bg-green-100 text-green-800' :
              requisition.status === 'filled' ? 'bg-blue-100 text-blue-800' :
              'bg-gray-100 text-gray-800'
            }`}>
              {requisition.status}
            </span>
          </div>
        </div>
        <Link
          to={`/reqs/${reqId}/report`}
          className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 text-sm"
        >
          View Report
        </Link>
      </div>

      <div className="bg-white rounded-lg border p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">Attached Jobs</h3>
          <Link
            to={`/jobs/new?reqId=${reqId}`}
            className="px-3 py-1.5 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
          >
            Create Job for this Req
          </Link>
        </div>

        {jobs.length === 0 ? (
          <p className="text-gray-500 text-sm">No jobs attached to this requisition yet.</p>
        ) : (
          <div className="space-y-3">
            {jobs.map(job => (
              <div key={job.id} className="flex items-center justify-between p-3 bg-gray-50 rounded border">
                <div>
                  <Link to={`/jobs/${job.id}`} className="text-blue-600 hover:underline font-medium">
                    {job.title}
                  </Link>
                  <div className="text-xs text-gray-500 mt-0.5">
                    {job.location || 'No location'} &middot; {job.status}
                  </div>
                </div>
                <Link
                  to={`/jobs/${job.id}/pipeline`}
                  className="text-sm text-blue-600 hover:underline"
                >
                  Pipeline
                </Link>
              </div>
            ))}
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
