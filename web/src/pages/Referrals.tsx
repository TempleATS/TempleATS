import { useState, useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import { api, type Referral, type Job } from '../api/client';
import { useAuth } from '../hooks/use-auth';
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

export default function Referrals() {
  const { user } = useAuth();
  const [referrals, setReferrals] = useState<Referral[]>([]);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);

  // Refer candidate form
  const [showReferForm, setShowReferForm] = useState(false);
  const [refName, setRefName] = useState('');
  const [refEmail, setRefEmail] = useState('');
  const [refPhone, setRefPhone] = useState('');
  const [refJobId, setRefJobId] = useState('');
  const [refResume, setRefResume] = useState<File | null>(null);
  const [refSending, setRefSending] = useState(false);
  const [refSuccess, setRefSuccess] = useState('');
  const [refError, setRefError] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Generate link form
  const [linkJobId, setLinkJobId] = useState('');
  const [generatedLink, setGeneratedLink] = useState('');
  const [linkCopied, setLinkCopied] = useState(false);
  const [linkError, setLinkError] = useState('');

  useEffect(() => {
    Promise.all([
      api.referrals.list().then(data => setReferrals(data || [])).catch(() => []),
      api.jobs.list().then(setJobs).catch(() => []),
    ]).finally(() => setLoading(false));
  }, []);

  const openJobs = jobs.filter(j => j.status === 'open');

  const handleRefer = async () => {
    if (!refName.trim() || !refEmail.trim() || !refJobId) return;
    setRefSending(true);
    setRefError('');
    setRefSuccess('');
    try {
      const result = await api.referrals.create({
        name: refName,
        email: refEmail,
        phone: refPhone || undefined,
        jobId: refJobId,
        resume: refResume || undefined,
      });
      setRefSuccess(`Referral created! Application ID: ${result.applicationId}`);
      setRefName('');
      setRefEmail('');
      setRefPhone('');
      setRefJobId('');
      setRefResume(null);
      // Refresh list
      api.referrals.list().then(data => setReferrals(data || [])).catch(() => {});
    } catch (err: any) {
      setRefError(err.message || 'Failed to create referral');
    } finally {
      setRefSending(false);
    }
  };

  const handleGenerateLink = async () => {
    if (!linkJobId) return;
    setLinkError('');
    setGeneratedLink('');
    setLinkCopied(false);
    try {
      const result = await api.referrals.createLink(linkJobId);
      const orgSlug = user?.orgSlug || '';
      const link = `${window.location.origin}/careers/${orgSlug}/jobs/${linkJobId}/apply?ref=${result.token}`;
      setGeneratedLink(link);
      // Refresh list
      api.referrals.list().then(data => setReferrals(data || [])).catch(() => {});
    } catch (err: any) {
      setLinkError(err.message || 'Failed to generate link');
    }
  };

  const copyLink = (link: string) => {
    navigator.clipboard.writeText(link);
    setLinkCopied(true);
    setTimeout(() => setLinkCopied(false), 2000);
  };

  const buildLinkForToken = (ref: Referral) => {
    const orgSlug = user?.orgSlug || '';
    return `${window.location.origin}/careers/${orgSlug}/jobs/${ref.job_id}/apply?ref=${ref.token}`;
  };

  if (loading) {
    return (
      <DashboardLayout>
        <p className="text-gray-500">Loading...</p>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="max-w-5xl">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold">Referrals</h1>
          <button
            onClick={() => { setShowReferForm(!showReferForm); setRefSuccess(''); setRefError(''); }}
            className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 text-sm font-medium"
          >
            {showReferForm ? 'Close Form' : '+ Refer a Candidate'}
          </button>
        </div>

        {/* Refer Candidate Form */}
        {showReferForm && (
          <div className="bg-white rounded-lg border p-6 mb-6">
            <h2 className="text-lg font-semibold mb-4">Refer a Candidate</h2>
            {refError && <div className="mb-4 p-3 bg-red-50 text-red-700 rounded text-sm">{refError}</div>}
            {refSuccess && <div className="mb-4 p-3 bg-green-50 text-green-700 rounded text-sm">{refSuccess}</div>}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Job *</label>
                <select
                  value={refJobId}
                  onChange={e => setRefJobId(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                >
                  <option value="">Select a job...</option>
                  {openJobs.map(j => (
                    <option key={j.id} value={j.id}>{j.title}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Name *</label>
                <input
                  type="text"
                  value={refName}
                  onChange={e => setRefName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                  placeholder="Candidate name"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Email *</label>
                <input
                  type="email"
                  value={refEmail}
                  onChange={e => setRefEmail(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                  placeholder="candidate@email.com"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
                <input
                  type="tel"
                  value={refPhone}
                  onChange={e => setRefPhone(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                  placeholder="Optional"
                />
              </div>
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-1">Resume</label>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".pdf,.doc,.docx"
                  onChange={e => setRefResume(e.target.files?.[0] || null)}
                  className="hidden"
                />
                <div className="flex items-center gap-3">
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    className="px-3 py-2 border border-gray-300 rounded-md text-sm hover:bg-gray-50"
                  >
                    {refResume ? refResume.name : 'Choose file...'}
                  </button>
                  {refResume && (
                    <button
                      type="button"
                      onClick={() => setRefResume(null)}
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
                onClick={handleRefer}
                disabled={refSending || !refName.trim() || !refEmail.trim() || !refJobId}
                className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
              >
                {refSending ? 'Submitting...' : 'Submit Referral'}
              </button>
            </div>
          </div>
        )}

        {/* Generate Referral Link */}
        <div className="bg-white rounded-lg border p-6 mb-6">
          <h2 className="text-lg font-semibold mb-4">Share Referral Link</h2>
          <p className="text-sm text-gray-600 mb-3">Generate a link to share with someone. When they apply through it, you get credit.</p>
          {linkError && <div className="mb-3 p-3 bg-red-50 text-red-700 rounded text-sm">{linkError}</div>}
          <div className="flex items-end gap-3">
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-1">Job</label>
              <select
                value={linkJobId}
                onChange={e => { setLinkJobId(e.target.value); setGeneratedLink(''); }}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
              >
                <option value="">Select a job...</option>
                {openJobs.map(j => (
                  <option key={j.id} value={j.id}>{j.title}</option>
                ))}
              </select>
            </div>
            <button
              onClick={handleGenerateLink}
              disabled={!linkJobId}
              className="px-4 py-2 bg-gray-800 text-white rounded-md hover:bg-gray-900 disabled:opacity-50 text-sm font-medium"
            >
              Generate Link
            </button>
          </div>
          {generatedLink && (
            <div className="mt-3 flex items-center gap-2">
              <input
                type="text"
                readOnly
                value={generatedLink}
                className="flex-1 px-3 py-2 bg-gray-50 border border-gray-300 rounded-md text-sm text-gray-700"
              />
              <button
                onClick={() => copyLink(generatedLink)}
                className="px-3 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 text-sm font-medium whitespace-nowrap"
              >
                {linkCopied ? 'Copied!' : 'Copy'}
              </button>
            </div>
          )}
        </div>

        {/* Referrals Table */}
        <div className="bg-white rounded-lg border">
          <div className="p-4 border-b">
            <h2 className="text-lg font-semibold">All Referrals</h2>
          </div>
          {referrals.length === 0 ? (
            <div className="p-8 text-center text-gray-400">
              <p>No referrals yet. Refer a candidate or share a referral link to get started.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-gray-50">
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Referrer</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Candidate</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Job</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Status</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Source</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Date</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600"></th>
                  </tr>
                </thead>
                <tbody>
                  {referrals.map(ref => {
                    const candidateName = pgText(ref.candidate_name);
                    const appId = pgText(ref.application_id);
                    return (
                      <tr key={ref.id} className="border-b hover:bg-gray-50">
                        <td className="px-4 py-3">{ref.referrer_name}</td>
                        <td className="px-4 py-3">
                          {candidateName ? (
                            appId ? (
                              <Link to={`/applications/${appId}`} className="text-indigo-600 hover:text-indigo-800 font-medium">
                                {candidateName}
                              </Link>
                            ) : candidateName
                          ) : (
                            <span className="text-gray-400 italic">Awaiting application</span>
                          )}
                        </td>
                        <td className="px-4 py-3">{ref.job_title}</td>
                        <td className="px-4 py-3">
                          {ref.application_stage ? (
                            <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                              ref.application_stage === 'offer' ? 'bg-green-100 text-green-700' :
                              ref.application_stage === 'rejected' ? 'bg-red-100 text-red-700' :
                              ref.application_stage === 'applied' ? 'bg-blue-100 text-blue-700' :
                              'bg-yellow-100 text-yellow-700'
                            }`}>
                              {STAGE_LABELS[ref.application_stage] || ref.application_stage}
                            </span>
                          ) : (
                            <span className="text-xs text-gray-400">-</span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                            ref.source === 'direct' ? 'bg-purple-100 text-purple-700' : 'bg-blue-100 text-blue-700'
                          }`}>
                            {ref.source === 'direct' ? 'Direct' : 'Link'}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-gray-500">
                          {new Date(ref.created_at.Time).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                        </td>
                        <td className="px-4 py-3">
                          {ref.source === 'link' && !candidateName && (
                            <button
                              onClick={() => copyLink(buildLinkForToken(ref))}
                              className="text-xs text-indigo-600 hover:text-indigo-800"
                            >
                              Copy link
                            </button>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
}
