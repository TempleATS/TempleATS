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
  level: string | null;
  department: string | null;
  target_hires: number;
  status: string;
  hiring_manager_id: string | null;
  organization_id: string;
  opened_at: string;
  closed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Job {
  id: string;
  title: string;
  description: string;
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
  description: string;
  location: { String: string; Valid: boolean } | null;
  department: { String: string; Valid: boolean } | null;
  salary: { String: string; Valid: boolean } | null;
  status: string;
  created_at: { Time: string; Valid: boolean };
}

export interface ApplyResponse {
  applicationId: string;
  candidateId: string;
  message: string;
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
  },
  reqs: {
    list: () => request<Requisition[]>('/reqs'),
    create: (data: { title: string; level?: string; department?: string; targetHires?: number }) =>
      request<Requisition>('/reqs', { method: 'POST', body: JSON.stringify(data) }),
    get: (id: string) => request<{ requisition: Requisition; jobs: Job[] }>(`/reqs/${id}`),
    update: (id: string, data: { title: string; level?: string; department?: string; targetHires?: number; status: string }) =>
      request<Requisition>(`/reqs/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    attachJob: (reqId: string, jobId: string) =>
      request<Job>(`/reqs/${reqId}/jobs`, { method: 'POST', body: JSON.stringify({ jobId }) }),
  },
  jobs: {
    list: () => request<Job[]>('/jobs'),
    create: (data: { title: string; description: string; location?: string; department?: string; salary?: string; status?: string; requisitionId?: string }) =>
      request<Job>('/jobs', { method: 'POST', body: JSON.stringify(data) }),
    get: (id: string) => request<Job>(`/jobs/${id}`),
    update: (id: string, data: { title: string; description: string; location?: string; department?: string; salary?: string; status: string; requisitionId?: string }) =>
      request<Job>(`/jobs/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  },
  careers: {
    listJobs: (orgSlug: string) =>
      request<{ organization: CareerOrg; jobs: CareerJob[] }>(`/careers/${orgSlug}`),
    getJob: (orgSlug: string, jobId: string) =>
      request<CareerJob>(`/careers/${orgSlug}/jobs/${jobId}`),
    apply: (orgSlug: string, jobId: string, data: { name: string; email: string; phone?: string; resumeUrl?: string; resumeFilename?: string }) =>
      request<ApplyResponse>(`/careers/${orgSlug}/jobs/${jobId}/apply`, { method: 'POST', body: JSON.stringify(data) }),
  },
};
