import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Jobs from './Jobs';
import { renderWithProviders } from '../test/helpers';
import { api } from '../api/client';

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return {
    ...actual,
    api: {
      ...actual.api,
      jobs: {
        list: vi.fn(),
      },
    },
  };
});

const mockJobs = [
  {
    id: 'job-1',
    title: 'Senior Engineer',
    location: 'Remote',
    department: 'Engineering',
    salary: '150k',
    status: 'open',
    requisition_id: null,
    organization_id: 'org-1',
    created_at: '2026-01-15T00:00:00Z',
    updated_at: '2026-01-15T00:00:00Z',
    company_blurb: '',
    team_details: '',
    responsibilities: '',
    qualifications: '',
    closing_statement: '',
  },
  {
    id: 'job-2',
    title: 'Product Manager',
    location: 'NYC',
    department: 'Product',
    salary: '140k',
    status: 'draft',
    requisition_id: null,
    organization_id: 'org-1',
    created_at: '2026-01-10T00:00:00Z',
    updated_at: '2026-01-10T00:00:00Z',
    company_blurb: '',
    team_details: '',
    responsibilities: '',
    qualifications: '',
    closing_statement: '',
  },
  {
    id: 'job-3',
    title: 'Designer',
    location: 'London',
    department: 'Design',
    salary: '120k',
    status: 'closed',
    requisition_id: null,
    organization_id: 'org-1',
    created_at: '2026-01-05T00:00:00Z',
    updated_at: '2026-01-05T00:00:00Z',
    company_blurb: '',
    team_details: '',
    responsibilities: '',
    qualifications: '',
    closing_statement: '',
  },
];

const listMock = vi.mocked(api.jobs.list);

beforeEach(() => {
  vi.clearAllMocks();
  listMock.mockResolvedValue(mockJobs);
});

describe('Jobs page', () => {
  it('renders job list on load', async () => {
    renderWithProviders(<Jobs />);

    await waitFor(() => {
      expect(screen.getByText('Senior Engineer')).toBeInTheDocument();
    });

    expect(screen.getByText('Product Manager')).toBeInTheDocument();
    expect(screen.getByText('Designer')).toBeInTheDocument();
    expect(listMock).toHaveBeenCalledWith(undefined);
  });

  it('shows empty state when no jobs exist', async () => {
    listMock.mockResolvedValue([]);

    renderWithProviders(<Jobs />);

    await waitFor(() => {
      expect(screen.getByText('No jobs yet.')).toBeInTheDocument();
    });
  });

  it('shows search empty state when search has no results', async () => {
    listMock.mockResolvedValueOnce(mockJobs);

    renderWithProviders(<Jobs />);

    await waitFor(() => {
      expect(screen.getByText('Senior Engineer')).toBeInTheDocument();
    });

    listMock.mockResolvedValueOnce([]);
    const user = userEvent.setup();
    const searchInput = screen.getByPlaceholderText(/search jobs/i);
    await user.type(searchInput, 'nonexistent');

    await waitFor(() => {
      expect(screen.getByText('No jobs match your search.')).toBeInTheDocument();
    });
  });

  it('calls API with search query after debounce', async () => {
    renderWithProviders(<Jobs />);

    await waitFor(() => {
      expect(screen.getByText('Senior Engineer')).toBeInTheDocument();
    });

    listMock.mockResolvedValueOnce([mockJobs[0]]);
    const user = userEvent.setup();
    const searchInput = screen.getByPlaceholderText(/search jobs/i);
    await user.type(searchInput, 'engineer');

    await waitFor(() => {
      expect(listMock).toHaveBeenCalledWith('engineer');
    });
  });

  it('displays correct status badges', async () => {
    renderWithProviders(<Jobs />);

    await waitFor(() => {
      expect(screen.getByText('open')).toBeInTheDocument();
    });

    expect(screen.getByText('draft')).toBeInTheDocument();
    expect(screen.getByText('closed')).toBeInTheDocument();
  });

  it('shows Create Job button for admin users', async () => {
    renderWithProviders(<Jobs />, {
      user: {
        id: 'admin-1',
        email: 'admin@test.com',
        name: 'Admin',
        role: 'admin',
        orgId: 'org-1',
        orgSlug: 'test-org',
        orgName: 'Test Org',
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Create Job')).toBeInTheDocument();
    });
  });

  it('hides Create Job button for recruiter users', async () => {
    renderWithProviders(<Jobs />, {
      user: {
        id: 'rec-1',
        email: 'rec@test.com',
        name: 'Recruiter',
        role: 'recruiter',
        orgId: 'org-1',
        orgSlug: 'test-org',
        orgName: 'Test Org',
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Senior Engineer')).toBeInTheDocument();
    });

    expect(screen.queryByText('Create Job')).not.toBeInTheDocument();
  });

  it('displays location and department columns', async () => {
    renderWithProviders(<Jobs />);

    await waitFor(() => {
      expect(screen.getByText('Remote')).toBeInTheDocument();
    });

    expect(screen.getByText('Engineering')).toBeInTheDocument();
    expect(screen.getByText('NYC')).toBeInTheDocument();
    expect(screen.getByText('Product')).toBeInTheDocument();
  });

  it('has pipeline links for each job', async () => {
    renderWithProviders(<Jobs />);

    await waitFor(() => {
      expect(screen.getByText('Senior Engineer')).toBeInTheDocument();
    });

    const pipelineLinks = screen.getAllByText('Pipeline');
    expect(pipelineLinks).toHaveLength(3);
  });
});
