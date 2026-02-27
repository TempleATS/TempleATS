import { useState, useEffect, useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, type CareerOrg, type CareerJob } from '../../api/client';

function pgText(val: unknown): string | null {
  if (val === null || val === undefined) return null;
  if (typeof val === 'string') return val || null;
  if (typeof val === 'object' && val !== null && 'Valid' in val) {
    const t = val as { String: string; Valid: boolean };
    return t.Valid ? t.String : null;
  }
  return null;
}

export default function CareersPage() {
  const { orgSlug } = useParams<{ orgSlug: string }>();
  const [org, setOrg] = useState<CareerOrg | null>(null);
  const [jobs, setJobs] = useState<CareerJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [locationFilter, setLocationFilter] = useState('');
  const [teamFilter, setTeamFilter] = useState('');

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

  const locations = useMemo(() => {
    const set = new Set<string>();
    for (const job of jobs) {
      const loc = pgText(job.location);
      if (loc) set.add(loc);
    }
    return Array.from(set).sort();
  }, [jobs]);

  const teams = useMemo(() => {
    const set = new Set<string>();
    for (const job of jobs) {
      const dept = pgText(job.department);
      if (dept) set.add(dept);
    }
    return Array.from(set).sort();
  }, [jobs]);

  const filteredJobs = useMemo(() => {
    return jobs.filter(job => {
      if (locationFilter && pgText(job.location) !== locationFilter) return false;
      if (teamFilter && pgText(job.department) !== teamFilter) return false;
      return true;
    });
  }, [jobs, locationFilter, teamFilter]);

  const groupedByTeam = useMemo(() => {
    const map = new Map<string, CareerJob[]>();
    for (const job of filteredJobs) {
      const team = pgText(job.department) || 'Other';
      const list = map.get(team) || [];
      list.push(job);
      map.set(team, list);
    }
    return Array.from(map.entries()).sort((a, b) => a[0].localeCompare(b[0]));
  }, [filteredJobs]);

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
        {jobs.length > 0 && (locations.length > 0 || teams.length > 0) && (
          <div className="flex gap-4 mb-6">
            {locations.length > 0 && (
              <select
                value={locationFilter}
                onChange={e => setLocationFilter(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-md bg-white text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">All Locations</option>
                {locations.map(loc => (
                  <option key={loc} value={loc}>{loc}</option>
                ))}
              </select>
            )}
            {teams.length > 0 && (
              <select
                value={teamFilter}
                onChange={e => setTeamFilter(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-md bg-white text-sm text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">All Teams</option>
                {teams.map(team => (
                  <option key={team} value={team}>{team}</option>
                ))}
              </select>
            )}
          </div>
        )}

        {filteredJobs.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg border">
            <p className="text-gray-500">
              {jobs.length === 0 ? 'No open positions at this time.' : 'No positions match the selected filters.'}
            </p>
          </div>
        ) : (
          <div className="space-y-8">
            {groupedByTeam.map(([team, teamJobs]) => (
              <div key={team}>
                <h3 className="text-lg font-semibold text-gray-900 mb-3 pb-2 border-b border-gray-200">{team}</h3>
                <div className="space-y-3">
                  {teamJobs.map(job => (
                    <div
                      key={job.id}
                      className="flex items-center justify-between bg-white rounded-lg border p-5 hover:shadow-md transition-shadow"
                    >
                      <Link to={`/careers/${orgSlug}/jobs/${job.id}`} className="flex-1 min-w-0">
                        <h2 className="text-base font-semibold text-gray-900">{job.title}</h2>
                        <div className="flex gap-4 mt-1.5 text-sm text-gray-500">
                          {pgText(job.location) && <span>{pgText(job.location)}</span>}
                          {pgText(job.salary) && <span>{pgText(job.salary)}</span>}
                        </div>
                      </Link>
                      <Link
                        to={`/careers/${orgSlug}/jobs/${job.id}`}
                        className="ml-4 shrink-0 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 transition-colors"
                      >
                        Apply
                      </Link>
                    </div>
                  ))}
                </div>
              </div>
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
