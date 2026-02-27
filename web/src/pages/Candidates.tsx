import { useState, useEffect, useMemo, useRef } from 'react';
import { Link } from 'react-router-dom';
import { api, type CandidateListItem, type Job } from '../api/client';
import { useAuth } from '../hooks/use-auth';
import { STAGE_LABELS } from '../components/pipeline/KanbanBoard';
import DashboardLayout from '../components/layout/DashboardLayout';

type ResumeResult = {
  id: string;
  name: string;
  email: string;
  company: string;
  resume_filename: string;
  snippet: string;
};

const STAGES = ['applied', 'hr_screen', 'hm_review', 'first_interview', 'final_interview', 'offer', 'rejected'];
const REJECTION_REASONS = ['Not qualified', 'Position filled', 'Withdrew', 'No response', 'Culture fit', 'Compensation', 'Other'];

interface CandidateGroup {
  id: string;
  name: string;
  email: string;
  phone: unknown;
  resume_url: unknown;
  resume_filename: unknown;
  created_at: unknown;
  apps: { app_id: string; app_stage: string; job_title: string; applied_at: unknown }[];
}

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

const stageColor = (stage: string) => {
  switch (stage) {
    case 'applied': return 'bg-blue-100 text-blue-800';
    case 'hr_screen': return 'bg-cyan-100 text-cyan-800';
    case 'hm_review': return 'bg-yellow-100 text-yellow-800';
    case 'first_interview': return 'bg-purple-100 text-purple-800';
    case 'final_interview': return 'bg-indigo-100 text-indigo-800';
    case 'approval': return 'bg-amber-100 text-amber-800';
    case 'offer': return 'bg-green-100 text-green-800';
    case 'hired': return 'bg-emerald-100 text-emerald-800';
    case 'rejected': return 'bg-red-100 text-red-800';
    default: return 'bg-gray-100 text-gray-800';
  }
};

const stageBorderColor = (stage: string) => {
  switch (stage) {
    case 'applied': return 'border-blue-300 focus:ring-blue-400';
    case 'hr_screen': return 'border-cyan-300 focus:ring-cyan-400';
    case 'hm_review': return 'border-yellow-300 focus:ring-yellow-400';
    case 'first_interview': return 'border-purple-300 focus:ring-purple-400';
    case 'final_interview': return 'border-indigo-300 focus:ring-indigo-400';
    case 'offer': return 'border-green-300 focus:ring-green-400';
    case 'rejected': return 'border-red-300 focus:ring-red-400';
    default: return 'border-gray-300 focus:ring-gray-400';
  }
};

function groupByCandidateId(rows: CandidateListItem[]): CandidateGroup[] {
  const map = new Map<string, CandidateGroup>();
  for (const row of rows) {
    let group = map.get(row.id);
    if (!group) {
      group = {
        id: row.id,
        name: row.name,
        email: row.email,
        phone: row.phone,
        resume_url: row.resume_url,
        resume_filename: row.resume_filename,
        created_at: row.created_at,
        apps: [],
      };
      map.set(row.id, group);
    }
    if (row.app_id) {
      group.apps.push({
        app_id: row.app_id,
        app_stage: row.app_stage,
        job_title: row.job_title,
        applied_at: row.applied_at,
      });
    }
  }
  return Array.from(map.values());
}

