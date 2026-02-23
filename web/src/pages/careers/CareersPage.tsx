import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, type CareerOrg, type CareerJob } from '../../api/client';

function pgText(val: { String: string; Valid: boolean } | null): string | null {
  return val?.Valid ? val.String : null;
}

export default function CareersPage() {
  const { orgSlug } = useParams<{ orgSlug: string }>();
  const [org, setOrg] = useState<CareerOrg | null>(null);
  const [jobs, setJobs] = useState<CareerJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!orgSlug) return;
    api.careers.listJobs(orgSlug)
      .then(data => {
        setOrg(data.organization);
        setJobs(data.jobs);
      })
      .catch(() => setError('Company not found'))
      .finally(() => setLoading(false));
  }, [orgSlug]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Loading...</p>
      </div>
    );
  }

  if (error || !org) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Page Not Found</h1>
          <p className="text-gray-500">This careers page doesn't exist.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <div className="max-w-4xl mx-auto px-4 py-8">
          <h1 className="text-3xl font-bold text-gray-900">{org.name}</h1>
          <p className="text-gray-600 mt-1">Open Positions</p>
          {org.website && (
            <a href={org.website} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline text-sm mt-2 inline-block">
              {org.website}
            </a>
          )}
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        {jobs.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg border">
            <p className="text-gray-500">No open positions at this time.</p>
          </div>
        ) : (
          <div className="space-y-4">
            {jobs.map(job => (
              <Link
                key={job.id}
                to={`/careers/${orgSlug}/jobs/${job.id}`}
                className="block bg-white rounded-lg border p-6 hover:shadow-md transition-shadow"
              >
                <h2 className="text-lg font-semibold text-gray-900">{job.title}</h2>
                <div className="flex gap-4 mt-2 text-sm text-gray-500">
                  {pgText(job.location) && <span>{pgText(job.location)}</span>}
                  {pgText(job.department) && <span>{pgText(job.department)}</span>}
                  {pgText(job.salary) && <span>{pgText(job.salary)}</span>}
                </div>
              </Link>
            ))}
          </div>
        )}
      </main>

      <footer className="max-w-4xl mx-auto px-4 py-8 text-center text-sm text-gray-400">
        Powered by TempleATS
      </footer>
    </div>
  );
}
