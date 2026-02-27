import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api, type MyInterview } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';
import { STAGE_LABELS } from '../components/pipeline/KanbanBoard';

function pgText(val: unknown): string | null {
  if (val === null || val === undefined) return null;
  if (typeof val === 'string') return val || null;
  if (typeof val === 'object' && val !== null && 'Valid' in val) {
    const t = val as { String: string; Valid: boolean };
    return t.Valid ? t.String : null;
  }
  return null;
}

function pgTime(val: unknown): string | null {
  if (val === null || val === undefined) return null;
  if (typeof val === 'string') return val || null;
  if (typeof val === 'object' && val !== null && 'Time' in val) {
    const t = val as { Time: string; Valid: boolean };
    return t.Valid ? t.Time : null;
  }
  return null;
}

function stageColor(stage: string): string {
  switch (stage) {
    case 'applied': return 'bg-blue-100 text-blue-700';
    case 'hr_screen': return 'bg-purple-100 text-purple-700';
    case 'hm_review': return 'bg-indigo-100 text-indigo-700';
    case 'first_interview': return 'bg-yellow-100 text-yellow-700';
    case 'final_interview': return 'bg-orange-100 text-orange-700';
    case 'offer': return 'bg-green-100 text-green-700';
    case 'rejected': return 'bg-red-100 text-red-700';
    default: return 'bg-gray-100 text-gray-700';
  }
}

export default function Interviews() {
  const [interviews, setInterviews] = useState<MyInterview[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.interviews.mine()
      .then(data => setInterviews(data || []))
      .catch(() => [])
      .finally(() => setLoading(false));
  }, []);

  const active = interviews.filter(i => i.stage !== 'rejected' && i.stage !== 'offer');
  const past = interviews.filter(i => i.stage === 'rejected' || i.stage === 'offer');

  if (loading) {
    return <DashboardLayout><p className="text-gray-500">Loading...</p></DashboardLayout>;
  }

  return (
    <DashboardLayout>
      <div className="max-w-4xl">
        <h1 className="text-2xl font-bold mb-6">My Interviews</h1>

        {interviews.length === 0 ? (
          <div className="bg-white rounded-lg border p-8 text-center text-gray-400">
            <p>No interviews assigned to you yet.</p>
          </div>
        ) : (
          <div className="space-y-6">
            {/* Active interviews */}
            {active.length > 0 && (
              <div>
                <h2 className="text-sm font-medium text-gray-500 uppercase tracking-wide mb-3">Active ({active.length})</h2>
                <div className="space-y-3">
                  {active.map(interview => {
                    const resumeUrl = pgText(interview.candidate_resume_url);
                    const appliedAt = pgTime(interview.created_at);
                    return (
                      <Link
                        key={interview.id}
                        to={`/applications/${interview.id}`}
                        className="block bg-white rounded-lg border p-4 hover:border-blue-300 hover:shadow-sm transition-all"
                      >
                        <div className="flex items-center justify-between">
                          <div>
                            <p className="text-sm font-semibold text-gray-900">{interview.candidate_name}</p>
                            <p className="text-sm text-gray-600 mt-0.5">{interview.job_title}</p>
                          </div>
                          <div className="flex items-center gap-3">
                            {resumeUrl && (
                              <span
                                onClick={e => { e.preventDefault(); window.open(resumeUrl, '_blank'); }}
                                className="text-xs text-blue-600 hover:text-blue-800 font-medium cursor-pointer"
                              >
                                Resume
                              </span>
                            )}
                            <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${stageColor(interview.stage)}`}>
                              {STAGE_LABELS[interview.stage] || interview.stage}
                            </span>
                          </div>
                        </div>
                        {appliedAt && (
                          <p className="text-xs text-gray-400 mt-1">
                            Applied {new Date(appliedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                          </p>
                        )}
                      </Link>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Past interviews */}
            {past.length > 0 && (
              <div>
                <h2 className="text-sm font-medium text-gray-500 uppercase tracking-wide mb-3">Completed ({past.length})</h2>
                <div className="space-y-2">
                  {past.map(interview => (
                    <Link
                      key={interview.id}
                      to={`/applications/${interview.id}`}
                      className="block bg-white rounded-lg border p-3 hover:bg-gray-50 transition-colors"
                    >
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-sm text-gray-700">{interview.candidate_name}</p>
                          <p className="text-xs text-gray-500">{interview.job_title}</p>
                        </div>
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${stageColor(interview.stage)}`}>
                          {STAGE_LABELS[interview.stage] || interview.stage}
                        </span>
                      </div>
                    </Link>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
