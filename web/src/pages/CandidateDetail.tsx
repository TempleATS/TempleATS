import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, type Candidate, type CandidateApplication } from '../api/client';
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

export default function CandidateDetail() {
  const { candidateId } = useParams<{ candidateId: string }>();
  const [candidate, setCandidate] = useState<Candidate | null>(null);
  const [applications, setApplications] = useState<CandidateApplication[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!candidateId) return;
    api.candidates.get(candidateId)
      .then(data => {
        setCandidate(data.candidate);
        setApplications(data.applications);
      })
      .finally(() => setLoading(false));
  }, [candidateId]);

  if (loading) {
    return <DashboardLayout><p className="text-gray-500">Loading...</p></DashboardLayout>;
  }

  if (!candidate) {
    return <DashboardLayout><p className="text-gray-500">Candidate not found.</p></DashboardLayout>;
  }

  return (
    <DashboardLayout>
      <Link to="/candidates" className="text-blue-600 hover:underline text-sm mb-4 inline-block">
        &larr; All candidates
      </Link>

      <div className="bg-white rounded-lg border p-6 mb-6">
        <h2 className="text-2xl font-semibold text-gray-900">{candidate.name}</h2>
        <div className="mt-2 space-y-1 text-sm text-gray-600">
          <p>{candidate.email}</p>
          {pgText(candidate.phone) && <p>{pgText(candidate.phone)}</p>}
          {pgText(candidate.resume_url) && (
            <a href={pgText(candidate.resume_url)!} target="_blank" rel="noopener noreferrer"
               className="text-blue-600 hover:underline">
              View Resume
            </a>
          )}
        </div>
      </div>

      <h3 className="text-lg font-semibold text-gray-900 mb-4">Applications</h3>
      {applications.length === 0 ? (
        <p className="text-gray-500 text-sm">No applications.</p>
      ) : (
        <div className="space-y-3">
          {applications.map(app => (
            <div key={app.id} className="bg-white rounded-lg border p-4 flex items-center justify-between">
              <div>
                <p className="font-medium text-gray-900">{app.job_title}</p>
                <p className="text-xs text-gray-500 mt-1">
                  Applied {app.created_at?.Time ? new Date(app.created_at.Time).toLocaleDateString() : ''}
                </p>
              </div>
              <div className="flex items-center gap-3">
                <span className={`text-xs px-2 py-1 rounded-full font-medium ${stageColor(app.stage)}`}>
                  {app.stage}
                </span>
                {pgText(app.rejection_reason) && (
                  <span className="text-xs text-red-500">{pgText(app.rejection_reason)}</span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </DashboardLayout>
  );
}
