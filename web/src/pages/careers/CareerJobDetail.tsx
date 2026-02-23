import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, type CareerJob } from '../../api/client';

function pgText(val: { String: string; Valid: boolean } | null): string | null {
  return val?.Valid ? val.String : null;
}

export default function CareerJobDetail() {
  const { orgSlug, jobId } = useParams<{ orgSlug: string; jobId: string }>();
  const [job, setJob] = useState<CareerJob | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!orgSlug || !jobId) return;
    api.careers.getJob(orgSlug, jobId)
      .then(setJob)
      .catch(() => setError('Job not found'))
      .finally(() => setLoading(false));
  }, [orgSlug, jobId]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Loading...</p>
      </div>
    );
  }

  if (error || !job) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Job Not Found</h1>
          <p className="text-gray-500">This job posting doesn't exist or is no longer open.</p>
          <Link to={`/careers/${orgSlug}`} className="text-blue-600 hover:underline text-sm mt-4 inline-block">
            Back to all positions
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <div className="max-w-4xl mx-auto px-4 py-8">
          <Link to={`/careers/${orgSlug}`} className="text-blue-600 hover:underline text-sm mb-4 inline-block">
            &larr; All positions
          </Link>
          <h1 className="text-3xl font-bold text-gray-900">{job.title}</h1>
          <div className="flex gap-4 mt-2 text-sm text-gray-500">
            {pgText(job.location) && <span>{pgText(job.location)}</span>}
            {pgText(job.department) && <span>{pgText(job.department)}</span>}
            {pgText(job.salary) && <span>{pgText(job.salary)}</span>}
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <div className="bg-white rounded-lg border p-8">
          <div className="prose max-w-none whitespace-pre-wrap text-gray-700">
            {job.description}
          </div>

          <div className="mt-8 pt-6 border-t">
            <Link
              to={`/careers/${orgSlug}/jobs/${job.id}/apply`}
              className="inline-block px-6 py-3 bg-blue-600 text-white rounded-md hover:bg-blue-700 font-medium"
            >
              Apply for this position
            </Link>
          </div>
        </div>
      </main>

      <footer className="max-w-4xl mx-auto px-4 py-8 text-center text-sm text-gray-400">
        Powered by TempleATS
      </footer>
    </div>
  );
}
