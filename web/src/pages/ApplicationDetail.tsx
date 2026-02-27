import { useState, useEffect, useRef, type FormEvent } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, type ApplicationDetail as AppDetailType, type CandidateApplication, type CandidateContact, type InterviewFeedback, type InterviewSchedule, type AvailableBlock, type InterviewAssignment, type TeamMember, type EmailNotification, type EmailTemplate, type Job } from '../api/client';
import { useAuth } from '../hooks/use-auth';
import DashboardLayout from '../components/layout/DashboardLayout';
import { STAGE_LABELS } from '../components/pipeline/KanbanBoard';

function pgText(val: unknown): string | null {
  if (val === null || val === undefined) return null;
  if (typeof val === 'string') return val || null;
  if (typeof val === 'object' && val !== null && 'Valid' in val) {
    const t = val as { String: string; Valid: boolean };
    return t.Valid ? t.String : null;
  }
  return null;
}

function pgTime(val: unknown): string | null {
  if (val === null || val === undefined) return null;
  if (typeof val === 'string') return val || null;
  if (typeof val === 'object' && val !== null && 'Time' in val) {
    const t = val as { Time: string; Valid: boolean };
    return t.Valid ? t.Time : null;
  }
  return null;
}

const stageColor = (stage: string) => {
  switch (stage) {
    case 'applied': return 'bg-blue-100 text-blue-800';
    case 'hr_screen': return 'bg-cyan-100 text-cyan-800';
    case 'hm_review': return 'bg-yellow-100 text-yellow-800';
    case 'first_interview': return 'bg-purple-100 text-purple-800';
    case 'final_interview': return 'bg-indigo-100 text-indigo-800';
    case 'offer': return 'bg-green-100 text-green-800';
    case 'rejected': return 'bg-red-100 text-red-800';
    default: return 'bg-gray-100 text-gray-800';
  }
};

const PIPELINE_STAGES = ['applied', 'hr_screen', 'hm_review', 'first_interview', 'final_interview', 'offer', 'rejected'];

const REJECTION_REASONS = [
  { value: 'not_qualified', label: 'Not Qualified' },
  { value: 'culture_fit', label: 'Culture Fit' },
  { value: 'salary_mismatch', label: 'Salary Mismatch' },
  { value: 'position_filled', label: 'Position Filled' },
  { value: 'withdrew', label: 'Withdrew' },
  { value: 'no_show', label: 'No Show' },
  { value: 'failed_assessment', label: 'Failed Assessment' },
  { value: 'other', label: 'Other' },
];

