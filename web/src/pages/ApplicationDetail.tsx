import { useState, useEffect, type FormEvent } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, type ApplicationDetail as AppDetailType } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

function pgText(val: { String: string; Valid: boolean } | null): string | null {
  return val?.Valid ? val.String : null;
}

const stageColor = (stage: string) => {
  switch (stage) {
    case 'applied': return 'bg-blue-100 text-blue-800';
    case 'screening': return 'bg-yellow-100 text-yellow-800';
    case 'interview': return 'bg-purple-100 text-purple-800';
    case 'offer': return 'bg-green-100 text-green-800';
    case 'hired': return 'bg-emerald-100 text-emerald-800';
    case 'rejected': return 'bg-red-100 text-red-800';
    default: return 'bg-gray-100 text-gray-800';
  }
};

export default function ApplicationDetailPage() {
  const { appId } = useParams<{ appId: string }>();
  const [data, setData] = useState<AppDetailType | null>(null);
  const [loading, setLoading] = useState(true);
  const [noteContent, setNoteContent] = useState('');
  const [noteLoading, setNoteLoading] = useState(false);

  const loadApp = () => {
    if (!appId) return;
    api.applications.get(appId)
      .then(setData)
      .finally(() => setLoading(false));
  };

  useEffect(() => { loadApp(); }, [appId]);

  const handleAddNote = async (e: FormEvent) => {
    e.preventDefault();
    if (!appId || !noteContent.trim()) return;
    setNoteLoading(true);
    try {
      await api.applications.addNote(appId, noteContent.trim());
      setNoteContent('');
      loadApp();
    } finally {
      setNoteLoading(false);
    }
  };

  if (loading || !data) {
    return <DashboardLayout><p className="text-gray-500">Loading...</p></DashboardLayout>;
  }

  const { application: app, transitions, notes } = data;

  return (
    <DashboardLayout>
      <Link to={`/jobs/${app.job_id}/pipeline`} className="text-blue-600 hover:underline text-sm mb-4 inline-block">
        &larr; Back to pipeline
      </Link>

      <div className="grid grid-cols-3 gap-6">
        {/* Main info */}
        <div className="col-span-2 space-y-6">
          <div className="bg-white rounded-lg border p-6">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-xl font-semibold text-gray-900">{app.candidate_name}</h2>
                <p className="text-sm text-gray-500">{app.candidate_email}</p>
                {pgText(app.candidate_phone) && (
                  <p className="text-sm text-gray-500">{pgText(app.candidate_phone)}</p>
                )}
              </div>
              <span className={`text-xs px-3 py-1 rounded-full font-medium ${stageColor(app.stage)}`}>
                {app.stage}
              </span>
            </div>
            <div className="mt-3 text-sm text-gray-600">
              <p>Applied for: <span className="font-medium">{app.job_title}</span></p>
              {pgText(app.candidate_resume_url) && (
                <a href={pgText(app.candidate_resume_url)!} target="_blank" rel="noopener noreferrer"
                   className="text-blue-600 hover:underline mt-1 inline-block">
                  View Resume
                </a>
              )}
            </div>
            {app.stage === 'rejected' && pgText(app.rejection_reason) && (
              <div className="mt-4 p-3 bg-red-50 rounded border border-red-200">
                <p className="text-sm font-medium text-red-800">Rejected: {pgText(app.rejection_reason)}</p>
                {pgText(app.rejection_notes) && (
                  <p className="text-sm text-red-600 mt-1">{pgText(app.rejection_notes)}</p>
                )}
              </div>
            )}
          </div>

          {/* Notes */}
          <div className="bg-white rounded-lg border p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Notes</h3>
            <form onSubmit={handleAddNote} className="mb-4">
              <textarea
                value={noteContent}
                onChange={e => setNoteContent(e.target.value)}
                rows={3}
                placeholder="Add a note..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              />
              <button
                type="submit"
                disabled={!noteContent.trim() || noteLoading}
                className="mt-2 px-4 py-1.5 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm"
              >
                {noteLoading ? 'Adding...' : 'Add Note'}
              </button>
            </form>
            {notes.length === 0 ? (
              <p className="text-sm text-gray-400">No notes yet.</p>
            ) : (
              <div className="space-y-3">
                {notes.map(note => (
                  <div key={note.id} className="border-l-2 border-gray-200 pl-3">
                    <p className="text-sm text-gray-800">{note.content}</p>
                    <p className="text-xs text-gray-400 mt-1">
                      {note.author_name || 'Unknown'} &middot;{' '}
                      {note.created_at?.Time ? new Date(note.created_at.Time).toLocaleString() : ''}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Timeline */}
        <div className="bg-white rounded-lg border p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Timeline</h3>
          <div className="space-y-4">
            {transitions.map(t => (
              <div key={t.id} className="flex gap-3">
                <div className="w-2 h-2 rounded-full bg-blue-400 mt-1.5 flex-shrink-0" />
                <div>
                  <p className="text-sm text-gray-800">
                    {pgText(t.from_stage) ? (
                      <>{pgText(t.from_stage)} &rarr; <span className="font-medium">{t.to_stage}</span></>
                    ) : (
                      <span className="font-medium">Applied</span>
                    )}
                  </p>
                  <p className="text-xs text-gray-400">
                    {pgText(t.moved_by_name) && `${pgText(t.moved_by_name)} · `}
                    {t.created_at?.Time ? new Date(t.created_at.Time).toLocaleString() : ''}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}
