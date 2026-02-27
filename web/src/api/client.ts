const API_BASE = '/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body.error || 'Request failed');
  }

  return res.json();
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  orgId: string;
  orgSlug: string;
  orgName: string;
}

export interface Requisition {
  id: string;
  title: string;
  job_code: string | null;
  level: string | null;
  department: string | null;
  target_hires: number;
  status: string;
  hiring_manager_id: string | null;
  recruiter_id: string | null;
  organization_id: string;
  opened_at: string;
  closed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Job {
  id: string;
  title: string;
  company_blurb: string;
  team_details: string;
  responsibilities: string;
  qualifications: string;
  closing_statement: string;
  location: string | null;
  department: string | null;
  salary: string | null;
  status: string;
  requisition_id: string | null;
  organization_id: string;
  created_at: string;
  updated_at: string;
}

export interface CareerOrg {
  name: string;
  slug: string;
  logoUrl: string | null;
  website: string | null;
}

export interface CareerJob {
  id: string;
  title: string;
  company_blurb: string;
  team_details: string;
  responsibilities: string;
  qualifications: string;
  closing_statement: string;
  location: { String: string; Valid: boolean } | null;
  department: { String: string; Valid: boolean } | null;
  salary: { String: string; Valid: boolean } | null;
  status: string;
  created_at: { Time: string; Valid: boolean };
}

export interface OrgDefaults {
  defaultCompanyBlurb: string;
  defaultClosingStatement: string;
}

export interface ApplyResponse {
  applicationId: string;
  candidateId: string;
  message: string;
}

export interface PipelineApplication {
  id: string;
  stage: string;
  rejection_reason: { String: string; Valid: boolean } | null;
  rejection_notes: { String: string; Valid: boolean } | null;
  candidate_id: string;
  job_id: string;
  created_at: { Time: string; Valid: boolean };
  updated_at: { Time: string; Valid: boolean };
  candidate_name: string;
  candidate_email: string;
  candidate_resume_url: { String: string; Valid: boolean } | null;
  candidate_company: { String: string; Valid: boolean } | null;
}

export interface PipelineData {
  job: Job;
  stages: Record<string, PipelineApplication[]>;
}

export type PgText = { String: string; Valid: boolean } | null;
export type PgTimestamp = { Time: string; Valid: boolean };

export interface Candidate {
  id: string;
  name: string;
  email: string;
  phone: PgText;
  resume_url: PgText;
  resume_filename: PgText;
  company: PgText;
  linkedin_url: PgText;
  organization_id: string;
  created_at: PgTimestamp;
  updated_at: PgTimestamp;
}

export interface CandidateContact {
  id: string;
  candidate_id: string;
  category: string;
  label: string;
  value: string;
  created_at: PgTimestamp;
}

export interface CandidateListItem {
  id: string;
  name: string;
  email: string;
  phone: PgText;
  resume_url: PgText;
  resume_filename: PgText;
  created_at: PgTimestamp;
  app_id: string;
  app_stage: string;
  job_title: string;
  applied_at: PgTimestamp;
}

export interface CandidateApplication {
  id: string;
  stage: string;
  rejection_reason: PgText;
  rejection_notes: PgText;
  candidate_id: string;
  job_id: string;
  created_at: PgTimestamp;
  updated_at: PgTimestamp;
  job_title: string;
  job_status: string;
}

export interface Note {
  id: string;
  content: string;
  application_id: string;
  author_id: string;
  author_name?: string;
  created_at: PgTimestamp;
}

export interface InterviewFeedback {
  id: string;
  application_id: string;
  stage: string;
  interview_type: PgText;
  recommendation: string;
  content: string;
  author_id: string;
  author_name: string;
  created_at: PgTimestamp;
  updated_at: PgTimestamp;
}

export interface ApplicationDetail {
  application: {
    id: string;
    stage: string;
    rejection_reason: { String: string; Valid: boolean } | null;
    rejection_notes: { String: string; Valid: boolean } | null;
    candidate_id: string;
    job_id: string;
    created_at: { Time: string; Valid: boolean };
    updated_at: { Time: string; Valid: boolean };
    candidate_name: string;
    candidate_email: string;
    candidate_phone: { String: string; Valid: boolean } | null;
    candidate_resume_url: { String: string; Valid: boolean } | null;
    candidate_company: { String: string; Valid: boolean } | null;
    candidate_linkedin_url: { String: string; Valid: boolean } | null;
    job_title: string;
    job_location: { String: string; Valid: boolean } | null;
    job_department: { String: string; Valid: boolean } | null;
    org_name: string;
  };
  transitions: {
    id: string;
    application_id: string;
    from_stage: { String: string; Valid: boolean } | null;
    to_stage: string;
    moved_by_name: { String: string; Valid: boolean } | null;
    created_at: { Time: string; Valid: boolean };
  }[];
  notes: Note[];
  feedback: InterviewFeedback[];
}

export interface TeamMember {
  id: string;
  email: string;
  name: string;
  role: string;
  organization_id: string;
  created_at: PgTimestamp;
}

export interface Invitation {
  id: string;
  email: string;
  role: string;
  token: string;
  organization_id: string;
  expires_at: PgTimestamp;
  accepted_at: PgTimestamp | null;
  created_at: PgTimestamp;
}

export interface TeamData {
  members: TeamMember[];
  invitations: Invitation[];
}

export interface InterviewAssignment {
  id: string;
  application_id: string;
  interviewer_id: string;
  interviewer_name: string;
  interviewer_email: string;
  created_at: PgTimestamp;
}

export interface MyInterview {
  id: string;
  stage: string;
  candidate_name: string;
  candidate_email: string;
  candidate_resume_url: PgText;
  job_title: string;
  job_id: string;
  created_at: PgTimestamp;
}

export interface SmtpSettings {
  configured: boolean;
  host: string;
  port: number;
  username: string;
  password: string;
  fromEmail: string;
  fromName: string;
  tls: boolean;
}

export interface EmailTemplate {
  id: string;
  organization_id: string;
  stage: string;
  subject: string;
  body: string;
  enabled: boolean;
  created_at: PgTimestamp;
  updated_at: PgTimestamp;
}

export interface EmailNotification {
  id: string;
  organization_id: string;
  type: string;
  recipient_email: string;
  recipient_name: string;
  subject: string;
  body: string;
  status: string;
  error_message: PgText;
  application_id: PgText;
  note_id: PgText;
  triggered_by_id: PgText;
  created_at: PgTimestamp;
}

export interface ReqReport {
  requisition: Requisition;
  funnel: Record<string, number>;
  timeToHire: {
    avgDaysInStage: Record<string, number>;
  };
  rejections: {
    total: number;
    byReason: Record<string, number>;
  };
  byJob: { job_id: string; job_title: string; total: number; hired: number; rejected: number }[];
  fillProgress: { hired: number; target: number };
}

export interface MetricSnapshot {
  id: string;
  requisition_id: string;
  organization_id: string;
  created_by_id: string;
  label: string | null;
  funnel_applied: number;
  funnel_hr_screen: number;
  funnel_hm_review: number;
  funnel_first_interview: number;
  funnel_final_interview: number;
  funnel_offer: number;
  funnel_hired: number;
  funnel_rejected: number;
  tp_applied: number;
  tp_first_interview: number;
  tp_final_interview: number;
  tp_offer: number;
  tp_hired: number;
  ratio_first_to_final: number | null;
  ratio_final_to_offer: number | null;
  ratio_offer_to_hired: number | null;
  total_applications: number;
  target_hires: number;
  created_at: string;
}

export interface PersonStats {
  user_id: string;
  user_name: string;
  total_reqs: number;
  open_reqs: number;
  total_applications: number;
  total_hired: number;
  total_rejected: number;
  tp_first_interview: number;
  tp_final_interview: number;
  tp_offer: number;
  tp_hired: number;
  ratio_first_to_final: number | null;
  ratio_final_to_offer: number | null;
  ratio_offer_to_hired: number | null;
}

export interface DashboardMetricsData {
  total_reqs: number;
  open_reqs: number;
  total_applications: number;
  total_hired: number;
  total_rejected: number;
  org_conversions: {
    tp_first_interview: number;
    tp_final_interview: number;
    tp_offer: number;
    tp_hired: number;
    tp_applied: number;
    ratio_first_to_final?: number;
    ratio_final_to_offer?: number;
    ratio_offer_to_hired?: number;
  };
  recruiter_stats: PersonStats[];
  hm_stats: PersonStats[];
}

export interface CalendarConnection {
  id: string;
  provider: string;
  calendar_email: string;
  created_at: PgTimestamp;
}

export interface InterviewSlot {
  id: string;
  schedule_id: string;
  start_time: PgTimestamp;
  end_time: PgTimestamp;
  selected: boolean;
}

export interface ScheduleInterviewer {
  schedule_id: string;
  user_id: string;
  calendar_event_id: PgText;
  name: string;
  email: string;
}

export interface InterviewSchedule {
  id: string;
  application_id: string;
  token: string;
  status: string;
  duration_minutes: number;
  location: PgText;
  meeting_url: PgText;
  notes: PgText;
  created_by: string;
  confirmed_at: PgTimestamp | null;
  created_at: PgTimestamp;
  created_by_name: string;
  slots: InterviewSlot[];
  interviewers: ScheduleInterviewer[];
}

export interface AvailableBlock {
  start: string;
  end: string;
}

export interface PublicSchedule {
  id: string;
  status: string;
  duration_minutes: number;
  location: PgText;
  meeting_url: PgText;
  notes: PgText;
  created_at: PgTimestamp;
  confirmed_at: PgTimestamp | null;
  job_title: string;
  candidate_name: string;
  org_name: string;
  slots: InterviewSlot[];
}

export interface Referral {
  id: string;
  source: string;
  referrer_name: string;
  candidate_name: PgText;
  job_title: string;
  job_id: string;
  application_stage: string;
  application_id: PgText;
  token: string;
  created_at: PgTimestamp;
}

export const api = {
  auth: {
    signup: (data: { email: string; name: string; password: string; orgName: string; orgSlug: string }) =>
      request<User>('/auth/signup', { method: 'POST', body: JSON.stringify(data) }),
    login: (data: { email: string; password: string }) =>
      request<User>('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
    logout: () =>
      request<{ status: string }>('/auth/logout', { method: 'POST' }),
    me: () =>
      request<User>('/auth/me'),
    acceptInvite: (data: { token: string; name: string; password: string }) =>
      request<User>('/auth/accept-invite', { method: 'POST', body: JSON.stringify(data) }),
    ssoEnabled: () => request<{ enabled: boolean }>('/auth/sso/enabled'),
    ssoUrl: () => request<{ url: string }>('/auth/sso/url'),
  },
  reqs: {
    list: (q?: string) => request<Requisition[]>(q ? `/reqs?q=${encodeURIComponent(q)}` : '/reqs'),
    create: (data: { title: string; jobCode?: string; level?: string; department?: string; targetHires?: number; hiringManagerId?: string; recruiterId?: string }) =>
      request<Requisition>('/reqs', { method: 'POST', body: JSON.stringify(data) }),
    get: (id: string) => request<{ requisition: Requisition; jobs: Job[] }>(`/reqs/${id}`),
    update: (id: string, data: { title: string; jobCode?: string; level?: string; department?: string; targetHires?: number; status: string; hiringManagerId?: string; recruiterId?: string }) =>
      request<Requisition>(`/reqs/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request<{ status: string }>(`/reqs/${id}`, { method: 'DELETE' }),
    attachJob: (reqId: string, jobId: string) =>
      request<Job>(`/reqs/${reqId}/jobs`, { method: 'POST', body: JSON.stringify({ jobId }) }),
    report: (id: string) => request<ReqReport>(`/reqs/${id}/report`),
    createSnapshot: (reqId: string, label?: string) =>
      request<MetricSnapshot>(`/reqs/${reqId}/snapshots`, { method: 'POST', body: JSON.stringify({ label: label || '' }) }),
    listSnapshots: (reqId: string) =>
      request<MetricSnapshot[]>(`/reqs/${reqId}/snapshots`),
    deleteSnapshot: (reqId: string, snapId: string) =>
      request<{ status: string }>(`/reqs/${reqId}/snapshots/${snapId}`, { method: 'DELETE' }),
  },
  metrics: {
    dashboard: () => request<DashboardMetricsData>('/metrics/dashboard'),
  },
  interviews: {
    mine: () => request<MyInterview[]>('/my-interviews'),
  },
  jobs: {
    list: (q?: string) => request<Job[]>(q ? `/jobs?q=${encodeURIComponent(q)}` : '/jobs'),
    create: (data: { title: string; companyBlurb: string; teamDetails: string; responsibilities: string; qualifications: string; closingStatement: string; location?: string; department?: string; salary?: string; status?: string; requisitionId?: string }) =>
      request<Job>('/jobs', { method: 'POST', body: JSON.stringify(data) }),
    get: (id: string) => request<Job>(`/jobs/${id}`),
    update: (id: string, data: { title: string; companyBlurb: string; teamDetails: string; responsibilities: string; qualifications: string; closingStatement: string; location?: string; department?: string; salary?: string; status: string; requisitionId?: string }) =>
      request<Job>(`/jobs/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    pipeline: (id: string) => request<PipelineData>(`/jobs/${id}/pipeline`),
  },
  applications: {
    get: (id: string) => request<ApplicationDetail>(`/applications/${id}`),
    updateStage: (id: string, data: { stage: string; rejectionReason?: string; rejectionNotes?: string }) =>
      request<PipelineApplication>(`/applications/${id}/stage`, { method: 'PUT', body: JSON.stringify(data) }),
    addNote: (id: string, content: string) =>
      request<Note>(`/applications/${id}/notes`, { method: 'POST', body: JSON.stringify({ content }) }),
    listInterviewers: (id: string) =>
      request<InterviewAssignment[]>(`/applications/${id}/interviewers`),
    assignInterviewer: (appId: string, userId: string) =>
      request<InterviewAssignment>(`/applications/${appId}/interviewers`, { method: 'POST', body: JSON.stringify({ userId }) }),
    removeInterviewer: (appId: string, userId: string) =>
      request<{ status: string }>(`/applications/${appId}/interviewers/${userId}`, { method: 'DELETE' }),
    addFeedback: (appId: string, data: { stage: string; interviewType?: string; recommendation: string; content: string }) =>
      request<InterviewFeedback>(`/applications/${appId}/feedback`, { method: 'POST', body: JSON.stringify(data) }),
    updateFeedback: (appId: string, feedbackId: string, data: { interviewType?: string; recommendation: string; content: string }) =>
      request<InterviewFeedback>(`/applications/${appId}/feedback/${feedbackId}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteFeedback: (appId: string, feedbackId: string) =>
      request<{ status: string }>(`/applications/${appId}/feedback/${feedbackId}`, { method: 'DELETE' }),
    sendEmail: (appId: string, data: { subject: string; body: string }) =>
      request<{ status: string }>(`/applications/${appId}/email`, { method: 'POST', body: JSON.stringify(data) }),
    listNotifications: (appId: string) =>
      request<EmailNotification[]>(`/applications/${appId}/notifications`),
    generatePacket: (appId: string) =>
      request<{ status: string }>(`/applications/${appId}/hiring-packet`, { method: 'POST' }),
  },
  candidates: {
    create: async (data: { name: string; email: string; phone?: string; jobId?: string; resume?: File }): Promise<{ candidateId: string; applicationId?: string }> => {
      const form = new FormData();
      form.append('name', data.name);
      form.append('email', data.email);
      if (data.phone) form.append('phone', data.phone);
      if (data.jobId) form.append('jobId', data.jobId);
      if (data.resume) form.append('resume', data.resume);
      const res = await fetch('/api/candidates', { method: 'POST', body: form, credentials: 'include' });
      if (!res.ok) { const e = await res.json().catch(() => ({})); throw new Error(e.error || 'Failed to create candidate'); }
      return res.json();
    },
    list: (q?: string) => request<CandidateListItem[]>(`/candidates${q ? `?q=${encodeURIComponent(q)}` : ''}`),
    get: (id: string) => request<{ candidate: Candidate; applications: CandidateApplication[] }>(`/candidates/${id}`),
    update: (candidateId: string, data: { email: string; phone?: string; linkedinUrl?: string }) =>
      request<Candidate>(`/candidates/${candidateId}`, { method: 'PUT', body: JSON.stringify(data) }),
    addToJob: (candidateId: string, jobId: string) =>
      request<{ id: string; stage: string; candidate_id: string; job_id: string }>(`/candidates/${candidateId}/applications`, { method: 'POST', body: JSON.stringify({ jobId }) }),
    listContacts: (candidateId: string) =>
      request<CandidateContact[]>(`/candidates/${candidateId}/contacts`),
    addContact: (candidateId: string, data: { category: string; label: string; value: string }) =>
      request<CandidateContact>(`/candidates/${candidateId}/contacts`, { method: 'POST', body: JSON.stringify(data) }),
    deleteContact: (candidateId: string, contactId: string) =>
      request<{ status: string }>(`/candidates/${candidateId}/contacts/${contactId}`, { method: 'DELETE' }),
    uploadResume: async (candidateId: string, file: File): Promise<Candidate> => {
      const form = new FormData();
      form.append('resume', file);
      const res = await fetch(`/api/candidates/${candidateId}/resume`, { method: 'POST', body: form, credentials: 'include' });
      if (!res.ok) { const e = await res.json().catch(() => ({})); throw new Error(e.error || 'Upload failed'); }
      return res.json();
    },
  },
  careers: {
    listJobs: (orgSlug: string) =>
      request<{ organization: CareerOrg; jobs: CareerJob[] }>(`/careers/${orgSlug}`),
    getJob: (orgSlug: string, jobId: string) =>
      request<CareerJob>(`/careers/${orgSlug}/jobs/${jobId}`),
    apply: async (orgSlug: string, jobId: string, data: { name: string; email: string; phone?: string; resume?: File }, ref?: string): Promise<ApplyResponse> => {
      const refParam = ref ? `?ref=${encodeURIComponent(ref)}` : '';
      if (data.resume) {
        const form = new FormData();
        form.append('name', data.name);
        form.append('email', data.email);
        if (data.phone) form.append('phone', data.phone);
        form.append('resume', data.resume);
        const res = await fetch(`${API_BASE}/careers/${orgSlug}/jobs/${jobId}/apply${refParam}`, {
          method: 'POST',
          credentials: 'include',
          body: form,
        });
        if (!res.ok) {
          const body = await res.json().catch(() => ({ error: res.statusText }));
          throw new ApiError(res.status, body.error || 'Request failed');
        }
        return res.json();
      }
      return request<ApplyResponse>(`/careers/${orgSlug}/jobs/${jobId}/apply${refParam}`, {
        method: 'POST',
        body: JSON.stringify({ name: data.name, email: data.email, phone: data.phone }),
      });
    },
  },
  team: {
    list: () => request<TeamData>('/team'),
    invite: (data: { email: string; role: string }) =>
      request<Invitation>('/team/invite', { method: 'POST', body: JSON.stringify(data) }),
    update: (userId: string, data: { role: string }) =>
      request<TeamMember>(`/team/${userId}`, { method: 'PUT', body: JSON.stringify(data) }),
    remove: (userId: string) =>
      request<{ status: string }>(`/team/${userId}`, { method: 'DELETE' }),
  },
  settings: {
    getDefaults: () => request<OrgDefaults>('/settings/defaults'),
    updateDefaults: (data: { defaultCompanyBlurb: string; defaultClosingStatement: string }) =>
      request<{ status: string }>('/settings/defaults', { method: 'PUT', body: JSON.stringify(data) }),
    updateOrgName: (data: { name: string }) =>
      request<{ status: string }>('/settings/org-name', { method: 'PUT', body: JSON.stringify(data) }),
    getSmtp: () => request<SmtpSettings>('/settings/smtp'),
    updateSmtp: (data: { host: string; port: number; username: string; password: string; fromEmail: string; fromName: string; tls: boolean }) =>
      request<{ status: string }>('/settings/smtp', { method: 'PUT', body: JSON.stringify(data) }),
    testSmtp: () =>
      request<{ status: string; sentTo: string }>('/settings/smtp/test', { method: 'POST' }),
    getEmailTemplates: () => request<EmailTemplate[]>('/settings/email-templates'),
    updateEmailTemplate: (data: { stage: string; subject: string; body: string; enabled: boolean }) =>
      request<EmailTemplate>('/settings/email-templates', { method: 'PUT', body: JSON.stringify(data) }),
  },
  account: {
    getCalendar: () => request<CalendarConnection | null>('/account/calendar'),
    disconnectCalendar: () => request<{ status: string }>('/account/calendar', { method: 'DELETE' }),
    getGoogleAuthUrl: () => request<{ url: string }>('/auth/google/url'),
  },
  scheduling: {
    checkAvailability: (appId: string, data: { interviewerIds: string[]; startDate: string; endDate: string }) =>
      request<AvailableBlock[]>(`/applications/${appId}/availability`, { method: 'POST', body: JSON.stringify(data) }),
    create: (appId: string, data: { slots: { start: string; end: string }[]; durationMinutes: number; location?: string; meetingUrl?: string; notes?: string; interviewerIds: string[] }) =>
      request<{ schedule: InterviewSchedule; slots: InterviewSlot[]; interviewers: ScheduleInterviewer[]; bookingLink: string }>(`/applications/${appId}/schedule`, { method: 'POST', body: JSON.stringify(data) }),
    list: (appId: string) => request<InterviewSchedule[]>(`/applications/${appId}/schedules`),
  },
  publicSchedule: {
    get: (token: string) => request<PublicSchedule>(`/schedule/${token}`),
    confirm: (token: string, slotId: string) =>
      request<{ status: string }>(`/schedule/${token}/confirm`, { method: 'POST', body: JSON.stringify({ slotId }) }),
  },
  query: {
    run: (sql: string) =>
      request<{ columns: string[]; rows: (string | number | boolean | null)[][]; count: number }>('/query', { method: 'POST', body: JSON.stringify({ sql }) }),
  },
  search: {
    resumes: (q: string) =>
      request<{ results: { id: string; name: string; email: string; company: string; resume_filename: string; snippet: string }[]; count: number }>('/search/resumes', { method: 'POST', body: JSON.stringify({ q }) }),
  },
  referrals: {
    list: () => request<Referral[]>('/referrals'),
    create: async (data: { name: string; email: string; phone?: string; jobId: string; resume?: File }): Promise<{ referralId: string; applicationId: string; candidateId: string }> => {
      if (data.resume) {
        const form = new FormData();
        form.append('name', data.name);
        form.append('email', data.email);
        if (data.phone) form.append('phone', data.phone);
        form.append('jobId', data.jobId);
        form.append('resume', data.resume);
        const res = await fetch(`${API_BASE}/referrals`, {
          method: 'POST',
          credentials: 'include',
          body: form,
        });
        if (!res.ok) {
          const body = await res.json().catch(() => ({ error: res.statusText }));
          throw new ApiError(res.status, body.error || 'Request failed');
        }
        return res.json();
      }
      return request('/referrals', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },
    createLink: (jobId: string) =>
      request<{ referralId: string; token: string }>('/referrals/link', { method: 'POST', body: JSON.stringify({ jobId }) }),
  },
};
