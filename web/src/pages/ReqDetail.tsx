import { useState, useEffect, useRef } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { api, type Requisition, type Job, type TeamMember } from '../api/client';
import { useAuth } from '../hooks/use-auth';
import DashboardLayout from '../components/layout/DashboardLayout';

export default function ReqDetail() {
  const { reqId } = useParams<{ reqId: string }>();
  const navigate = useNavigate();
  const { isAtLeast } = useAuth();
  const [requisition, setRequisition] = useState<Requisition | null>(null);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [managers, setManagers] = useState<TeamMember[]>([]);
  const [recruiters, setRecruiters] = useState<TeamMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showPicker, setShowPicker] = useState(false);
  const [showRecruiterPicker, setShowRecruiterPicker] = useState(false);
  const pickerRef = useRef<HTMLDivElement>(null);
  const recruiterPickerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!showPicker && !showRecruiterPicker) return;
    const handleClick = (e: MouseEvent) => {
      if (showPicker && pickerRef.current && !pickerRef.current.contains(e.target as Node)) {
        setShowPicker(false);
      }
      if (showRecruiterPicker && recruiterPickerRef.current && !recruiterPickerRef.current.contains(e.target as Node)) {
        setShowRecruiterPicker(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [showPicker, showRecruiterPicker]);

  useEffect(() => {
    if (!reqId) return;
    Promise.all([
      api.reqs.get(reqId),
      api.team.list(),
    ]).then(([reqData, teamData]) => {
      setRequisition(reqData.requisition);
      setJobs(reqData.jobs);
      const eligible = teamData.members.filter(m =>
        m.role === 'hiring_manager' || m.role === 'admin' || m.role === 'super_admin'
      );
      setManagers(eligible);
      const eligibleRecruiters = teamData.members.filter(m =>
        m.role === 'recruiter' || m.role === 'admin' || m.role === 'super_admin'
      );
      setRecruiters(eligibleRecruiters);
    }).finally(() => setLoading(false));
  }, [reqId]);

  const hmName = managers.find(m => m.id === requisition?.hiring_manager_id)?.name || 'Unassigned';
  const recruiterName = recruiters.find(m => m.id === requisition?.recruiter_id)?.name || 'Unassigned';

  const handleReassign = async (newHmId: string) => {
    if (!requisition || !reqId) return;
    setSaving(true);
    try {
      const updated = await api.reqs.update(reqId, {
        title: requisition.title,
        jobCode: requisition.job_code || undefined,
        level: requisition.level || undefined,
        department: requisition.department || undefined,
        targetHires: requisition.target_hires,
        status: requisition.status,
        hiringManagerId: newHmId || undefined,
      });
      setRequisition(updated);
    } finally {
      setSaving(false);
    }
  };

  const handleReassignRecruiter = async (newRecruiterId: string) => {
    if (!requisition || !reqId) return;
    setSaving(true);
    try {
      const updated = await api.reqs.update(reqId, {
        title: requisition.title,
        jobCode: requisition.job_code || undefined,
        level: requisition.level || undefined,
        department: requisition.department || undefined,
        targetHires: requisition.target_hires,
        status: requisition.status,
        recruiterId: newRecruiterId,
      });
      setRequisition(updated);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!reqId || !confirm('Delete this requisition? Attached jobs will be unlinked.')) return;
    await api.reqs.delete(reqId);
    navigate('/reqs');
  };

  if (loading) {
    return <DashboardLayout><p className="text-gray-500">Loading...</p></DashboardLayout>;
  }

  if (!requisition) {
    return <DashboardLayout><p className="text-gray-500">Requisition not found.</p></DashboardLayout>;
  }

  return (
    <DashboardLayout>
      <div className="mb-4">
        <Link to="/reqs" className="text-sm text-gray-500 hover:text-gray-700">&larr; Back to Requisitions</Link>
      </div>

      <div className="flex items-start justify-between mb-6">
        <div>
          <h2 className="text-2xl font-semibold text-gray-900">{requisition.title}</h2>
          <div className="flex gap-4 mt-2 text-sm text-gray-600">
            {requisition.job_code && <span>Job Code: {requisition.job_code}</span>}
            {requisition.level && <span>Level: {requisition.level}</span>}
            {requisition.department && <span>Dept: {requisition.department}</span>}
            <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
              requisition.status === 'open' ? 'bg-green-100 text-green-800' :
              requisition.status === 'filled' ? 'bg-blue-100 text-blue-800' :
              'bg-gray-100 text-gray-800'
            }`}>
              {requisition.status}
            </span>
          </div>
        </div>
        <div className="flex gap-2">
          <Link
            to={`/reqs/${reqId}/report`}
            className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 text-sm"
          >
            View Report
          </Link>
          {isAtLeast('admin') && (
            <button
              onClick={handleDelete}
              className="px-4 py-2 bg-red-50 text-red-600 rounded-md hover:bg-red-100 text-sm font-medium"
            >
              Delete
            </button>
          )}
        </div>
      </div>

      {/* Hiring Manager */}
      <div className="bg-white rounded-lg border p-6 mb-6">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-lg font-medium text-gray-900">Hiring Manager</h3>
          {isAtLeast('admin') && (
            <div ref={pickerRef} className="relative">
              <button
                onClick={() => setShowPicker(!showPicker)}
                className="w-7 h-7 flex items-center justify-center bg-blue-600 text-white rounded-full hover:bg-blue-700 text-lg font-bold leading-none"
              >
                +
              </button>
              {showPicker && (
                <div className="absolute right-0 top-9 z-10 w-64 border rounded-md bg-white shadow-lg divide-y max-h-48 overflow-y-auto">
                  {managers.filter(m => m.id !== requisition.hiring_manager_id).map(m => (
                    <button
                      key={m.id}
                      onClick={() => { handleReassign(m.id); setShowPicker(false); }}
                      disabled={saving}
                      className="w-full text-left px-3 py-2 hover:bg-gray-50 text-sm disabled:opacity-50"
                    >
                      <span className="font-medium text-gray-900">{m.name}</span>
                      <span className="ml-2 text-gray-400">{m.role.replace('_', ' ')}</span>
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
        {requisition.hiring_manager_id ? (
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-50 text-blue-800 rounded-full text-sm font-medium">
              {hmName}
              {isAtLeast('admin') && (
                <button
                  onClick={() => handleReassign('')}
                  className="ml-1 text-blue-400 hover:text-blue-600"
                  title="Remove"
                >
                  &times;
                </button>
              )}
            </span>
            {saving && <span className="text-sm text-gray-400">Saving...</span>}
          </div>
        ) : (
          <p className="text-sm text-gray-500">No hiring manager assigned</p>
        )}
      </div>

      {/* Recruiter */}
      <div className="bg-white rounded-lg border p-6 mb-6">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-lg font-medium text-gray-900">Recruiter</h3>
          {isAtLeast('admin') && (
            <div ref={recruiterPickerRef} className="relative">
              <button
                onClick={() => setShowRecruiterPicker(!showRecruiterPicker)}
                className="w-7 h-7 flex items-center justify-center bg-orange-500 text-white rounded-full hover:bg-orange-600 text-lg font-bold leading-none"
              >
                +
              </button>
              {showRecruiterPicker && (
                <div className="absolute right-0 top-9 z-10 w-64 border rounded-md bg-white shadow-lg divide-y max-h-48 overflow-y-auto">
                  {recruiters.filter(m => m.id !== requisition.recruiter_id).map(m => (
                    <button
                      key={m.id}
                      onClick={() => { handleReassignRecruiter(m.id); setShowRecruiterPicker(false); }}
                      disabled={saving}
                      className="w-full text-left px-3 py-2 hover:bg-gray-50 text-sm disabled:opacity-50"
                    >
                      <span className="font-medium text-gray-900">{m.name}</span>
                      <span className="ml-2 text-gray-400">{m.role.replace('_', ' ')}</span>
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
        {requisition.recruiter_id ? (
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-orange-50 text-orange-800 rounded-full text-sm font-medium">
              {recruiterName}
              {isAtLeast('admin') && (
                <button
                  onClick={() => handleReassignRecruiter('')}
                  className="ml-1 text-orange-400 hover:text-orange-600"
                  title="Remove"
                >
                  &times;
                </button>
              )}
            </span>
            {saving && <span className="text-sm text-gray-400">Saving...</span>}
          </div>
        ) : (
          <p className="text-sm text-gray-500">No recruiter assigned</p>
        )}
      </div>

      <div className="bg-white rounded-lg border p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">Attached Jobs</h3>
          <Link
            to={`/jobs/new?reqId=${reqId}`}
            className="px-3 py-1.5 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
          >
            Create Job for this Req
          </Link>
        </div>

        {jobs.length === 0 ? (
          <p className="text-gray-500 text-sm">No jobs attached to this requisition yet.</p>
        ) : (
          <div className="space-y-3">
            {jobs.map(job => (
              <div key={job.id} className="flex items-center justify-between p-3 bg-gray-50 rounded border">
                <div>
                  <Link to={`/jobs/${job.id}`} className="text-blue-600 hover:underline font-medium">
                    {job.title}
                  </Link>
                  <div className="text-xs text-gray-500 mt-0.5">
                    {job.location || 'No location'} &middot; {job.status}
                  </div>
                </div>
                <Link
                  to={`/jobs/${job.id}/pipeline`}
                  className="text-sm text-blue-600 hover:underline"
                >
                  Pipeline
                </Link>
              </div>
            ))}
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
