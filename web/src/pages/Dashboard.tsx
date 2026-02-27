import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api, type DashboardMetricsData, type PersonStats } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

function pct(ratio: number | null | undefined): string {
  if (ratio == null) return '--';
  return `${(ratio * 100).toFixed(1)}%`;
}

function PersonTable({ title, rows, roleLabel }: { title: string; rows: PersonStats[]; roleLabel: string }) {
  if (rows.length === 0) {
    return (
      <div className="bg-white rounded-lg border p-6 mb-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-2">{title}</h3>
        <p className="text-sm text-gray-400">No {roleLabel.toLowerCase()}s assigned to requisitions yet.</p>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg border overflow-hidden mb-6">
      <div className="px-6 py-4 border-b">
        <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50 border-b">
            <tr>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">{roleLabel}</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Reqs</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Open</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Apps</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Hired</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Rejected</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">1st→Final</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Final→Offer</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Offer→Hired</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {rows.map(row => (
              <tr key={row.user_id} className="hover:bg-gray-50">
                <td className="px-4 py-3 font-medium text-gray-900">{row.user_name}</td>
                <td className="px-4 py-3 text-sm text-gray-600">{row.total_reqs}</td>
                <td className="px-4 py-3 text-sm text-gray-600">{row.open_reqs}</td>
                <td className="px-4 py-3 text-sm text-gray-600">{row.total_applications}</td>
                <td className="px-4 py-3 text-sm text-emerald-600 font-medium">{row.total_hired}</td>
                <td className="px-4 py-3 text-sm text-red-500">{row.total_rejected}</td>
                <td className="px-4 py-3 text-sm text-purple-600 font-medium">{pct(row.ratio_first_to_final)}</td>
                <td className="px-4 py-3 text-sm text-emerald-600 font-medium">{pct(row.ratio_final_to_offer)}</td>
                <td className="px-4 py-3 text-sm text-blue-600 font-medium">{pct(row.ratio_offer_to_hired)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default function Dashboard() {
  const [data, setData] = useState<DashboardMetricsData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.metrics.dashboard().then(setData).catch(() => {}).finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <DashboardLayout><p className="text-gray-500">Loading dashboard...</p></DashboardLayout>;
  }

  if (!data) {
    return (
      <DashboardLayout>
        <h2 className="text-2xl font-semibold text-gray-900 mb-6">Dashboard</h2>
        <p className="text-gray-500">No metrics data available yet. Create requisitions and move candidates through the pipeline to see stats.</p>
      </DashboardLayout>
    );
  }

  const { org_conversions: oc } = data;

  return (
    <DashboardLayout>
      <h2 className="text-2xl font-semibold text-gray-900 mb-6">Dashboard</h2>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-8">
        <Link to="/reqs" className="bg-white p-5 rounded-lg border hover:shadow-md transition-shadow">
          <p className="text-sm text-gray-500">Open Reqs</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{data.open_reqs}</p>
        </Link>
        <div className="bg-white p-5 rounded-lg border">
          <p className="text-sm text-gray-500">Total Reqs</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{data.total_reqs}</p>
        </div>
        <Link to="/candidates" className="bg-white p-5 rounded-lg border hover:shadow-md transition-shadow">
          <p className="text-sm text-gray-500">Total Applications</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{data.total_applications}</p>
        </Link>
        <div className="bg-white p-5 rounded-lg border">
          <p className="text-sm text-gray-500">Hired</p>
          <p className="text-3xl font-bold text-emerald-600 mt-1">{data.total_hired}</p>
        </div>
        <div className="bg-white p-5 rounded-lg border">
          <p className="text-sm text-gray-500">Rejected</p>
          <p className="text-3xl font-bold text-red-500 mt-1">{data.total_rejected}</p>
        </div>
      </div>

      {/* Org-Wide Conversion Ratios */}
      <div className="mb-8">
        <h3 className="text-lg font-semibold text-gray-900 mb-3">Org-Wide Conversion Ratios</h3>
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-white rounded-lg border p-5 text-center">
            <p className="text-sm text-gray-500 mb-1">1st Interview → Final</p>
            <p className="text-3xl font-bold text-purple-600">{pct(oc.ratio_first_to_final)}</p>
            <p className="text-xs text-gray-400 mt-1">{oc.tp_final_interview || 0} of {oc.tp_first_interview || 0} candidates</p>
          </div>
          <div className="bg-white rounded-lg border p-5 text-center">
            <p className="text-sm text-gray-500 mb-1">Final → Offer</p>
            <p className="text-3xl font-bold text-emerald-600">{pct(oc.ratio_final_to_offer)}</p>
            <p className="text-xs text-gray-400 mt-1">{oc.tp_offer || 0} of {oc.tp_final_interview || 0} candidates</p>
          </div>
          <div className="bg-white rounded-lg border p-5 text-center">
            <p className="text-sm text-gray-500 mb-1">Offer → Hired</p>
            <p className="text-3xl font-bold text-blue-600">{pct(oc.ratio_offer_to_hired)}</p>
            <p className="text-xs text-gray-400 mt-1">{oc.tp_hired || 0} of {oc.tp_offer || 0} candidates</p>
          </div>
        </div>
      </div>

      {/* Recruiter Performance */}
      <PersonTable title="Recruiter Performance" rows={data.recruiter_stats} roleLabel="Recruiter" />

      {/* Hiring Manager Performance */}
      <PersonTable title="Hiring Manager Performance" rows={data.hm_stats} roleLabel="Hiring Manager" />
    </DashboardLayout>
  );
}
