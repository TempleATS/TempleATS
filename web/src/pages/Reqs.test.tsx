import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Reqs from './Reqs';
import { renderWithProviders } from '../test/helpers';
import { api } from '../api/client';

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return {
    ...actual,
    api: {
      ...actual.api,
      reqs: {
        list: vi.fn(),
      },
      team: {
        list: vi.fn(),
      },
    },
  };
});

const mockReqs: import('../api/client').Requisition[] = [
  {
    id: 'req-1',
    title: 'Backend Engineer',
    job_code: 'BE-001',
    level: 'Senior',
    department: 'Engineering',
    target_hires: 2,
    status: 'open',
    hiring_manager_id: 'hm-1',
    recruiter_id: 'rec-1',
    organization_id: 'org-1',
    opened_at: '2026-01-15T00:00:00Z',
    closed_at: null,
    created_at: '2026-01-15T00:00:00Z',
    updated_at: '2026-01-15T00:00:00Z',
  },
  {
    id: 'req-2',
    title: 'Product Lead',
    job_code: 'PL-001',
    level: 'Lead',
    department: 'Product',
    target_hires: 1,
    status: 'filled',
    hiring_manager_id: 'hm-2',
    recruiter_id: null,
    organization_id: 'org-1',
    opened_at: '2026-01-10T00:00:00Z',
    closed_at: '2026-02-01T00:00:00Z',
    created_at: '2026-01-10T00:00:00Z',
    updated_at: '2026-01-10T00:00:00Z',
  },
];

const mockTeam: import('../api/client').TeamData = {
  members: [
    { id: 'hm-1', email: 'hm1@test.com', name: 'Alice Manager', role: 'hiring_manager', organization_id: 'org-1', created_at: { Time: '2026-01-01T00:00:00Z', Valid: true } },
    { id: 'hm-2', email: 'hm2@test.com', name: 'Bob Director', role: 'hiring_manager', organization_id: 'org-1', created_at: { Time: '2026-01-01T00:00:00Z', Valid: true } },
  ],
  invitations: [],
};

const reqsListMock = vi.mocked(api.reqs.list);
const teamListMock = vi.mocked(api.team.list);

beforeEach(() => {
  vi.clearAllMocks();
  reqsListMock.mockResolvedValue(mockReqs);
  teamListMock.mockResolvedValue(mockTeam);
});

describe('Reqs page', () => {
  it('renders requisition list on load', async () => {
    renderWithProviders(<Reqs />);

    await waitFor(() => {
      expect(screen.getByText('Backend Engineer')).toBeInTheDocument();
    });

    expect(screen.getByText('Product Lead')).toBeInTheDocument();
    expect(reqsListMock).toHaveBeenCalledWith(undefined);
    expect(teamListMock).toHaveBeenCalledTimes(1);
  });

  it('shows empty state when no requisitions exist', async () => {
    reqsListMock.mockResolvedValue([]);

    renderWithProviders(<Reqs />);

    await waitFor(() => {
      expect(screen.getByText('No requisitions yet.')).toBeInTheDocument();
    });
  });

  it('shows search empty state when search has no results', async () => {
    renderWithProviders(<Reqs />);

    await waitFor(() => {
      expect(screen.getByText('Backend Engineer')).toBeInTheDocument();
    });

    reqsListMock.mockResolvedValueOnce([]);
    const user = userEvent.setup();
    const searchInput = screen.getByPlaceholderText(/search requisitions/i);
    await user.type(searchInput, 'nonexistent');

    await waitFor(() => {
      expect(screen.getByText('No requisitions match your search.')).toBeInTheDocument();
    });
  });

  it('calls API with search query after debounce', async () => {
    renderWithProviders(<Reqs />);

    await waitFor(() => {
      expect(screen.getByText('Backend Engineer')).toBeInTheDocument();
    });

    reqsListMock.mockResolvedValueOnce([mockReqs[0]]);
    const user = userEvent.setup();
    const searchInput = screen.getByPlaceholderText(/search requisitions/i);
    await user.type(searchInput, 'backend');

    await waitFor(() => {
      expect(reqsListMock).toHaveBeenCalledWith('backend');
    });
  });

  it('displays hiring manager names', async () => {
    renderWithProviders(<Reqs />);

    await waitFor(() => {
      expect(screen.getByText('Alice Manager')).toBeInTheDocument();
    });

    expect(screen.getByText('Bob Director')).toBeInTheDocument();
  });

  it('displays correct status badges', async () => {
    renderWithProviders(<Reqs />);

    await waitFor(() => {
      expect(screen.getByText('open')).toBeInTheDocument();
    });

    expect(screen.getByText('filled')).toBeInTheDocument();
  });

  it('displays job code and department', async () => {
    renderWithProviders(<Reqs />);

    await waitFor(() => {
      expect(screen.getByText('BE-001')).toBeInTheDocument();
    });

    expect(screen.getByText('Engineering')).toBeInTheDocument();
    expect(screen.getByText('PL-001')).toBeInTheDocument();
    expect(screen.getByText('Product')).toBeInTheDocument();
  });

  it('shows Create Requisition button for admin users', async () => {
    renderWithProviders(<Reqs />, {
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
      expect(screen.getByText('Create Requisition')).toBeInTheDocument();
    });
  });

  it('hides Create Requisition button for recruiter users', async () => {
    renderWithProviders(<Reqs />);

    await waitFor(() => {
      expect(screen.getByText('Backend Engineer')).toBeInTheDocument();
    });

    expect(screen.queryByText('Create Requisition')).not.toBeInTheDocument();
  });

  it('does not re-fetch team members on search', async () => {
    renderWithProviders(<Reqs />);

    await waitFor(() => {
      expect(screen.getByText('Backend Engineer')).toBeInTheDocument();
    });

    expect(teamListMock).toHaveBeenCalledTimes(1);

    reqsListMock.mockResolvedValueOnce([mockReqs[0]]);
    const user = userEvent.setup();
    await user.type(screen.getByPlaceholderText(/search requisitions/i), 'eng');

    await waitFor(() => {
      expect(reqsListMock).toHaveBeenCalledWith('eng');
    });

    // Team list should NOT have been called again
    expect(teamListMock).toHaveBeenCalledTimes(1);
  });
});
