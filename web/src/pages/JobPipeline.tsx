import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, type PipelineApplication, type PipelineData } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';
import KanbanBoard from '../components/pipeline/KanbanBoard';

const REJECTION_REASONS = [
  { value: 'not_qualified', label: 'Not Qualified' },
  { value: 'culture_fit', label: 'Culture Fit' },
  { value: 'salary_mismatch', label: 'Salary Mismatch' },
  { value: 'position_filled', label: 'Position Filled' },
  { value: 'withdrew', label: 'Withdrew' },
  { value: 'no_show', label: 'No Show' },
  { value: 'failed_assessment', label: 'Failed Assessment' },
  { value: 'other', label: 'Other' },
];

export default function JobPipeline() {
  const { jobId } = useParams<{ jobId: string }>();
  const [data, setData] = useState<PipelineData | null>(null);
  const [loading, setLoading] = useState(true);

  // Rejection dialog state
  const [rejectingApp, setRejectingApp] = useState<{ id: string; name: string } | null>(null);
  const [rejReason, setRejReason] = useState('');
  const [rejNotes, setRejNotes] = useState('');
  const [rejLoading, setRejLoading] = useState(false);

  const loadPipeline = useCallback(() => {
    if (!jobId) return;
    api.jobs.pipeline(jobId).then(setData).finally(() => setLoading(false));
  }, [jobId]);

  useEffect(() => { loadPipeline(); }, [loadPipeline]);

  const handleMoveStage = async (appId: string, newStage: string) => {
    if (newStage === 'rejected') {
      // Find the app to show name in dialog
      if (!data) return;
      for (const apps of Object.values(data.stages)) {
        const app = apps.find(a => a.id === appId);
        if (app) {
          setRejectingApp({ id: appId, name: app.candidate_name });
          return;
        }
      }
      return;
    }

    try {
      await api.applications.updateStage(appId, { stage: newStage });
      loadPipeline();
    } catch {
      // Reload to get consistent state
      loadPipeline();
    }
  };

  const handleReject = async () => {
    if (!rejectingApp || !rejReason) return;
    setRejLoading(true);
    try {
      await api.applications.updateStage(rejectingApp.id, {
        stage: 'rejected',
        rejectionReason: rejReason,
        rejectionNotes: rejNotes || undefined,
      });
      setRejectingApp(null);
      setRejReason('');
      setRejNotes('');
      loadPipeline();
    } catch {
      // stay on dialog
    } finally {
      setRejLoading(false);
    }
  };

  const handleCardClick = (_app: PipelineApplication) => {
    // Could navigate to application detail in future
  };

  if (loading || !data) {
    return (
      <DashboardLayout>
        <p className="text-gray-500">Loading pipeline...</p>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="flex items-center justify-between mb-6">
        <div>
          <Link to={`/jobs/${jobId}`} className="text-blue-600 hover:underline text-sm">
            &larr; Back to job
          </Link>
          <h2 className="text-2xl font-semibold text-gray-900 mt-1">{data.job.title} - Pipeline</h2>
        </div>
      </div>

      <KanbanBoard
        stages={data.stages}
        onMoveStage={handleMoveStage}
        onCardClick={handleCardClick}
      />

      {/* Rejection Dialog */}
      {rejectingApp && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900 mb-1">Reject Candidate</h3>
            <p className="text-sm text-gray-500 mb-4">Rejecting {rejectingApp.name}</p>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Reason *</label>
                <select
                  value={rejReason}
                  onChange={e => setRejReason(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Select reason...</option>
                  {REJECTION_REASONS.map(r => (
                    <option key={r.value} value={r.value}>{r.label}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Notes</label>
                <textarea
                  value={rejNotes}
                  onChange={e => setRejNotes(e.target.value)}
                  rows={3}
                  placeholder="Optional notes..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div className="flex gap-3 pt-2">
                <button
                  onClick={handleReject}
                  disabled={!rejReason || rejLoading}
                  className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50 text-sm font-medium"
                >
                  {rejLoading ? 'Rejecting...' : 'Reject'}
                </button>
                <button
                  onClick={() => { setRejectingApp(null); setRejReason(''); setRejNotes(''); }}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 text-sm"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </DashboardLayout>
  );
}
