import { useState, useEffect, type FormEvent } from 'react';
import { api, type EmailTemplate } from '../api/client';
import DashboardLayout from '../components/layout/DashboardLayout';

const STAGE_LABELS: Record<string, string> = {
  applied: 'Applied',
  hr_screen: 'HR Screen',
  hm_review: 'HM Review',
  first_interview: 'Initial Interview',
  final_interview: 'Final Interview',
  offer: 'Offer',
  rejected: 'Rejected',
};

const TEMPLATE_STAGES = ['applied', 'hr_screen', 'hm_review', 'first_interview', 'final_interview', 'offer', 'rejected'];

export default function Settings() {
  const [companyBlurb, setCompanyBlurb] = useState('');
  const [closingStatement, setClosingStatement] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState('');
  const [error, setError] = useState('');

  // SMTP state
  const [smtpHost, setSmtpHost] = useState('');
  const [smtpPort, setSmtpPort] = useState(587);
  const [smtpUsername, setSmtpUsername] = useState('');
  const [smtpPassword, setSmtpPassword] = useState('');
  const [smtpFromEmail, setSmtpFromEmail] = useState('');
  const [smtpFromName, setSmtpFromName] = useState('');
  const [smtpTls, setSmtpTls] = useState(true);
  const [smtpConfigured, setSmtpConfigured] = useState(false);
  const [smtpSaving, setSmtpSaving] = useState(false);
  const [smtpTesting, setSmtpTesting] = useState(false);

  // Email templates state
  const [templates, setTemplates] = useState<EmailTemplate[]>([]);
  const [editingTemplate, setEditingTemplate] = useState<string | null>(null);
  const [tmplSubject, setTmplSubject] = useState('');
  const [tmplBody, setTmplBody] = useState('');
  const [tmplEnabled, setTmplEnabled] = useState(false);
  const [tmplSaving, setTmplSaving] = useState(false);

  useEffect(() => {
    Promise.all([
      api.settings.getDefaults(),
      api.settings.getSmtp().catch(() => null),
      api.settings.getEmailTemplates().catch(() => []),
    ]).then(([defaults, smtp, tmpls]) => {
      setCompanyBlurb(defaults.defaultCompanyBlurb);
      setClosingStatement(defaults.defaultClosingStatement);
      if (smtp && smtp.configured) {
        setSmtpConfigured(true);
        setSmtpHost(smtp.host);
        setSmtpPort(smtp.port);
        setSmtpUsername(smtp.username);
        setSmtpPassword(smtp.password);
        setSmtpFromEmail(smtp.fromEmail);
        setSmtpFromName(smtp.fromName || '');
        setSmtpTls(smtp.tls);
      }
      setTemplates(tmpls);
    }).catch(() => {
      setError('Failed to load settings');
    }).finally(() => setLoading(false));
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(''); setSaved(''); setSaving(true);
    try {
      await api.settings.updateDefaults({ defaultCompanyBlurb: companyBlurb, defaultClosingStatement: closingStatement });
      setSaved('defaults');
      setTimeout(() => setSaved(''), 3000);
    } catch (err: any) { setError(err.message); }
    finally { setSaving(false); }
  };

  const handleSmtpSave = async () => {
    setError(''); setSaved(''); setSmtpSaving(true);
    try {
      await api.settings.updateSmtp({
        host: smtpHost, port: smtpPort, username: smtpUsername, password: smtpPassword,
        fromEmail: smtpFromEmail, fromName: smtpFromName, tls: smtpTls,
      });
      setSmtpConfigured(true);
      setSaved('smtp');
      setTimeout(() => setSaved(''), 3000);
    } catch (err: any) { setError(err.message); }
    finally { setSmtpSaving(false); }
  };

  const handleSmtpTest = async () => {
    setError(''); setSaved(''); setSmtpTesting(true);
    try {
      const res = await api.settings.testSmtp();
      setSaved('smtp-test');
      setError('');
      alert(`Test email sent to ${res.sentTo}`);
    } catch (err: any) { setError('SMTP test failed: ' + err.message); }
    finally { setSmtpTesting(false); }
  };

  const openTemplateEditor = (stage: string) => {
    const existing = templates.find(t => t.stage === stage);
    setEditingTemplate(stage);
    setTmplSubject(existing?.subject || '');
    setTmplBody(existing?.body || '');
    setTmplEnabled(existing?.enabled || false);
  };

  const handleTemplateSave = async () => {
    if (!editingTemplate) return;
    setError(''); setTmplSaving(true);
    try {
      const updated = await api.settings.updateEmailTemplate({
        stage: editingTemplate, subject: tmplSubject, body: tmplBody, enabled: tmplEnabled,
      });
      setTemplates(prev => {
        const filtered = prev.filter(t => t.stage !== editingTemplate);
        return [...filtered, updated].sort((a, b) => a.stage.localeCompare(b.stage));
      });
      setEditingTemplate(null);
      setSaved('template');
      setTimeout(() => setSaved(''), 3000);
    } catch (err: any) { setError(err.message); }
    finally { setTmplSaving(false); }
  };

  if (loading) {
    return <DashboardLayout><p className="text-gray-500">Loading...</p></DashboardLayout>;
  }

  return (
    <DashboardLayout>
      <div className="max-w-2xl space-y-6">
        <h2 className="text-2xl font-semibold text-gray-900">Settings</h2>

        {error && (
          <div className="p-3 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{error}</div>
        )}

        {/* Job Description Defaults */}
        <div className="bg-white p-6 rounded-lg border">
          <h3 className="text-lg font-medium text-gray-900 mb-1">Job Description Defaults</h3>
          <p className="text-sm text-gray-500 mb-4">
            These values will prepopulate the Company Blurb and Closing Statement when creating new job descriptions.
          </p>

          {saved === 'defaults' && (
            <div className="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded text-sm">Defaults saved successfully.</div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Default Company Blurb</label>
              <textarea
                value={companyBlurb}
                onChange={e => setCompanyBlurb(e.target.value)}
                rows={5}
                placeholder="About your company — this will prepopulate for all new job descriptions."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Default Closing Statement</label>
              <textarea
                value={closingStatement}
                onChange={e => setClosingStatement(e.target.value)}
                rows={4}
                placeholder="Standard closing statement — e.g., equal opportunity employer notice."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <button
              type="submit"
              disabled={saving}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
            >
              {saving ? 'Saving...' : 'Save Defaults'}
            </button>
          </form>
        </div>

        {/* SMTP Configuration */}
        <div className="bg-white p-6 rounded-lg border">
          <div className="flex items-center justify-between mb-1">
            <h3 className="text-lg font-medium text-gray-900">Email (SMTP)</h3>
            {smtpConfigured && (
              <span className="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-700 font-medium">Configured</span>
            )}
          </div>
          <p className="text-sm text-gray-500 mb-4">
            Configure your SMTP server to enable email notifications and candidate communication.
          </p>

          {saved === 'smtp' && (
            <div className="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded text-sm">SMTP settings saved.</div>
          )}

          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">SMTP Host</label>
                <input type="text" value={smtpHost} onChange={e => setSmtpHost(e.target.value)}
                  placeholder="smtp.gmail.com" className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Port</label>
                <input type="number" value={smtpPort} onChange={e => setSmtpPort(Number(e.target.value))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm" />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Username</label>
                <input type="text" value={smtpUsername} onChange={e => setSmtpUsername(e.target.value)}
                  placeholder="your@email.com" className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Password</label>
                <input type="password" value={smtpPassword} onChange={e => setSmtpPassword(e.target.value)}
                  placeholder="App password or SMTP password" className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm" />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">From Email</label>
                <input type="email" value={smtpFromEmail} onChange={e => setSmtpFromEmail(e.target.value)}
                  placeholder="noreply@yourcompany.com" className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">From Name</label>
                <input type="text" value={smtpFromName} onChange={e => setSmtpFromName(e.target.value)}
                  placeholder="Your Company" className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm" />
              </div>
            </div>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input type="checkbox" checked={smtpTls} onChange={e => setSmtpTls(e.target.checked)} className="rounded" />
              Use TLS (STARTTLS)
            </label>
            <div className="flex gap-2 pt-2">
              <button onClick={handleSmtpSave} disabled={smtpSaving || !smtpHost || !smtpUsername || !smtpPassword || !smtpFromEmail}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm font-medium">
                {smtpSaving ? 'Saving...' : 'Save SMTP Settings'}
              </button>
              {smtpConfigured && (
                <button onClick={handleSmtpTest} disabled={smtpTesting}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 disabled:opacity-50 text-sm font-medium">
                  {smtpTesting ? 'Sending...' : 'Send Test Email'}
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Stage Notification Templates */}
        <div className="bg-white p-6 rounded-lg border">
          <h3 className="text-lg font-medium text-gray-900 mb-1">Stage Notification Templates</h3>
          <p className="text-sm text-gray-500 mb-4">
            Configure automatic emails sent to candidates when they move to specific stages.
            Use template variables: <code className="bg-gray-100 px-1 rounded text-xs">{'{{.CandidateName}}'}</code>, <code className="bg-gray-100 px-1 rounded text-xs">{'{{.JobTitle}}'}</code>, <code className="bg-gray-100 px-1 rounded text-xs">{'{{.CompanyName}}'}</code>
          </p>

          {saved === 'template' && (
            <div className="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded text-sm">Template saved.</div>
          )}

          <div className="space-y-2">
            {TEMPLATE_STAGES.map(stage => {
              const tmpl = templates.find(t => t.stage === stage);
              const isEditing = editingTemplate === stage;
              return (
                <div key={stage} className="border rounded-md">
                  <div
                    className="flex items-center justify-between px-4 py-3 cursor-pointer hover:bg-gray-50"
                    onClick={() => isEditing ? setEditingTemplate(null) : openTemplateEditor(stage)}
                  >
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-gray-900">{STAGE_LABELS[stage] || stage}</span>
                      {tmpl?.enabled ? (
                        <span className="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-700">Active</span>
                      ) : (
                        <span className="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-500">Inactive</span>
                      )}
                    </div>
                    <span className="text-gray-400 text-sm">{isEditing ? '−' : '+'}</span>
                  </div>

                  {isEditing && (
                    <div className="px-4 pb-4 space-y-3 border-t">
                      <div className="pt-3">
                        <label className="block text-sm font-medium text-gray-700 mb-1">Subject</label>
                        <input type="text" value={tmplSubject} onChange={e => setTmplSubject(e.target.value)}
                          placeholder="e.g., Update on your application for {{.JobTitle}}"
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm" />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Body (HTML)</label>
                        <textarea value={tmplBody} onChange={e => setTmplBody(e.target.value)} rows={6}
                          placeholder="<p>Hi {{.CandidateName}},</p><p>Thank you for your application...</p>"
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm font-mono" />
                      </div>
                      <label className="flex items-center gap-2 text-sm text-gray-700">
                        <input type="checkbox" checked={tmplEnabled} onChange={e => setTmplEnabled(e.target.checked)} className="rounded" />
                        Enable this template
                      </label>
                      <div className="flex gap-2">
                        <button onClick={handleTemplateSave} disabled={tmplSaving || !tmplSubject || !tmplBody}
                          className="px-3 py-1.5 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm font-medium">
                          {tmplSaving ? 'Saving...' : 'Save Template'}
                        </button>
                        <button onClick={() => setEditingTemplate(null)}
                          className="px-3 py-1.5 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 text-sm">Cancel</button>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}
