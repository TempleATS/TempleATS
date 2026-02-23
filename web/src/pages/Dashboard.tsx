import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api, type Requisition, type Job, type Candidate } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

export default function Dashboard() {
  const [reqs, setReqs] = useState<Requisition[]>([]);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [candidates, setCandidates] = useState<Candidate[]>([]);

  useEffect(() => {
    api.reqs.list().then(setReqs);
    api.jobs.list().then(setJobs);
    api.candidates.list().then(setCandidates);
  }, []);

  const openJobs = jobs.filter(j => j.status === 'open').length;
  const activeReqs = reqs.filter(r => r.status === 'open').length;

  return (
    <DashboardLayout>
      <h2 className="text-2xl font-semibold text-gray-900 mb-6">Dashboard</h2>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <Link to="/jobs" className="bg-white p-6 rounded-lg shadow-sm border hover:shadow-md transition-shadow">
          <p className="text-sm text-gray-500">Open Jobs</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{openJobs}</p>
        </Link>
        <Link to="/reqs" className="bg-white p-6 rounded-lg shadow-sm border hover:shadow-md transition-shadow">
          <p className="text-sm text-gray-500">Active Requisitions</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{activeReqs}</p>
        </Link>
        <Link to="/candidates" className="bg-white p-6 rounded-lg shadow-sm border hover:shadow-md transition-shadow">
          <p className="text-sm text-gray-500">Total Candidates</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{candidates.length}</p>
        </Link>
      </div>
    </DashboardLayout>
  );
}
