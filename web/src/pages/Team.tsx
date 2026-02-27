import { useState, useEffect, type FormEvent } from 'react';
import { api, type TeamData } from '../api/client';
import { useAuth } from '../hooks/use-auth';
import DashboardLayout from '../components/layout/DashboardLayout';

const ROLE_LABELS: Record<string, string> = {
  super_admin: 'Super Admin',
  admin: 'Admin',
  recruiter: 'Recruiter',
  hiring_manager: 'Hiring Manager',
  interviewer: 'Interviewer',
};

const INVITABLE_ROLES_SUPER = ['admin', 'recruiter', 'hiring_manager', 'interviewer'];
const INVITABLE_ROLES_ADMIN = ['recruiter', 'hiring_manager', 'interviewer'];
const INVITABLE_ROLES_RECRUITER = ['hiring_manager', 'interviewer'];

export default function Team() {
  const { user, isAtLeast } = useAuth();
  const [data, setData] = useState<TeamData | null>(null);
  const [loading, setLoading] = useState(true);

  // Invite form
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('interviewer');
  const [inviting, setInviting] = useState(false);
  const [inviteMsg, setInviteMsg] = useState('');

  const canInvite = isAtLeast('recruiter');
  const canManage = isAtLeast('admin');
  const invitableRoles = user?.role === 'super_admin' ? INVITABLE_ROLES_SUPER
    : user?.role === 'admin' ? INVITABLE_ROLES_ADMIN
    : INVITABLE_ROLES_RECRUITER;

  const loadTeam = () => {
    api.team.list().then(setData).finally(() => setLoading(false));
  };

  useEffect(() => { loadTeam(); }, []);

  const handleInvite = async (e: FormEvent) => {
    e.preventDefault();
    setInviting(true);
    setInviteMsg('');
    try {
      const inv = await api.team.invite({ email, role });
      setInviteMsg(`Invitation sent! Share this link: ${window.location.origin}/accept-invite/${inv.token}`);
      setEmail('');
      loadTeam();
    } catch (err: any) {
      setInviteMsg(err.message);
    } finally {
      setInviting(false);
    }
  };

  const handleRemove = async (userId: string) => {
    if (!confirm('Remove this team member?')) return;
    try {
      await api.team.remove(userId);
      loadTeam();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleRoleChange = async (userId: string, newRole: string) => {
    try {
      await api.team.update(userId, { role: newRole });
      loadTeam();
    } catch (err: any) {
      alert(err.message);
    }
  };

  if (loading) {
    return <DashboardLayout><p className="text-gray-500">Loading...</p></DashboardLayout>;
  }

  return (
    <DashboardLayout>
      <h2 className="text-2xl font-semibold text-gray-900 mb-6">Team</h2>

      {/* Invite Form */}
      {canInvite && (
        <div className="bg-white rounded-lg border p-6 mb-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Invite Team Member</h3>
          <form onSubmit={handleInvite} className="flex gap-3 items-end">
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
              <input
                type="email"
                value={email}
                onChange={e => setEmail(e.target.value)}
                required
                placeholder="colleague@company.com"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Role</label>
              <select
                value={role}
                onChange={e => setRole(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {invitableRoles.map(r => (
                  <option key={r} value={r}>{ROLE_LABELS[r]}</option>
                ))}
              </select>
            </div>
            <button
              type="submit"
              disabled={inviting}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
            >
              {inviting ? 'Inviting...' : 'Send Invite'}
            </button>
          </form>
          {inviteMsg && (
            <p className="mt-3 text-sm text-gray-600 bg-gray-50 p-3 rounded break-all">{inviteMsg}</p>
          )}
        </div>
      )}

      {/* Members Table */}
      <div className="bg-white rounded-lg border overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Role</th>
              {canManage && <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {data?.members.map(member => (
              <tr key={member.id}>
                <td className="px-6 py-4 text-sm text-gray-900">
                  {member.name}
                  {member.id === user?.id && <span className="ml-2 text-xs text-gray-400">(you)</span>}
                </td>
                <td className="px-6 py-4 text-sm text-gray-500">{member.email}</td>
                <td className="px-6 py-4 text-sm">
                  {canManage && member.id !== user?.id ? (
                    <select
                      value={member.role}
                      onChange={e => handleRoleChange(member.id, e.target.value)}
                      className="px-2 py-1 border border-gray-300 rounded text-sm"
                    >
                      {(user?.role === 'super_admin'
                        ? ['super_admin', 'admin', 'recruiter', 'hiring_manager', 'interviewer']
                        : ['recruiter', 'hiring_manager', 'interviewer']
                      ).map(r => (
                        <option key={r} value={r}>{ROLE_LABELS[r]}</option>
                      ))}
                    </select>
                  ) : (
                    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                      member.role === 'super_admin' ? 'bg-purple-100 text-purple-800' :
                      member.role === 'admin' ? 'bg-blue-100 text-blue-800' :
                      member.role === 'recruiter' ? 'bg-orange-100 text-orange-800' :
                      member.role === 'hiring_manager' ? 'bg-green-100 text-green-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {ROLE_LABELS[member.role] || member.role}
                    </span>
                  )}
                </td>
                {canManage && (
                  <td className="px-6 py-4 text-right">
                    {member.id !== user?.id && (
                      <button
                        onClick={() => handleRemove(member.id)}
                        className="text-red-600 hover:text-red-800 text-sm"
                      >
                        Remove
                      </button>
                    )}
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pending Invitations */}
      {canInvite && data?.invitations && data.invitations.filter(i => !i.accepted_at).length > 0 && (
        <div className="mt-6">
          <h3 className="text-lg font-medium text-gray-900 mb-3">Pending Invitations</h3>
          <div className="bg-white rounded-lg border overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Role</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Invite Link</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {data.invitations.filter(i => !i.accepted_at).map(inv => (
                  <tr key={inv.id}>
                    <td className="px-6 py-4 text-sm text-gray-900">{inv.email}</td>
                    <td className="px-6 py-4 text-sm text-gray-500">{ROLE_LABELS[inv.role] || inv.role}</td>
                    <td className="px-6 py-4 text-sm text-gray-500 truncate max-w-xs">
                      {window.location.origin}/accept-invite/{inv.token}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </DashboardLayout>
  );
}