export default function Candidates() {
  const { isAtLeast } = useAuth();
  const [rawRows, setRawRows] = useState<CandidateListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [selectedApp, setSelectedApp] = useState<Record<string, number>>({});

  // Resume search state
  const [resumeResults, setResumeResults] = useState<ResumeResult[]>([]);
  const [resumeCount, setResumeCount] = useState<number | null>(null);
  const [resumeSearching, setResumeSearching] = useState(false);
  const [resumeError, setResumeError] = useState('');

  // Add candidate form state
  const [showAddForm, setShowAddForm] = useState(false);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [addName, setAddName] = useState('');
  const [addEmail, setAddEmail] = useState('');
  const [addPhone, setAddPhone] = useState('');
  const [addJobId, setAddJobId] = useState('');
  const [addResume, setAddResume] = useState<File | null>(null);
  const [addSending, setAddSending] = useState(false);
  const [addError, setAddError] = useState('');
  const [addSuccess, setAddSuccess] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Rejection dialog state
  const [rejectingAppId, setRejectingAppId] = useState<string | null>(null);
  const [rejectionReason, setRejectionReason] = useState('');
  const [rejectionNotes, setRejectionNotes] = useState('');
  const [rejecting, setRejecting] = useState(false);

  const reload = () => {
    api.candidates.list().then(setRawRows);
  };

  useEffect(() => {
    setLoading(true);
    api.candidates.list()
      .then(setRawRows)
      .finally(() => setLoading(false));
  }, []);

  const runResumeSearch = async () => {
    if (!search.trim()) {
      setResumeResults([]);
      setResumeCount(null);
      setResumeError('');
      return;
    }
    setResumeSearching(true);
    setResumeError('');
    try {
      const res = await api.search.resumes(search);
      setResumeResults(res.results);
      setResumeCount(res.count);
    } catch (err: unknown) {
      setResumeError(err instanceof Error ? err.message : 'Search failed');
      setResumeResults([]);
      setResumeCount(null);
    } finally {
      setResumeSearching(false);
    }
  };

  const clearSearch = () => {
    setSearch('');
    setResumeResults([]);
    setResumeCount(null);
    setResumeError('');
  };

  useEffect(() => {
    if (showAddForm && jobs.length === 0) {
      api.jobs.list().then(setJobs).catch(() => {});
    }
  }, [showAddForm]);

  const openJobs = jobs.filter(j => j.status === 'open');

  const handleAddCandidate = async () => {
    if (!addName.trim() || !addEmail.trim()) return;
    setAddSending(true);
    setAddError('');
    setAddSuccess('');
    try {
      await api.candidates.create({
        name: addName,
        email: addEmail,
        phone: addPhone || undefined,
        jobId: addJobId || undefined,
        resume: addResume || undefined,
      });
      setAddSuccess('Candidate added successfully!');
      setAddName('');
      setAddEmail('');
      setAddPhone('');
      setAddJobId('');
      setAddResume(null);
      if (fileInputRef.current) fileInputRef.current.value = '';
      reload();
    } catch (err: any) {
      setAddError(err.message || 'Failed to add candidate');
    } finally {
      setAddSending(false);
    }
  };

  const candidates = useMemo(() => groupByCandidateId(rawRows), [rawRows]);

  const getActiveApp = (group: CandidateGroup) => {
    if (group.apps.length === 0) return null;
    const idx = selectedApp[group.id] ?? 0;
    return group.apps[idx] || group.apps[0];
  };

  const handleStageChange = async (appId: string, newStage: string) => {
    if (newStage === 'rejected') {
      setRejectingAppId(appId);
      setRejectionReason('');
      setRejectionNotes('');
      return;
    }
    try {
      await api.applications.updateStage(appId, { stage: newStage });
      setRawRows(prev => prev.map(r =>
        r.app_id === appId ? { ...r, app_stage: newStage } : r
      ));
    } catch {
      reload();
    }
  };

  const handleReject = async () => {
    if (!rejectingAppId || !rejectionReason) return;
    setRejecting(true);
    try {
      await api.applications.updateStage(rejectingAppId, {
        stage: 'rejected',
        rejectionReason,
        rejectionNotes: rejectionNotes || undefined,
      });
      setRawRows(prev => prev.map(r =>
        r.app_id === rejectingAppId ? { ...r, app_stage: 'rejected' } : r
      ));
      setRejectingAppId(null);
    } catch {
      reload();
    } finally {
      setRejecting(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-semibold text-gray-900">Candidates</h2>
        {isAtLeast('recruiter') && (
          <button
            onClick={() => { setShowAddForm(!showAddForm); setAddSuccess(''); setAddError(''); }}
            className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 text-sm font-medium"
          >
            {showAddForm ? 'Close Form' : '+ Add Candidate'}
          </button>
        )}
      </div>

      {showAddForm && (
        <div className="bg-white rounded-lg border p-6 mb-6">
          <h3 className="text-lg font-semibold mb-4">Add Candidate</h3>
          {addError && <div className="mb-4 p-3 bg-red-50 text-red-700 rounded text-sm">{addError}</div>}
          {addSuccess && <div className="mb-4 p-3 bg-green-50 text-green-700 rounded text-sm">{addSuccess}</div>}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name *</label>
              <input
                type="text"
                value={addName}
                onChange={e => setAddName(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                placeholder="Candidate name"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Email *</label>
              <input
                type="email"
                value={addEmail}
                onChange={e => setAddEmail(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                placeholder="candidate@email.com"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
              <input
                type="tel"
                value={addPhone}
                onChange={e => setAddPhone(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                placeholder="Optional"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Job</label>
              <select
                value={addJobId}
                onChange={e => setAddJobId(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
              >
                <option value="">No job (add to pool)</option>
                {openJobs.map(j => (
                  <option key={j.id} value={j.id}>{j.title}</option>
                ))}
              </select>
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">Resume</label>
              <input
                ref={fileInputRef}
                type="file"
                accept=".pdf,.doc,.docx"
                onChange={e => setAddResume(e.target.files?.[0] || null)}
                className="hidden"
              />
              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  className="px-3 py-2 border border-gray-300 rounded-md text-sm hover:bg-gray-50"
                >
                  {addResume ? addResume.name : 'Choose file...'}
                </button>
                {addResume && (
                  <button
                    type="button"
                    onClick={() => { setAddResume(null); if (fileInputRef.current) fileInputRef.current.value = ''; }}
                    className="text-sm text-red-600 hover:text-red-700"
                  >
                    Remove
                  </button>
                )}
              </div>
              <p className="text-xs text-gray-400 mt-1">PDF, DOC, or DOCX (max 10MB)</p>
            </div>
          </div>
          <div className="mt-4">
            <button
              onClick={handleAddCandidate}
              disabled={!addName.trim() || !addEmail.trim() || addSending}
              className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
            >
              {addSending ? 'Adding...' : 'Add Candidate'}
            </button>
          </div>
        </div>
      )}

      <div className="mb-4">
        <p className="text-xs text-gray-500 mb-1">
          Search resumes with <code className="bg-gray-100 px-1 rounded">AND</code>, <code className="bg-gray-100 px-1 rounded">OR</code>, <code className="bg-gray-100 px-1 rounded">NOT</code>, <code className="bg-gray-100 px-1 rounded">"phrases"</code>
        </p>
        <div className="flex gap-2 max-w-2xl">
          <input
            type="text"
            value={search}
            onChange={e => setSearch(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') runResumeSearch(); }}
            placeholder='e.g. python AND "machine learning" NOT junior'
            className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
          />
          <button
            onClick={runResumeSearch}
            disabled={resumeSearching || !search.trim()}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
          >
            {resumeSearching ? 'Searching...' : 'Search'}
          </button>
          {resumeCount !== null && (
            <button
              onClick={clearSearch}
              className="px-3 py-2 text-sm text-gray-500 hover:text-gray-700"
            >
              Clear
            </button>
          )}
        </div>
      </div>

      {resumeError && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{resumeError}</div>
      )}

      {resumeCount !== null && (
        <>
          <p className="text-sm text-gray-500 mb-3">{resumeCount} candidate{resumeCount !== 1 ? 's' : ''} found</p>
          {resumeResults.length > 0 ? (
            <div className="space-y-3 mb-6">
              {resumeResults.map(r => (
                <div key={r.id} className="bg-white border border-gray-200 rounded-lg p-4 hover:border-blue-300 transition-colors">
                  <div className="flex items-start justify-between">
                    <div>
                      <Link to={`/candidates/${r.id}`} className="text-base font-medium text-blue-600 hover:underline">{r.name}</Link>
                      <p className="text-sm text-gray-600 mt-0.5">{r.email}{r.company ? ` · ${r.company}` : ''}</p>
                      {r.resume_filename && <p className="text-xs text-gray-400 mt-0.5">{r.resume_filename}</p>}
                    </div>
                  </div>
                  {r.snippet && (
                    <p className="text-sm text-gray-500 mt-2 whitespace-pre-line line-clamp-4">{r.snippet}</p>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 bg-white rounded-lg border mb-6">
              <p className="text-gray-500">No resumes match your search.</p>
            </div>
          )}
        </>
      )}

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
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Job</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Stage</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Resume</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Applied</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {candidates.map(group => {
                const active = getActiveApp(group);
                return (
                  <tr key={group.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      {active ? (
                        <Link to={`/applications/${active.app_id}`} className="text-blue-600 hover:underline font-medium">
                          {group.name}
                        </Link>
                      ) : (
                        <span className="font-medium text-gray-900">{group.name}</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">{group.email}</td>
                    <td className="px-4 py-3">
                      {group.apps.length > 1 ? (
                        <select
                          value={selectedApp[group.id] ?? 0}
                          onChange={e => setSelectedApp(prev => ({ ...prev, [group.id]: Number(e.target.value) }))}
                          className="text-sm border border-gray-300 rounded-md px-2 py-1 focus:outline-none focus:ring-2 focus:ring-blue-500 max-w-[180px]"
                        >
                          {group.apps.map((app, i) => (
                            <option key={app.app_id} value={i}>{app.job_title || 'Untitled'}</option>
                          ))}
                        </select>
                      ) : active ? (
                        <span className="text-sm text-gray-600">{active.job_title || '-'}</span>
                      ) : (
                        <span className="text-sm text-gray-400">-</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {active ? (
                        <select
                          value={active.app_stage}
                          onChange={e => handleStageChange(active.app_id, e.target.value)}
                          className={`text-xs font-medium px-2 py-1 rounded-full border cursor-pointer focus:outline-none focus:ring-2 ${stageColor(active.app_stage)} ${stageBorderColor(active.app_stage)}`}
                        >
                          {STAGES.map(s => (
                            <option key={s} value={s}>{STAGE_LABELS[s] || s}</option>
                          ))}
                        </select>
                      ) : (
                        <span className="text-xs text-gray-400">No Application</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {pgText(group.resume_url) ? (
                        <a href={pgText(group.resume_url)!} target="_blank" rel="noopener noreferrer"
                           className="text-blue-600 hover:underline text-sm">
                          {pgText(group.resume_filename) || 'View'}
                        </a>
                      ) : (
                        <span className="text-sm text-gray-400">-</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">
                      {active && pgTime(active.applied_at)
                        ? new Date(pgTime(active.applied_at)!).toLocaleDateString()
                        : pgTime(group.created_at)
                          ? new Date(pgTime(group.created_at)!).toLocaleDateString()
                          : '-'}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Rejection Dialog */}
      {rejectingAppId && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Reject Candidate</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Reason *</label>
                <select
                  value={rejectionReason}
                  onChange={e => setRejectionReason(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                >
                  <option value="">Select a reason...</option>
                  {REJECTION_REASONS.map(r => (
                    <option key={r} value={r}>{r}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Notes (optional)</label>
                <textarea
                  value={rejectionNotes}
                  onChange={e => setRejectionNotes(e.target.value)}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  placeholder="Additional notes..."
                />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => setRejectingAppId(null)}
                  className="px-4 py-2 text-sm text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
                >
                  Cancel
                </button>
                <button
                  onClick={handleReject}
                  disabled={!rejectionReason || rejecting}
                  className="px-4 py-2 text-sm text-white bg-red-600 rounded-md hover:bg-red-700 disabled:opacity-50"
                >
                  {rejecting ? 'Rejecting...' : 'Reject'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </DashboardLayout>
  );
}
