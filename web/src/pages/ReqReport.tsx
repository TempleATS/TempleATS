import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell, Legend,
} from 'recharts';
import { api, type ReqReport as ReqReportType } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

const STAGE_ORDER = ['applied', 'screening', 'interview', 'offer', 'hired', 'rejected'];
const STAGE_COLORS: Record<string, string> = {
  applied: '#60a5fa',
  screening: '#fbbf24',
  interview: '#a78bfa',
  offer: '#34d399',
  hired: '#059669',
  rejected: '#f87171',
};

const PIE_COLORS = ['#60a5fa', '#f87171', '#fbbf24', '#a78bfa', '#34d399', '#fb923c', '#6b7280', '#e879f9'];

const REASON_LABELS: Record<string, string> = {
  not_qualified: 'Not Qualified',
  culture_fit: 'Culture Fit',
  salary_mismatch: 'Salary Mismatch',
  position_filled: 'Position Filled',
  withdrew: 'Withdrew',
  no_show: 'No Show',
  failed_assessment: 'Failed Assessment',
  other: 'Other',
};

export default function ReqReport() {
  const { reqId } = useParams<{ reqId: string }>();
  const [data, setData] = useState<ReqReportType | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!reqId) return;
    api.reqs.report(reqId).then(setData).finally(() => setLoading(false));
  }, [reqId]);

  if (loading || !data) {
    return <DashboardLayout><p className="text-gray-500">Loading report...</p></DashboardLayout>;
  }

  const { requisition: req, funnel, rejections, byJob, fillProgress, timeToHire } = data;

  const funnelData = STAGE_ORDER.map(stage => ({
    stage: stage.charAt(0).toUpperCase() + stage.slice(1),
    count: funnel[stage] || 0,
    fill: STAGE_COLORS[stage],
  }));

  const rejData = Object.entries(rejections.byReason).map(([reason, count]) => ({
    name: REASON_LABELS[reason] || reason,
    value: count,
  }));

  return (
    <DashboardLayout>
      <Link to={`/reqs/${reqId}`} className="text-blue-600 hover:underline text-sm mb-4 inline-block">
        &larr; Back to requisition
      </Link>
      <h2 className="text-2xl font-semibold text-gray-900 mb-1">{req.title} - Report</h2>
      <p className="text-sm text-gray-500 mb-6">
        {req.level ? `${req.level} · ` : ''}{req.department || 'No department'} · Status: {req.status}
      </p>

      {/* Fill Progress */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        <div className="bg-white rounded-lg border p-4">
          <p className="text-sm text-gray-500">Target Hires</p>
          <p className="text-3xl font-bold text-gray-900">{fillProgress.target}</p>
        </div>
        <div className="bg-white rounded-lg border p-4">
          <p className="text-sm text-gray-500">Hired</p>
          <p className="text-3xl font-bold text-emerald-600">{fillProgress.hired}</p>
        </div>
        <div className="bg-white rounded-lg border p-4">
          <p className="text-sm text-gray-500">Total Applicants</p>
          <p className="text-3xl font-bold text-gray-900">
            {Object.values(funnel).reduce((a, b) => a + b, 0)}
          </p>
        </div>
        <div className="bg-white rounded-lg border p-4">
          <p className="text-sm text-gray-500">Rejected</p>
          <p className="text-3xl font-bold text-red-500">{funnel.rejected || 0}</p>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-6 mb-8">
        {/* Funnel Chart */}
        <div className="bg-white rounded-lg border p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Pipeline Funnel</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={funnelData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="stage" tick={{ fontSize: 12 }} />
              <YAxis allowDecimals={false} />
              <Tooltip />
              <Bar dataKey="count" name="Candidates">
                {funnelData.map((entry, i) => (
                  <Cell key={i} fill={entry.fill} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Rejection Breakdown */}
        <div className="bg-white rounded-lg border p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Rejection Reasons</h3>
          {rejData.length === 0 ? (
            <p className="text-sm text-gray-400 mt-12 text-center">No rejections yet.</p>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie data={rejData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={100} label>
                  {rejData.map((_, i) => (
                    <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>

      {/* Avg Time in Stage */}
      {Object.keys(timeToHire.avgDaysInStage).length > 0 && (
        <div className="bg-white rounded-lg border p-6 mb-8">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Average Time in Stage (days)</h3>
          <div className="grid grid-cols-5 gap-4">
            {STAGE_ORDER.filter(s => s !== 'rejected' && s !== 'hired').map(stage => (
              <div key={stage} className="text-center">
                <p className="text-2xl font-bold text-gray-900">
                  {timeToHire.avgDaysInStage[stage] !== undefined
                    ? timeToHire.avgDaysInStage[stage].toFixed(1)
                    : '-'}
                </p>
                <p className="text-xs text-gray-500 capitalize">{stage}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Per-Job Breakdown */}
      {byJob.length > 0 && (
        <div className="bg-white rounded-lg border p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">By Job</h3>
          <table className="w-full">
            <thead className="bg-gray-50 border-b">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Job</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Applied</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Hired</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Rejected</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {byJob.map(j => (
                <tr key={j.job_id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-900">{j.job_title}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">{j.total}</td>
                  <td className="px-4 py-3 text-sm text-emerald-600 font-medium">{j.hired}</td>
                  <td className="px-4 py-3 text-sm text-red-500">{j.rejected}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </DashboardLayout>
  );
}
