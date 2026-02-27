import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, type PipelineData } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

const REASON_LABELS: Record<string, string> = {
  not_qualified: 'Not Qualified',
  culture_fit: 'Culture Fit',
  salary_mismatch: 'Salary Mismatch',
  position_filled: 'Position Filled',
  withdrew: 'Withdrew',
  no_show: 'No Show',
  failed_assessment: 'Failed Assessment',
  other: 'Other',
};

function pgText(val: unknown): string | null {
  if (val === null || val === undefined) return null;
  if (typeof val === 'string') return val || null;
  if (typeof val === 'object' && val !== null && 'Valid' in val) {
    const t = val as { String: string; Valid: boolean };
    return t.Valid ? t.String : null;
  }
  return null;
}

export default function RejectedList() {
  const { jobId } = useParams<{ jobId: string }>();
  const [data, setData] = useState<PipelineData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!jobId) return;
    api.jobs.pipeline(jobId).then(setData).finally(() => setLoading(false));
  }, [jobId]);

  if (loading || !data) {
    return (
      <DashboardLayout>
        <p className="text-gray-500">Loading...</p>
      </DashboardLayout>
    );
  }

  const rejected = data.stages.rejected || [];

  return (
    <DashboardLayout>
      <div className="mb-4">
        <Link to={`/jobs/${jobId}/pipeline`} className="text-sm text-gray-500 hover:text-gray-700">
          &larr; Back to Pipeline
        </Link>
      </div>

      <h2 className="text-2xl font-semibold text-gray-900 mb-1">{data.job.title}</h2>
      <p className="text-sm text-gray-500 mb-6">Rejected candidates ({rejected.length})</p>

      {rejected.length === 0 ? (
        <p className="text-gray-500 text-sm">No rejected candidates.</p>
      ) : (
        <div className="bg-white rounded-lg border divide-y">
          {rejected.map(app => {
            const reason = pgText(app.rejection_reason);
            const notes = pgText(app.rejection_notes);
            return (
              <div key={app.id} className="flex items-center justify-between px-4 py-3">
                <div className="flex items-center gap-4">
                  <div>
                    <Link
                      to={`/applications/${app.id}`}
                      className="text-sm font-medium text-gray-900 hover:text-blue-600"
                    >
                      {app.candidate_name}
                    </Link>
                    <p className="text-xs text-gray-500">{app.candidate_email}</p>
                  </div>
                </div>
                <div className="text-right">
                  {reason && (
                    <span className="text-xs bg-red-50 text-red-700 px-2 py-0.5 rounded">
                      {REASON_LABELS[reason] || reason}
                    </span>
                  )}
                  {notes && (
                    <p className="text-xs text-gray-400 mt-0.5 max-w-xs truncate">{notes}</p>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </DashboardLayout>
  );
}