export default function ApplicationDetailPage() {
  const { appId } = useParams<{ appId: string }>();
  const { isAtLeast, user } = useAuth();
  const [data, setData] = useState<AppDetailType | null>(null);
  const [loading, setLoading] = useState(true);
  const [noteContent, setNoteContent] = useState('');
  const [noteLoading, setNoteLoading] = useState(false);
  const [stageLoading, setStageLoading] = useState(false);

  // Rejection dialog
  const [showRejectDialog, setShowRejectDialog] = useState(false);
  const [rejReason, setRejReason] = useState('');
  const [rejNotes, setRejNotes] = useState('');
  const [rejSendEmail, setRejSendEmail] = useState(true);
  const [rejEmailSubject, setRejEmailSubject] = useState('');
  const [rejEmailBody, setRejEmailBody] = useState('');
  const [showAllTimeline, setShowAllTimeline] = useState(false);
  const [otherApps, setOtherApps] = useState<CandidateApplication[]>([]);
  const [resumeUploading, setResumeUploading] = useState(false);
  const resumeInputRef = useRef<HTMLInputElement>(null);

  // Feedback
  const [showFeedbackForm, setShowFeedbackForm] = useState(false);
  const [feedbackStage, setFeedbackStage] = useState('');
  const [feedbackType, setFeedbackType] = useState('');
  const [feedbackContent, setFeedbackContent] = useState('');
  const [feedbackRec, setFeedbackRec] = useState('');
  const [feedbackLoading, setFeedbackLoading] = useState(false);
  const [editingFeedback, setEditingFeedback] = useState<InterviewFeedback | null>(null);
  const [editContent, setEditContent] = useState('');
  const [editRec, setEditRec] = useState('');
  const [editType, setEditType] = useState('');

  // @mention autocomplete
  const [teamMembers, setTeamMembers] = useState<TeamMember[]>([]);
  const [showMentions, setShowMentions] = useState(false);
  const [mentionFilter, setMentionFilter] = useState('');
  const noteRef = useRef<HTMLTextAreaElement>(null);

  // Compose email
  const [showEmailDialog, setShowEmailDialog] = useState(false);
  const [emailSubject, setEmailSubject] = useState('');
  const [emailBody, setEmailBody] = useState('');
  const [emailSending, setEmailSending] = useState(false);
  const [emailTemplates, setEmailTemplates] = useState<EmailTemplate[]>([]);

  // Notifications audit
  const [notifications, setNotifications] = useState<EmailNotification[]>([]);

  // Edit candidate contact
  const [editingContact, setEditingContact] = useState(false);
  const [editEmail, setEditEmail] = useState('');
  const [editPhone, setEditPhone] = useState('');
  const [editLinkedin, setEditLinkedin] = useState('');
  const [savingContact, setSavingContact] = useState(false);

  // Additional contacts
  const [contacts, setContacts] = useState<CandidateContact[]>([]);
  const [addingCategory, setAddingCategory] = useState<string | null>(null);
  const [newContactLabel, setNewContactLabel] = useState('');
  const [newContactValue, setNewContactValue] = useState('');
  const [contactSaving, setContactSaving] = useState(false);

  // Add to job dialog
  const [packetLoading, setPacketLoading] = useState(false);
  const [showAddToJob, setShowAddToJob] = useState(false);
  const [availableJobs, setAvailableJobs] = useState<Job[]>([]);
  const [selectedJobId, setSelectedJobId] = useState('');
  const [addingToJob, setAddingToJob] = useState(false);
  const [addJobError, setAddJobError] = useState('');

  // Scheduling
  const [schedules, setSchedules] = useState<InterviewSchedule[]>([]);
  const [showScheduleDialog, setShowScheduleDialog] = useState(false);
  const [schedInterviewers, setSchedInterviewers] = useState<InterviewAssignment[]>([]);
  const [selectedInterviewerIds, setSelectedInterviewerIds] = useState<string[]>([]);
  const [schedDateStart, setSchedDateStart] = useState('');
  const [schedDateEnd, setSchedDateEnd] = useState('');
  const [availableBlocks, setAvailableBlocks] = useState<AvailableBlock[]>([]);
  const [checkingAvail, setCheckingAvail] = useState(false);
  const [selectedSlots, setSelectedSlots] = useState<{ start: string; end: string }[]>([]);
  const [schedDuration, setSchedDuration] = useState(60);
  const [schedLocation, setSchedLocation] = useState('');
  const [schedMeetingUrl, setSchedMeetingUrl] = useState('');
  const [schedNotes, setSchedNotes] = useState('');
  const [schedSending, setSchedSending] = useState(false);
  const [schedStep, setSchedStep] = useState(1);

  const loadApp = () => {
    if (!appId) return;
    api.applications.get(appId)
      .then(setData)
      .finally(() => setLoading(false));
  };

  useEffect(() => { loadApp(); }, [appId]);

  // Load team members for @mention autocomplete
  useEffect(() => {
    api.team.list().then(d => setTeamMembers(d.members)).catch(() => {});
  }, []);

  // Load notifications audit trail
  useEffect(() => {
    if (appId) {
      api.applications.listNotifications(appId).then(setNotifications).catch(() => {});
    }
  }, [appId]);

  useEffect(() => {
    if (data?.application.candidate_id) {
      api.candidates.get(data.application.candidate_id)
        .then(res => {
          const sorted = [...res.applications].sort((a, b) => {
            const ta = pgTime(a.created_at) || '';
            const tb = pgTime(b.created_at) || '';
            return tb.localeCompare(ta);
          });
          setOtherApps(sorted);
        })
        .catch(() => {});
      api.candidates.listContacts(data.application.candidate_id)
        .then(setContacts)
        .catch(() => {});
    }
  }, [data?.application.candidate_id, appId]);

  const reloadContacts = () => {
    if (data?.application.candidate_id) {
      api.candidates.listContacts(data.application.candidate_id).then(setContacts).catch(() => {});
    }
  };

  const handleStageChange = async (newStage: string) => {
    if (!appId || !data) return;
    if (newStage === data.application.stage) return;

    if (newStage === 'rejected') {
      setRejSendEmail(true);
      setRejEmailSubject('');
      setRejEmailBody('');
      // Load templates and pre-fill rejection email
      api.settings.getEmailTemplates().then(templates => {
        setEmailTemplates(templates || []);
        const rejTemplate = (templates || []).find(t => t.stage === 'rejected' && t.enabled);
        if (rejTemplate) {
          const render = (s: string) => s
            .replace(/\{\{\.CandidateName\}\}/g, data.application.candidate_name)
            .replace(/\{\{\.JobTitle\}\}/g, data.application.job_title)
            .replace(/\{\{\.CompanyName\}\}/g, user?.orgName || '');
          setRejEmailSubject(render(rejTemplate.subject));
          setRejEmailBody(render(rejTemplate.body));
        } else {
          setRejEmailSubject(`Update on your application for ${data.application.job_title}`);
          setRejEmailBody(`Dear ${data.application.candidate_name},\n\nThank you for your interest in the ${data.application.job_title} position at ${user?.orgName || 'our company'}. After careful consideration, we have decided to move forward with other candidates at this time.\n\nWe appreciate your time and wish you the best in your job search.\n\nBest regards,\n${user?.orgName || 'The Team'}`);
        }
      }).catch(() => {});
      setShowRejectDialog(true);
      return;
    }

    setStageLoading(true);
    try {
      await api.applications.updateStage(appId, { stage: newStage });
      loadApp();
    } finally {
      setStageLoading(false);
    }
  };

  const handleReject = async () => {
    if (!appId || !rejReason) return;
    setStageLoading(true);
    try {
      await api.applications.updateStage(appId, {
        stage: 'rejected',
        rejectionReason: rejReason,
        rejectionNotes: rejNotes || undefined,
      });
      // Send rejection email if enabled
      if (rejSendEmail && rejEmailSubject.trim() && rejEmailBody.trim()) {
        try {
          await api.applications.sendEmail(appId, { subject: rejEmailSubject, body: rejEmailBody });
        } catch {
          // Don't block rejection if email fails
        }
      }
      setShowRejectDialog(false);
      setRejReason('');
      setRejNotes('');
      setRejEmailSubject('');
      setRejEmailBody('');
      loadApp();
    } finally {
      setStageLoading(false);
    }
  };

  const handleAddNote = async (e: FormEvent) => {
    e.preventDefault();
    if (!appId || !noteContent.trim()) return;
    setNoteLoading(true);
    try {
      await api.applications.addNote(appId, noteContent.trim());
      setNoteContent('');
      loadApp();
    } finally {
      setNoteLoading(false);
    }
  };

  const handleAddFeedback = async () => {
    if (!appId || !feedbackContent.trim() || !feedbackStage) return;
    setFeedbackLoading(true);
    try {
      await api.applications.addFeedback(appId, {
        stage: feedbackStage,
        interviewType: feedbackType || undefined,
        recommendation: feedbackRec,
        content: feedbackContent.trim(),
      });
      setFeedbackContent('');
      setFeedbackRec('none');
      setFeedbackStage('');
      setFeedbackType('');
      setShowFeedbackForm(false);
      loadApp();
    } finally {
      setFeedbackLoading(false);
    }
  };

  const handleUpdateFeedback = async () => {
    if (!appId || !editingFeedback || !editContent.trim()) return;
    setFeedbackLoading(true);
    try {
      await api.applications.updateFeedback(appId, editingFeedback.id, {
        interviewType: editType || undefined,
        recommendation: editRec,
        content: editContent.trim(),
      });
      setEditingFeedback(null);
      loadApp();
    } finally {
      setFeedbackLoading(false);
    }
  };

  const handleDeleteFeedback = async (feedbackId: string) => {
    if (!appId) return;
    setFeedbackLoading(true);
    try {
      await api.applications.deleteFeedback(appId, feedbackId);
      loadApp();
    } finally {
      setFeedbackLoading(false);
    }
  };

  // @mention handler for notes textarea
  const handleNoteChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const val = e.target.value;
    setNoteContent(val);
    const cursorPos = e.target.selectionStart;
    const textBefore = val.slice(0, cursorPos);
    const atMatch = textBefore.match(/@(\w*)$/);
    if (atMatch) {
      setMentionFilter(atMatch[1].toLowerCase());
      setShowMentions(true);
    } else {
      setShowMentions(false);
    }
  };

  const insertMention = (name: string) => {
    const textarea = noteRef.current;
    if (!textarea) return;
    const cursorPos = textarea.selectionStart;
    const textBefore = noteContent.slice(0, cursorPos);
    const textAfter = noteContent.slice(cursorPos);
    const atIdx = textBefore.lastIndexOf('@');
    const newText = textBefore.slice(0, atIdx) + '@' + name + ' ' + textAfter;
    setNoteContent(newText);
    setShowMentions(false);
    setTimeout(() => {
      const newPos = atIdx + name.length + 2;
      textarea.focus();
      textarea.setSelectionRange(newPos, newPos);
    }, 0);
  };

  const filteredMembers = teamMembers.filter(m =>
    m.name.toLowerCase().includes(mentionFilter) && m.id !== user?.id
  );

  // Send email to candidate
  const handleSendEmail = async () => {
    if (!appId || !emailSubject.trim() || !emailBody.trim()) return;
    setEmailSending(true);
    try {
      await api.applications.sendEmail(appId, { subject: emailSubject, body: emailBody });
      setShowEmailDialog(false);
      setEmailSubject('');
      setEmailBody('');
      // Refresh notifications
      api.applications.listNotifications(appId).then(setNotifications).catch(() => {});
      alert('Email sent successfully');
    } catch (err: any) {
      alert('Failed to send email: ' + err.message);
    } finally {
      setEmailSending(false);
    }
  };

  const startEditContact = () => {
    setEditEmail(app?.candidate_email || '');
    setEditPhone(pgText(app?.candidate_phone) || '');
    setEditLinkedin(pgText(app?.candidate_linkedin_url) || '');
    setEditingContact(true);
  };

  const handleSaveContact = async () => {
    if (!data || !editEmail.trim()) return;
    setSavingContact(true);
    try {
      await api.candidates.update(data.application.candidate_id, {
        email: editEmail.trim(),
        phone: editPhone.trim() || undefined,
        linkedinUrl: editLinkedin.trim() || undefined,
      });
      setEditingContact(false);
      loadApp();
    } catch (err: any) {
      alert('Failed to update: ' + err.message);
    } finally {
      setSavingContact(false);
    }
  };

  const handleResumeUpload = async (file: File) => {
    if (!data) return;
    setResumeUploading(true);
    try {
      await api.candidates.uploadResume(data.application.candidate_id, file);
      // Refresh application data to get new resume URL
      const updated = await api.applications.get(appId!);
      setData(updated);
    } catch (err: any) {
      alert(err.message || 'Failed to upload resume');
    } finally {
      setResumeUploading(false);
      if (resumeInputRef.current) resumeInputRef.current.value = '';
    }
  };

  const handleAddContact = async () => {
    if (!data || !newContactLabel || !newContactValue.trim() || !addingCategory) return;
    setContactSaving(true);
    try {
      await api.candidates.addContact(data.application.candidate_id, {
        category: addingCategory,
        label: newContactLabel,
        value: newContactValue.trim(),
      });
      setAddingCategory(null);
      setNewContactLabel('');
      setNewContactValue('');
      reloadContacts();
    } catch (err: any) {
      alert('Failed to add contact: ' + err.message);
    } finally {
      setContactSaving(false);
    }
  };

  const handleDeleteContact = async (contactId: string) => {
    if (!data) return;
    try {
      await api.candidates.deleteContact(data.application.candidate_id, contactId);
      reloadContacts();
    } catch (err: any) {
      alert('Failed to delete: ' + err.message);
    }
  };

  // Load schedules
  useEffect(() => {
    if (appId) {
      api.scheduling.list(appId).then(data => setSchedules(data || [])).catch(() => {});
    }
  }, [appId]);

  const openScheduleDialog = async () => {
    setShowScheduleDialog(true);
    setSchedStep(1);
    setSelectedInterviewerIds([]);
    setAvailableBlocks([]);
    setSelectedSlots([]);
    setSchedDuration(60);
    setSchedLocation('');
    setSchedMeetingUrl('');
    setSchedNotes('');
    setSchedDateStart('');
    setSchedDateEnd('');
    // Load interviewers for this application
    try {
      const interviewers = await api.applications.listInterviewers(appId!);
      setSchedInterviewers(interviewers);
      setSelectedInterviewerIds(interviewers.map(i => i.interviewer_id));
    } catch { setSchedInterviewers([]); }
  };

  const handleCheckAvailability = async () => {
    if (!appId || selectedInterviewerIds.length === 0 || !schedDateStart || !schedDateEnd) return;
    setCheckingAvail(true);
    try {
      const blocks = await api.scheduling.checkAvailability(appId, {
        interviewerIds: selectedInterviewerIds,
        startDate: new Date(schedDateStart + 'T09:00:00').toISOString(),
        endDate: new Date(schedDateEnd + 'T18:00:00').toISOString(),
      });
      setAvailableBlocks(blocks);
      setSchedStep(2);
    } catch { alert('Failed to check availability. Interviewers may not have connected their calendar.'); }
    finally { setCheckingAvail(false); }
  };

  const toggleSlot = (start: string, _end?: string) => {
    const startTime = new Date(start);
    const slotEnd = new Date(startTime.getTime() + schedDuration * 60000);
    const slot = { start: startTime.toISOString(), end: slotEnd.toISOString() };
    const exists = selectedSlots.find(s => s.start === slot.start);
    if (exists) {
      setSelectedSlots(selectedSlots.filter(s => s.start !== slot.start));
    } else {
      setSelectedSlots([...selectedSlots, slot]);
    }
  };

  const handleCreateSchedule = async () => {
    if (!appId || selectedSlots.length === 0) return;
    setSchedSending(true);
    try {
      await api.scheduling.create(appId, {
        slots: selectedSlots,
        durationMinutes: schedDuration,
        location: schedLocation || undefined,
        meetingUrl: schedMeetingUrl || undefined,
        notes: schedNotes || undefined,
        interviewerIds: selectedInterviewerIds,
      });
      setShowScheduleDialog(false);
      // Refresh schedules
      api.scheduling.list(appId).then(data => setSchedules(data || [])).catch(() => {});
      alert('Interview schedule sent to candidate!');
    } catch (err: any) {
      alert('Failed to create schedule: ' + err.message);
    } finally { setSchedSending(false); }
  };

  const openAddToJob = async () => {
    setShowAddToJob(true);
    setSelectedJobId('');
    setAddJobError('');
    try {
      const allJobs = await api.jobs.list();
      // Filter to open jobs, exclude jobs candidate is already in
      const existingJobIds = new Set([
        data?.application.job_id,
        ...otherApps.map(a => a.job_id),
      ]);
      setAvailableJobs(allJobs.filter(j => j.status === 'open' && !existingJobIds.has(j.id)));
    } catch {
      setAvailableJobs([]);
    }
  };

  const handleAddToJob = async () => {
    if (!data || !selectedJobId) return;
    setAddingToJob(true);
    setAddJobError('');
    try {
      await api.candidates.addToJob(data.application.candidate_id, selectedJobId);
      setShowAddToJob(false);
      // Refresh other apps list
      const res = await api.candidates.get(data.application.candidate_id);
      setOtherApps(res.applications.filter(a => a.id !== appId));
    } catch (err: any) {
      setAddJobError(err.message || 'Failed to add candidate to job');
    } finally {
      setAddingToJob(false);
    }
  };

  if (loading || !data) {
    return <DashboardLayout><p className="text-gray-500">Loading...</p></DashboardLayout>;
  }

  const { application: app, transitions, notes, feedback } = data;
  const company = pgText(app.candidate_company);
  const linkedin = pgText(app.candidate_linkedin_url);
  const phone = pgText(app.candidate_phone);
  const resumeUrl = pgText(app.candidate_resume_url);
  const canMoveStage = isAtLeast('hiring_manager');
  const canAddFeedback = isAtLeast('interviewer');
  const visibleTransitions = showAllTimeline ? transitions : transitions.slice(0, 3);
  const hasMoreTimeline = transitions.length > 3;
  const interviewStages = ['hr_screen', 'hm_review', 'first_interview', 'final_interview'];
  const feedbackByStage: Record<string, InterviewFeedback[]> = {};
  for (const fb of feedback || []) {
    if (!feedbackByStage[fb.stage]) feedbackByStage[fb.stage] = [];
    feedbackByStage[fb.stage].push(fb);
  }

  return (
    <DashboardLayout>
      <Link to={`/jobs/${app.job_id}/pipeline`} className="text-sm text-gray-500 hover:text-gray-700 mb-4 inline-block">
        &larr; Back to pipeline
      </Link>

      <div className="grid grid-cols-3 gap-6">
        {/* Left column: candidate info + notes */}
        <div className="col-span-2 space-y-6">
          {/* Candidate card */}
          <div className="bg-white rounded-lg border p-6">
            <div>
              <h2 className="text-2xl font-semibold text-gray-900">{app.candidate_name}</h2>
              {company && <p className="text-sm text-gray-600 mt-0.5">{company}</p>}
            </div>

            {/* Primary contact fields */}
            <div className="mt-4 text-sm">
              {editingContact ? (
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-gray-400 text-xs uppercase tracking-wide mb-1">Email *</p>
                    <input
                      type="email"
                      value={editEmail}
                      onChange={e => setEditEmail(e.target.value)}
                      className="w-full px-2 py-1 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <div>
                    <p className="text-gray-400 text-xs uppercase tracking-wide mb-1">Phone</p>
                    <input
                      type="tel"
                      value={editPhone}
                      onChange={e => setEditPhone(e.target.value)}
                      placeholder="Add phone..."
                      className="w-full px-2 py-1 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <div>
                    <p className="text-gray-400 text-xs uppercase tracking-wide mb-1">Online Presence</p>
                    <input
                      type="url"
                      value={editLinkedin}
                      onChange={e => setEditLinkedin(e.target.value)}
                      placeholder="Add URL..."
                      className="w-full px-2 py-1 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <div className="flex items-end gap-2">
                    <button
                      onClick={handleSaveContact}
                      disabled={savingContact || !editEmail.trim()}
                      className="px-3 py-1 text-sm text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50"
                    >
                      {savingContact ? 'Saving...' : 'Save'}
                    </button>
                    <button
                      onClick={() => setEditingContact(false)}
                      className="px-3 py-1 text-sm text-gray-700 bg-gray-100 rounded hover:bg-gray-200"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <div className="space-y-3">
                  {/* Email section */}
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <p className="text-gray-400 text-xs uppercase tracking-wide">Email</p>
                      {isAtLeast('recruiter') && (
                        <button onClick={startEditContact} className="text-xs text-blue-600 hover:text-blue-800">edit</button>
                      )}
                    </div>
                    <a href={`mailto:${app.candidate_email}`} className="text-blue-600 hover:underline">
                      {app.candidate_email}
                    </a>
                    <span className="ml-2 text-xs text-gray-400">Primary</span>
                    {contacts.filter(c => c.category === 'email').map(c => (
                      <div key={c.id} className="flex items-center gap-2 mt-1">
                        <a href={`mailto:${c.value}`} className="text-blue-600 hover:underline">{c.value}</a>
                        <span className="text-xs bg-gray-100 text-gray-600 px-1.5 py-0.5 rounded">{c.label}</span>
                        {isAtLeast('recruiter') && (
                          <button onClick={() => handleDeleteContact(c.id)} className="text-xs text-red-500 hover:text-red-700">&times;</button>
                        )}
                      </div>
                    ))}
                    {isAtLeast('recruiter') && addingCategory !== 'email' && (
                      <button
                        onClick={() => { setAddingCategory('email'); setNewContactLabel('Work'); setNewContactValue(''); }}
                        className="text-xs text-blue-600 hover:text-blue-800 mt-1"
                      >
                        + Add Email
                      </button>
                    )}
                    {addingCategory === 'email' && (
                      <div className="flex items-center gap-2 mt-1">
                        <select value={newContactLabel} onChange={e => setNewContactLabel(e.target.value)} className="text-xs border border-gray-300 rounded px-1 py-0.5">
                          <option value="Work">Work</option>
                          <option value="Personal">Personal</option>
                          <option value="Other">Other</option>
                        </select>
                        <input type="email" value={newContactValue} onChange={e => setNewContactValue(e.target.value)} placeholder="email@example.com" className="flex-1 px-2 py-0.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" />
                        <button onClick={handleAddContact} disabled={contactSaving || !newContactValue.trim()} className="text-xs text-white bg-blue-600 px-2 py-0.5 rounded hover:bg-blue-700 disabled:opacity-50">
                          {contactSaving ? '...' : 'Add'}
                        </button>
                        <button onClick={() => setAddingCategory(null)} className="text-xs text-gray-500 hover:text-gray-700">&times;</button>
                      </div>
                    )}
                  </div>

                  {/* Phone section */}
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <p className="text-gray-400 text-xs uppercase tracking-wide">Phone</p>
                    </div>
                    {phone ? (
                      <>
                        <a href={`tel:${phone}`} className="text-gray-900">{phone}</a>
                        <span className="ml-2 text-xs text-gray-400">Primary</span>
                      </>
                    ) : (
                      <span className="text-gray-400 italic">Not provided</span>
                    )}
                    {contacts.filter(c => c.category === 'phone').map(c => (
                      <div key={c.id} className="flex items-center gap-2 mt-1">
                        <a href={`tel:${c.value}`} className="text-gray-900">{c.value}</a>
                        <span className="text-xs bg-gray-100 text-gray-600 px-1.5 py-0.5 rounded">{c.label}</span>
                        {isAtLeast('recruiter') && (
                          <button onClick={() => handleDeleteContact(c.id)} className="text-xs text-red-500 hover:text-red-700">&times;</button>
                        )}
                      </div>
                    ))}
                    {isAtLeast('recruiter') && addingCategory !== 'phone' && (
                      <button
                        onClick={() => { setAddingCategory('phone'); setNewContactLabel('Mobile'); setNewContactValue(''); }}
                        className="text-xs text-blue-600 hover:text-blue-800 mt-1"
                      >
                        + Add Phone
                      </button>
                    )}
                    {addingCategory === 'phone' && (
                      <div className="flex items-center gap-2 mt-1">
                        <select value={newContactLabel} onChange={e => setNewContactLabel(e.target.value)} className="text-xs border border-gray-300 rounded px-1 py-0.5">
                          <option value="Mobile">Mobile</option>
                          <option value="Home">Home</option>
                          <option value="Work">Work</option>
                          <option value="Other">Other</option>
                        </select>
                        <input type="tel" value={newContactValue} onChange={e => setNewContactValue(e.target.value)} placeholder="+1 555-0100" className="flex-1 px-2 py-0.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" />
                        <button onClick={handleAddContact} disabled={contactSaving || !newContactValue.trim()} className="text-xs text-white bg-blue-600 px-2 py-0.5 rounded hover:bg-blue-700 disabled:opacity-50">
                          {contactSaving ? '...' : 'Add'}
                        </button>
                        <button onClick={() => setAddingCategory(null)} className="text-xs text-gray-500 hover:text-gray-700">&times;</button>
                      </div>
                    )}
                  </div>

                  {/* Online Presence section */}
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <p className="text-gray-400 text-xs uppercase tracking-wide">Online Presence</p>
                    </div>
                    {linkedin ? (
                      <div className="flex items-center gap-2">
                        <a href={linkedin} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline text-sm truncate max-w-xs">
                          {linkedin.replace(/^https?:\/\/(www\.)?/, '').replace(/\/$/, '')}
                        </a>
                        <span className="text-xs text-gray-400 flex-shrink-0">Primary</span>
                      </div>
                    ) : (
                      <span className="text-gray-400 italic text-sm">Not provided</span>
                    )}
                    {contacts.filter(c => c.category === 'online_presence').map(c => (
                      <div key={c.id} className="flex items-center gap-2 mt-1">
                        <a href={c.value.startsWith('http') ? c.value : `https://${c.value}`} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline truncate">
                          {c.value.replace(/^https?:\/\/(www\.)?/, '').replace(/\/$/, '')}
                        </a>
                        <span className="text-xs bg-gray-100 text-gray-600 px-1.5 py-0.5 rounded">{c.label}</span>
                        {isAtLeast('recruiter') && (
                          <button onClick={() => handleDeleteContact(c.id)} className="text-xs text-red-500 hover:text-red-700">&times;</button>
                        )}
                      </div>
                    ))}
                    {isAtLeast('recruiter') && addingCategory !== 'online_presence' && (
                      <button
                        onClick={() => { setAddingCategory('online_presence'); setNewContactLabel('LinkedIn'); setNewContactValue(''); }}
                        className="text-xs text-blue-600 hover:text-blue-800 mt-1"
                      >
                        + Add Online Presence
                      </button>
                    )}
                    {addingCategory === 'online_presence' && (
                      <div className="flex items-center gap-2 mt-1">
                        <select value={newContactLabel} onChange={e => setNewContactLabel(e.target.value)} className="text-xs border border-gray-300 rounded px-1 py-0.5">
                          <option value="LinkedIn">LinkedIn</option>
                          <option value="GitHub">GitHub</option>
                          <option value="Google Scholar">Google Scholar</option>
                          <option value="Portfolio">Portfolio</option>
                          <option value="Twitter/X">Twitter/X</option>
                          <option value="Other">Other</option>
                        </select>
                        <input type="url" value={newContactValue} onChange={e => setNewContactValue(e.target.value)} placeholder="https://..." className="flex-1 px-2 py-0.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" />
                        <button onClick={handleAddContact} disabled={contactSaving || !newContactValue.trim()} className="text-xs text-white bg-blue-600 px-2 py-0.5 rounded hover:bg-blue-700 disabled:opacity-50">
                          {contactSaving ? '...' : 'Add'}
                        </button>
                        <button onClick={() => setAddingCategory(null)} className="text-xs text-gray-500 hover:text-gray-700">&times;</button>
                      </div>
                    )}
                  </div>

                  {/* Resume */}
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <p className="text-gray-400 text-xs uppercase tracking-wide">Resume</p>
                      {isAtLeast('recruiter') && (
                        <>
                          <input
                            ref={resumeInputRef}
                            type="file"
                            accept=".pdf,.doc,.docx"
                            className="hidden"
                            onChange={e => { const f = e.target.files?.[0]; if (f) handleResumeUpload(f); }}
                          />
                          <button
                            onClick={() => resumeInputRef.current?.click()}
                            disabled={resumeUploading}
                            className="text-xs text-blue-600 hover:text-blue-800 font-medium disabled:opacity-50"
                          >
                            {resumeUploading ? 'Uploading...' : resumeUrl ? 'Replace' : '+ Add'}
                          </button>
                        </>
                      )}
                    </div>
                    {resumeUrl ? (
                      <a href={resumeUrl} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline text-sm">
                        View Resume
                      </a>
                    ) : (
                      <span className="text-gray-400 italic text-sm">No resume uploaded</span>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Rejection info */}
            {app.stage === 'rejected' && pgText(app.rejection_reason) && (
              <div className="mt-4 p-3 bg-red-50 rounded border border-red-200">
                <p className="text-sm font-medium text-red-800">Rejected: {pgText(app.rejection_reason)}</p>
                {pgText(app.rejection_notes) && (
                  <p className="text-sm text-red-600 mt-1">{pgText(app.rejection_notes)}</p>
                )}
              </div>
            )}

            {/* Email candidate button */}
            {isAtLeast('recruiter') && (
              <div className="mt-4 pt-4 border-t">
                <button
                  onClick={() => {
                    setEmailSubject(`Regarding your application for ${app.job_title}`);
                    setEmailBody('');
                    api.settings.getEmailTemplates().then(t => setEmailTemplates(t || [])).catch(() => setEmailTemplates([]));
                    setShowEmailDialog(true);
                  }}
                  className="px-4 py-2 bg-white border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 text-sm font-medium"
                >
                  Email Candidate
                </button>
              </div>
            )}
          </div>

          {/* Job info + stage + actions */}
          <div className="bg-white rounded-lg border p-4">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-xs text-gray-400 uppercase tracking-wide mb-1">Applied to</p>
                <Link to={`/jobs/${app.job_id}`} className="text-sm font-medium text-blue-600 hover:underline">
                  {app.job_title}
                </Link>
                <div className="flex gap-3 mt-1 text-xs text-gray-500">
                  {pgText(app.job_department) && <span>{pgText(app.job_department)}</span>}
                  {pgText(app.job_location) && <span>{pgText(app.job_location)}</span>}
                </div>
              </div>
              {canMoveStage ? (
                <select
                  value={app.stage}
                  onChange={e => handleStageChange(e.target.value)}
                  disabled={stageLoading}
                  className={`text-sm px-3 py-1.5 rounded-full font-medium border-0 cursor-pointer disabled:opacity-50 ${stageColor(app.stage)}`}
                >
                  {PIPELINE_STAGES.map(s => (
                    <option key={s} value={s}>{STAGE_LABELS[s] || s}</option>
                  ))}
                </select>
              ) : (
                <span className={`text-xs px-3 py-1 rounded-full font-medium ${stageColor(app.stage)}`}>
                  {STAGE_LABELS[app.stage] || app.stage}
                </span>
              )}
            </div>
            {isAtLeast('recruiter') && (
              <div className="mt-3 pt-3 border-t flex items-center gap-2">
                <button
                  onClick={openScheduleDialog}
                  className="flex items-center gap-1 px-3 py-1.5 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 text-sm font-medium"
                >
                  + Schedule Interview
                </button>
                {(app.stage === 'final_interview' || app.stage === 'offer') && (
                  <button
                    onClick={async () => {
                      setPacketLoading(true);
                      try {
                        await api.applications.generatePacket(appId!);
                        alert('Hiring packet sent to your email.');
                      } catch (err: any) {
                        alert(err.message || 'Failed to generate packet');
                      } finally {
                        setPacketLoading(false);
                      }
                    }}
                    disabled={packetLoading}
                    className="flex items-center gap-1 px-3 py-1.5 bg-amber-600 text-white rounded-md hover:bg-amber-700 text-sm font-medium disabled:opacity-50"
                  >
                    {packetLoading ? 'Generating...' : 'Hiring Packet'}
                  </button>
                )}
              </div>
            )}
          </div>

          {/* Interview Feedback */}
          <div className="bg-white rounded-lg border p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">Interview Feedback</h3>
              {canAddFeedback && !showFeedbackForm && (
                <button
                  onClick={() => { setShowFeedbackForm(true); setFeedbackStage(''); setFeedbackType(''); setFeedbackContent(''); setFeedbackRec('none'); }}
                  className="flex items-center gap-1 px-3 py-1.5 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm font-medium"
                >
                  + Add Feedback
                </button>
              )}
            </div>

            {/* Add feedback form */}
            {showFeedbackForm && (
              <div className="mb-5 p-4 rounded-lg border border-blue-200 bg-blue-50/50 space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">Stage *</label>
                    <select
                      value={feedbackStage}
                      onChange={e => { setFeedbackStage(e.target.value); setFeedbackType(''); }}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                    >
                      <option value="">Select stage...</option>
                      <option value="hr_screen">HR Screen</option>
                      <option value="hm_review">HM Review</option>
                      <option value="first_interview">Initial Interview</option>
                      <option value="final_interview">Final Interview</option>
                    </select>
                  </div>
                  {feedbackStage === 'final_interview' && (
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">Interview Type</label>
                      <select
                        value={feedbackType}
                        onChange={e => setFeedbackType(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                      >
                        <option value="">General</option>
                        <option value="manager">Manager</option>
                        <option value="cross_functional">Cross Functional</option>
                        <option value="technical">Technical</option>
                        <option value="culture">Culture Fit</option>
                        <option value="presentation">Presentation</option>
                      </select>
                    </div>
                  )}
                  {feedbackStage !== 'final_interview' && <div />}
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Feedback *</label>
                  <textarea
                    value={feedbackContent}
                    onChange={e => setFeedbackContent(e.target.value)}
                    rows={3}
                    placeholder="Write your feedback..."
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  />
                </div>
                <div className="flex items-center gap-3">
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">Recommendation</label>
                    <select value={feedbackRec} onChange={e => setFeedbackRec(e.target.value)} className="text-sm px-3 py-2 border border-gray-300 rounded-md">
                      <option value="">No score</option>
                      <option value="1">1 - Strong No</option>
                      <option value="2">2 - No</option>
                      <option value="3">3 - Yes</option>
                      <option value="4">4 - Strong Yes</option>
                    </select>
                  </div>
                  <div className="flex gap-2 pt-4">
                    <button
                      onClick={handleAddFeedback}
                      disabled={feedbackLoading || !feedbackContent.trim() || !feedbackStage}
                      className="px-4 py-2 bg-blue-600 text-white rounded-md text-sm hover:bg-blue-700 disabled:opacity-50 font-medium"
                    >
                      {feedbackLoading ? 'Submitting...' : 'Submit'}
                    </button>
                    <button
                      onClick={() => setShowFeedbackForm(false)}
                      className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md text-sm hover:bg-gray-200"
                    >Cancel</button>
                  </div>
                </div>
              </div>
            )}

            {/* Existing feedback grouped by stage */}
            <div className="space-y-5">
              {interviewStages.map(stage => {
                const stageFeedback = feedbackByStage[stage] || [];
                if (stageFeedback.length === 0) return null;
                return (
                  <div key={stage}>
                    <p className="text-sm font-medium text-gray-700 mb-2">{STAGE_LABELS[stage] || stage}</p>
                    <div className="space-y-2">
                      {stageFeedback.map(fb => {
                        const fbType = pgText(fb.interview_type);
                        return (
                          <div key={fb.id} className="p-3 rounded border border-gray-100 bg-gray-50">
                            {editingFeedback?.id === fb.id ? (
                              <div className="space-y-2">
                                <textarea
                                  value={editContent}
                                  onChange={e => setEditContent(e.target.value)}
                                  rows={3}
                                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                                />
                                <div className="flex items-center gap-2 flex-wrap">
                                  {fb.stage === 'final_interview' && (
                                    <select value={editType} onChange={e => setEditType(e.target.value)} className="text-sm px-2 py-1 border border-gray-300 rounded-md">
                                      <option value="">General</option>
                                      <option value="manager">Manager</option>
                                      <option value="cross_functional">Cross Functional</option>
                                      <option value="technical">Technical</option>
                                      <option value="culture">Culture Fit</option>
                                      <option value="presentation">Presentation</option>
                                    </select>
                                  )}
                                  <select value={editRec} onChange={e => setEditRec(e.target.value)} className="text-sm px-2 py-1 border border-gray-300 rounded-md">
                                    <option value="">No score</option>
                                    <option value="1">1 - Strong No</option>
                                    <option value="2">2 - No</option>
                                    <option value="3">3 - Yes</option>
                                    <option value="4">4 - Strong Yes</option>
                                  </select>
                                  <button onClick={handleUpdateFeedback} disabled={feedbackLoading || !editContent.trim()} className="px-3 py-1 bg-blue-600 text-white rounded text-xs hover:bg-blue-700 disabled:opacity-50">Save</button>
                                  <button onClick={() => setEditingFeedback(null)} className="px-3 py-1 bg-gray-100 text-gray-700 rounded text-xs hover:bg-gray-200">Cancel</button>
                                </div>
                              </div>
                            ) : (
                              <>
                                <div className="flex items-start justify-between gap-2">
                                  <p className="text-sm text-gray-800">{fb.content}</p>
                                  <div className="flex items-center gap-1.5 flex-shrink-0">
                                    {fbType && (
                                      <span className="text-xs px-2 py-0.5 rounded-full bg-gray-200 text-gray-600 capitalize">
                                        {fbType.replace('_', ' ')}
                                      </span>
                                    )}
                                    {fb.recommendation && fb.recommendation !== 'none' && (
                                      <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                                        fb.recommendation === '4' ? 'bg-green-100 text-green-700' :
                                        fb.recommendation === '3' ? 'bg-blue-100 text-blue-700' :
                                        fb.recommendation === '2' ? 'bg-amber-100 text-amber-700' :
                                        'bg-red-100 text-red-700'
                                      }`}>
                                        {fb.recommendation === '4' ? '4 - Strong Yes' : fb.recommendation === '3' ? '3 - Yes' : fb.recommendation === '2' ? '2 - No' : '1 - Strong No'}
                                      </span>
                                    )}
                                  </div>
                                </div>
                                <div className="flex items-center justify-between mt-1.5">
                                  <p className="text-xs text-gray-400">
                                    {fb.author_name} &middot; {pgTime(fb.created_at) ? new Date(pgTime(fb.created_at)!).toLocaleString() : ''}
                                  </p>
                                  {user && fb.author_id === user.id && (
                                    <div className="flex gap-2">
                                      <button
                                        onClick={() => { setEditingFeedback(fb); setEditContent(fb.content); setEditRec(fb.recommendation); setEditType(pgText(fb.interview_type) || ''); }}
                                        className="text-xs text-blue-600 hover:text-blue-800"
                                      >Edit</button>
                                      <button
                                        onClick={() => handleDeleteFeedback(fb.id)}
                                        className="text-xs text-red-500 hover:text-red-700"
                                      >Delete</button>
                                    </div>
                                  )}
                                </div>
                              </>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
              {!(feedback || []).length && (
                <p className="text-sm text-gray-400">No feedback yet.</p>
              )}
            </div>
          </div>

          {/* Notes */}
          <div className="bg-white rounded-lg border p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Notes</h3>
            <form onSubmit={handleAddNote} className="mb-4">
              <div className="relative">
                <textarea
                  ref={noteRef}
                  value={noteContent}
                  onChange={handleNoteChange}
                  rows={3}
                  placeholder="Add a note... Use @name to mention a team member"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                />
                {showMentions && filteredMembers.length > 0 && (
                  <div className="absolute z-10 left-0 right-0 mt-1 bg-white border border-gray-200 rounded-md shadow-lg max-h-40 overflow-y-auto">
                    {filteredMembers.map(m => (
                      <button
                        key={m.id}
                        type="button"
                        onClick={() => insertMention(m.name)}
                        className="w-full text-left px-3 py-2 text-sm hover:bg-blue-50 flex items-center gap-2"
                      >
                        <span className="font-medium text-gray-900">{m.name}</span>
                        <span className="text-gray-400 text-xs">{m.email}</span>
                      </button>
                    ))}
                  </div>
                )}
              </div>
              <button
                type="submit"
                disabled={!noteContent.trim() || noteLoading}
                className="mt-2 px-4 py-1.5 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm"
              >
                {noteLoading ? 'Adding...' : 'Add Note'}
              </button>
            </form>
            {notes.length === 0 ? (
              <p className="text-sm text-gray-400">No notes yet.</p>
            ) : (
              <div className="space-y-3">
                {notes.map(note => (
                  <div key={note.id} className="border-l-2 border-gray-200 pl-3">
                    <p className="text-sm text-gray-800">{note.content}</p>
                    <p className="text-xs text-gray-400 mt-1">
                      {note.author_name || 'Unknown'} &middot;{' '}
                      {pgTime(note.created_at) ? new Date(pgTime(note.created_at)!).toLocaleString() : ''}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Right column: timeline + application history */}
        <div className="space-y-6">
          <div className="bg-white rounded-lg border p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Timeline</h3>
            <div className="space-y-4">
              {visibleTransitions.map(t => (
                <div key={t.id} className="flex gap-3">
                  <div className="w-2 h-2 rounded-full bg-blue-400 mt-1.5 flex-shrink-0" />
                  <div>
                    <p className="text-sm text-gray-800">
                      {pgText(t.from_stage) ? (
                        <>{STAGE_LABELS[pgText(t.from_stage)!] || pgText(t.from_stage)} &rarr; <span className="font-medium">{STAGE_LABELS[t.to_stage] || t.to_stage}</span></>
                      ) : (
                        <span className="font-medium">Applied</span>
                      )}
                    </p>
                    <p className="text-xs text-gray-400">
                      {pgText(t.moved_by_name) && `${pgText(t.moved_by_name)} · `}
                      {pgTime(t.created_at) ? new Date(pgTime(t.created_at)!).toLocaleString() : ''}
                    </p>
                  </div>
                </div>
              ))}
            </div>
            {hasMoreTimeline && (
              <button
                onClick={() => setShowAllTimeline(!showAllTimeline)}
                className="mt-3 text-xs text-blue-600 hover:text-blue-800 font-medium"
              >
                {showAllTimeline ? 'Show less' : `Show all (${transitions.length})`}
              </button>
            )}
          </div>

          {/* Application History */}
          <div className="bg-white rounded-lg border p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">Application History</h3>
              {isAtLeast('recruiter') && (
                <button
                  onClick={openAddToJob}
                  title="Add to another job pipeline"
                  className="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium text-gray-600 bg-gray-100 rounded-md hover:bg-blue-50 hover:text-blue-600 transition-colors"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-3.5 h-3.5">
                    <path d="M10.75 4.75a.75.75 0 00-1.5 0v4.5h-4.5a.75.75 0 000 1.5h4.5v4.5a.75.75 0 001.5 0v-4.5h4.5a.75.75 0 000-1.5h-4.5v-4.5z" />
                  </svg>
                  Add to Job
                </button>
              )}
            </div>
            {otherApps.length === 0 ? (
              <p className="text-sm text-gray-400">No applications.</p>
            ) : (
              <div className="space-y-3">
                {otherApps.map(a => {
                  const appliedAt = pgTime(a.created_at);
                  const isCurrent = a.id === appId;
                  return (
                    <Link
                      key={a.id}
                      to={`/applications/${a.id}`}
                      className={`block p-3 rounded border transition-colors ${isCurrent ? 'border-blue-300 bg-blue-50/50' : 'hover:bg-gray-50'}`}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <p className={`text-sm font-medium ${isCurrent ? 'text-blue-700' : 'text-gray-900'}`}>{a.job_title}</p>
                          {isCurrent && <span className="text-xs text-blue-500 font-medium">Current</span>}
                        </div>
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${stageColor(a.stage)}`}>
                          {STAGE_LABELS[a.stage] || a.stage}
                        </span>
                      </div>
                      {appliedAt && (
                        <p className="text-xs text-gray-400 mt-1">
                          {new Date(appliedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                        </p>
                      )}
                    </Link>
                  );
                })}
              </div>
            )}
          </div>

          {/* Scheduled Interviews */}
          <div className="bg-white rounded-lg border p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">Interviews</h3>
              {isAtLeast('recruiter') && (
                <button
                  onClick={openScheduleDialog}
                  className="flex items-center gap-1 px-3 py-1.5 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 text-sm font-medium"
                >
                  + Schedule
                </button>
              )}
            </div>
            {schedules.length === 0 ? (
              <p className="text-sm text-gray-400">No interviews scheduled.</p>
            ) : (
              <div className="space-y-3">
                {schedules.map(sch => {
                  const selSlot = sch.slots?.find(s => s.selected);
                  return (
                    <div key={sch.id} className="p-3 rounded border border-gray-100 bg-gray-50">
                      <div className="flex items-center justify-between">
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                          sch.status === 'confirmed' ? 'bg-green-100 text-green-700' :
                          sch.status === 'cancelled' ? 'bg-red-100 text-red-700' :
                          'bg-yellow-100 text-yellow-700'
                        }`}>
                          {sch.status}
                        </span>
                        <span className="text-xs text-gray-400">{sch.duration_minutes} min</span>
                      </div>
                      {selSlot && (
                        <p className="text-sm font-medium text-gray-800 mt-1">
                          {new Date(selSlot.start_time.Time).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}
                          {' '}at{' '}
                          {new Date(selSlot.start_time.Time).toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })}
                        </p>
                      )}
                      {!selSlot && sch.status === 'pending' && (
                        <p className="text-xs text-gray-500 mt-1">{sch.slots?.length || 0} time slots proposed</p>
                      )}
                      {sch.interviewers && sch.interviewers.length > 0 && (
                        <p className="text-xs text-gray-500 mt-1">
                          {sch.interviewers.map(i => i.name).join(', ')}
                        </p>
                      )}
                      <p className="text-xs text-gray-400 mt-1">
                        by {sch.created_by_name} &middot; {pgTime(sch.created_at) ? new Date(pgTime(sch.created_at)!).toLocaleDateString() : ''}
                      </p>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* Emails Sent */}
          {notifications.length > 0 && (
            <div className="bg-white rounded-lg border p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Emails</h3>
              <div className="space-y-3">
                {notifications.map(n => (
                  <div key={n.id} className="border-l-2 border-gray-200 pl-3">
                    <div className="flex items-center gap-2">
                      <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                        n.type === 'freeform' ? 'bg-blue-100 text-blue-700' :
                        n.type === 'stage_change' ? 'bg-purple-100 text-purple-700' :
                        'bg-amber-100 text-amber-700'
                      }`}>
                        {n.type === 'freeform' ? 'Email' : n.type === 'stage_change' ? 'Stage' : 'Mention'}
                      </span>
                      <span className={`text-xs ${n.status === 'sent' ? 'text-green-600' : 'text-red-600'}`}>
                        {n.status}
                      </span>
                    </div>
                    <p className="text-sm text-gray-800 mt-1">{n.subject}</p>
                    <p className="text-xs text-gray-400 mt-0.5">
                      To: {n.recipient_email} &middot; {pgTime(n.created_at) ? new Date(pgTime(n.created_at)!).toLocaleString() : ''}
                    </p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Add to Job Dialog */}
      {showAddToJob && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900 mb-1">Add to Job Pipeline</h3>
            <p className="text-sm text-gray-500 mb-4">
              Add <strong>{app.candidate_name}</strong> to another job as an applicant.
            </p>
            {addJobError && (
              <div className="mb-3 p-2 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{addJobError}</div>
            )}
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Job *</label>
                {availableJobs.length === 0 ? (
                  <p className="text-sm text-gray-500">No available open jobs to add this candidate to.</p>
                ) : (
                  <select
                    value={selectedJobId}
                    onChange={e => setSelectedJobId(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  >
                    <option value="">Select a job...</option>
                    {availableJobs.map(j => (
                      <option key={j.id} value={j.id}>{j.title}{j.department ? ` (${j.department})` : ''}</option>
                    ))}
                  </select>
                )}
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => setShowAddToJob(false)}
                  className="px-4 py-2 text-sm text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
                >
                  Cancel
                </button>
                <button
                  onClick={handleAddToJob}
                  disabled={!selectedJobId || addingToJob}
                  className="px-4 py-2 text-sm text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  {addingToJob ? 'Adding...' : 'Add to Pipeline'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Rejection Dialog */}
      {/* Compose Email Dialog */}
      {showEmailDialog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-lg shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900 mb-1">Email Candidate</h3>
            <p className="text-sm text-gray-500 mb-4">Send an email to {app.candidate_name} ({app.candidate_email})</p>
            <div className="space-y-4">
              {emailTemplates.filter(t => t.enabled).length > 0 && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Use Template</label>
                  <select
                    defaultValue=""
                    onChange={e => {
                      const tmpl = emailTemplates.find(t => t.id === e.target.value);
                      if (tmpl) {
                        const render = (s: string) => s
                          .replace(/\{\{\.CandidateName\}\}/g, app.candidate_name)
                          .replace(/\{\{\.JobTitle\}\}/g, app.job_title)
                          .replace(/\{\{\.CompanyName\}\}/g, user?.orgName || '');
                        setEmailSubject(render(tmpl.subject));
                        setEmailBody(render(tmpl.body));
                      }
                    }}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  >
                    <option value="">-- Select a template --</option>
                    {emailTemplates.filter(t => t.enabled).map(t => (
                      <option key={t.id} value={t.id}>{STAGE_LABELS[t.stage] || t.stage}</option>
                    ))}
                  </select>
                </div>
              )}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Subject</label>
                <input
                  type="text"
                  value={emailSubject}
                  onChange={e => setEmailSubject(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Body (HTML)</label>
                <textarea
                  value={emailBody}
                  onChange={e => setEmailBody(e.target.value)}
                  rows={8}
                  placeholder="Write your email content..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                />
              </div>
              <div className="flex gap-3 pt-2">
                <button
                  onClick={handleSendEmail}
                  disabled={emailSending || !emailSubject.trim() || !emailBody.trim()}
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
                >
                  {emailSending ? 'Sending...' : 'Send Email'}
                </button>
                <button
                  onClick={() => setShowEmailDialog(false)}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 text-sm"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Schedule Interview Dialog */}
      {showScheduleDialog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-2xl shadow-xl max-h-[85vh] overflow-y-auto">
            <h3 className="text-lg font-semibold text-gray-900 mb-1">Schedule Interview</h3>
            <p className="text-sm text-gray-500 mb-4">
              Schedule an interview for {app.candidate_name} - {app.job_title}
            </p>

            {schedStep === 1 && (
              <div className="space-y-4">
                {/* Interviewers */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Interviewers</label>
                  {schedInterviewers.length === 0 ? (
                    <p className="text-sm text-gray-500">No interviewers assigned. Assign interviewers first.</p>
                  ) : (
                    <div className="space-y-1">
                      {schedInterviewers.map(iv => (
                        <label key={iv.interviewer_id} className="flex items-center gap-2 text-sm">
                          <input
                            type="checkbox"
                            checked={selectedInterviewerIds.includes(iv.interviewer_id)}
                            onChange={e => {
                              if (e.target.checked) setSelectedInterviewerIds([...selectedInterviewerIds, iv.interviewer_id]);
                              else setSelectedInterviewerIds(selectedInterviewerIds.filter(id => id !== iv.interviewer_id));
                            }}
                          />
                          {iv.interviewer_name} <span className="text-gray-400">({iv.interviewer_email})</span>
                        </label>
                      ))}
                    </div>
                  )}
                </div>

                {/* Date range */}
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">From Date</label>
                    <input type="date" value={schedDateStart} onChange={e => setSchedDateStart(e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">To Date</label>
                    <input type="date" value={schedDateEnd} onChange={e => setSchedDateEnd(e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm" />
                  </div>
                </div>

                {/* Duration */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Duration</label>
                  <select value={schedDuration} onChange={e => setSchedDuration(Number(e.target.value))} className="px-3 py-2 border border-gray-300 rounded-md text-sm">
                    <option value={30}>30 minutes</option>
                    <option value={45}>45 minutes</option>
                    <option value={60}>60 minutes</option>
                    <option value={90}>90 minutes</option>
                  </select>
                </div>

                <div className="flex justify-end gap-2 pt-2">
                  <button onClick={() => setShowScheduleDialog(false)} className="px-4 py-2 text-sm text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200">Cancel</button>
                  <button
                    onClick={handleCheckAvailability}
                    disabled={checkingAvail || selectedInterviewerIds.length === 0 || !schedDateStart || !schedDateEnd}
                    className="px-4 py-2 text-sm text-white bg-indigo-600 rounded-md hover:bg-indigo-700 disabled:opacity-50"
                  >
                    {checkingAvail ? 'Checking...' : 'Check Availability'}
                  </button>
                </div>
              </div>
            )}

            {schedStep === 2 && (
              <div className="space-y-4">
                <p className="text-sm text-gray-600">
                  Select time slots to propose to the candidate. Click available blocks to add them.
                </p>

                {availableBlocks.length === 0 ? (
                  <div className="p-4 bg-yellow-50 rounded text-sm text-yellow-800">
                    No available blocks found. Try a different date range, or interviewers may not have connected their calendar.
                    <button onClick={() => setSchedStep(1)} className="block mt-2 text-indigo-600 hover:text-indigo-800 font-medium">Go back</button>
                  </div>
                ) : (
                  <>
                    <div className="max-h-60 overflow-y-auto border rounded p-2 space-y-1">
                      {availableBlocks.map((block, i) => {
                        const bStart = new Date(block.start);
                        const bEnd = new Date(block.end);
                        // Generate slot options within this block
                        const slots: Date[] = [];
                        let cursor = new Date(bStart);
                        while (cursor.getTime() + schedDuration * 60000 <= bEnd.getTime()) {
                          slots.push(new Date(cursor));
                          cursor = new Date(cursor.getTime() + 30 * 60000); // 30min increments
                        }
                        return (
                          <div key={i}>
                            <p className="text-xs font-semibold text-gray-500 mt-2 mb-1">
                              {bStart.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}
                            </p>
                            <div className="flex flex-wrap gap-1">
                              {slots.map(s => {
                                const iso = s.toISOString();
                                const isSelected = selectedSlots.some(sl => sl.start === iso);
                                return (
                                  <button
                                    key={iso}
                                    onClick={() => toggleSlot(iso, '')}
                                    className={`px-2 py-1 text-xs rounded border ${
                                      isSelected
                                        ? 'bg-indigo-600 text-white border-indigo-600'
                                        : 'bg-white text-gray-700 border-gray-300 hover:border-indigo-400'
                                    }`}
                                  >
                                    {s.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })}
                                  </button>
                                );
                              })}
                            </div>
                          </div>
                        );
                      })}
                    </div>

                    <p className="text-xs text-gray-500">{selectedSlots.length} slot(s) selected</p>

                    {/* Optional fields */}
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-xs font-medium text-gray-700 mb-1">Location</label>
                        <input type="text" value={schedLocation} onChange={e => setSchedLocation(e.target.value)} placeholder="e.g. Room 201" className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm" />
                      </div>
                      <div>
                        <label className="block text-xs font-medium text-gray-700 mb-1">Meeting URL</label>
                        <input type="url" value={schedMeetingUrl} onChange={e => setSchedMeetingUrl(e.target.value)} placeholder="https://zoom.us/..." className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm" />
                      </div>
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">Notes for candidate</label>
                      <textarea value={schedNotes} onChange={e => setSchedNotes(e.target.value)} rows={2} placeholder="Any instructions for the candidate..." className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm" />
                    </div>

                    <div className="flex justify-end gap-2 pt-2">
                      <button onClick={() => setSchedStep(1)} className="px-4 py-2 text-sm text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200">Back</button>
                      <button onClick={() => setShowScheduleDialog(false)} className="px-4 py-2 text-sm text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200">Cancel</button>
                      <button
                        onClick={handleCreateSchedule}
                        disabled={schedSending || selectedSlots.length === 0}
                        className="px-4 py-2 text-sm text-white bg-indigo-600 rounded-md hover:bg-indigo-700 disabled:opacity-50"
                      >
                        {schedSending ? 'Sending...' : `Send ${selectedSlots.length} Slot(s) to Candidate`}
                      </button>
                    </div>
                  </>
                )}
              </div>
            )}
          </div>
        </div>
      )}

      {showRejectDialog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-lg shadow-xl max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-semibold text-gray-900 mb-1">Reject Candidate</h3>
            <p className="text-sm text-gray-500 mb-4">Rejecting {app.candidate_name}</p>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Reason *</label>
                <select
                  value={rejReason}
                  onChange={e => setRejReason(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Select reason...</option>
                  {REJECTION_REASONS.map(r => (
                    <option key={r.value} value={r.value}>{r.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Notes</label>
                <textarea
                  value={rejNotes}
                  onChange={e => setRejNotes(e.target.value)}
                  rows={2}
                  placeholder="Internal notes (not sent to candidate)..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              {/* Rejection email */}
              <div className="border-t pt-4">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={rejSendEmail}
                    onChange={e => setRejSendEmail(e.target.checked)}
                    className="rounded border-gray-300 text-red-600 focus:ring-red-500"
                  />
                  <span className="text-sm font-medium text-gray-700">Send rejection email to candidate</span>
                </label>
                {rejSendEmail && (
                  <div className="mt-3 space-y-3">
                    {emailTemplates.filter(t => t.enabled).length > 0 && (
                      <div>
                        <label className="block text-xs font-medium text-gray-500 mb-1">Template</label>
                        <select
                          defaultValue=""
                          onChange={e => {
                            const tmpl = emailTemplates.find(t => t.id === e.target.value);
                            if (tmpl) {
                              const render = (s: string) => s
                                .replace(/\{\{\.CandidateName\}\}/g, app.candidate_name)
                                .replace(/\{\{\.JobTitle\}\}/g, app.job_title)
                                .replace(/\{\{\.CompanyName\}\}/g, user?.orgName || '');
                              setRejEmailSubject(render(tmpl.subject));
                              setRejEmailBody(render(tmpl.body));
                            }
                          }}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                        >
                          <option value="">-- Change template --</option>
                          {emailTemplates.filter(t => t.enabled).map(t => (
                            <option key={t.id} value={t.id}>{STAGE_LABELS[t.stage] || t.stage}</option>
                          ))}
                        </select>
                      </div>
                    )}
                    <div>
                      <label className="block text-xs font-medium text-gray-500 mb-1">Subject</label>
                      <input
                        type="text"
                        value={rejEmailSubject}
                        onChange={e => setRejEmailSubject(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-500 mb-1">Body</label>
                      <textarea
                        value={rejEmailBody}
                        onChange={e => setRejEmailBody(e.target.value)}
                        rows={6}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                      />
                    </div>
                  </div>
                )}
              </div>

              <div className="flex gap-3 pt-2">
                <button
                  onClick={handleReject}
                  disabled={!rejReason || stageLoading}
                  className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50 text-sm font-medium"
                >
                  {stageLoading ? 'Rejecting...' : rejSendEmail ? 'Reject & Send Email' : 'Reject'}
                </button>
                <button
                  onClick={() => { setShowRejectDialog(false); setRejReason(''); setRejNotes(''); }}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 text-sm"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </DashboardLayout>
  );
}
