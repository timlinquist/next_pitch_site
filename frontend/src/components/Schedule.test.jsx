import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Schedule from './Schedule';

// Mock the fetch function
global.fetch = vi.fn();

describe('Schedule Component', () => {
    const mockEvents = [
        {
            id: 1,
            title: 'Team Standup',
            description: 'Daily team sync meeting',
            start_time: '2025-03-06T09:00:00-08:00',
            end_time: '2025-03-06T09:30:00-08:00',
            created_at: '2025-03-05T18:00:28.849713-08:00',
            updated_at: '2025-03-05T18:00:28.849713-08:00'
        },
        {
            id: 2,
            title: 'Project Review',
            description: 'Weekly project status review',
            start_time: '2025-03-06T14:00:00-08:00',
            end_time: '2025-03-06T15:00:00-08:00',
            created_at: '2025-03-05T18:00:28.849713-08:00',
            updated_at: '2025-03-05T18:00:28.849713-08:00'
        }
    ];

    beforeEach(() => {
        vi.resetAllMocks();
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('renders loading state initially', () => {
        // Mock fetch to not resolve immediately
        global.fetch.mockImplementationOnce(() => new Promise(() => {}));

        render(<Schedule />);
        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('renders error state when fetch fails', async () => {
        // Mock fetch to return an error
        global.fetch.mockRejectedValueOnce(new Error('Failed to fetch'));

        await act(async () => {
            render(<Schedule />);
        });

        await waitFor(() => {
            expect(screen.getByText('Error loading schedule')).toBeInTheDocument();
        });
    });

    it('renders schedule entries when fetch succeeds', async () => {
        // Mock fetch to return a successful response
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(mockEvents)
        });

        await act(async () => {
            render(<Schedule />);
        });

        await waitFor(() => {
            expect(screen.getByText('Team Standup')).toBeInTheDocument();
            expect(screen.getByText('Project Review')).toBeInTheDocument();
        });
    });

    it('displays correct event information', async () => {
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(mockEvents)
        });

        await act(async () => {
            render(<Schedule />);
        });

        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });

        // Check if event titles are displayed
        expect(screen.getByText('Team Standup')).toBeInTheDocument();
        expect(screen.getByText('Project Review')).toBeInTheDocument();
    });

    it('handles event click correctly', async () => {
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(mockEvents)
        });

        // Mock window.alert
        const mockAlert = vi.spyOn(window, 'alert').mockImplementation(() => {});

        await act(async () => {
            render(<Schedule />);
        });

        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });

        // Find and click an event
        const event = screen.getByText('Team Standup');
        await userEvent.click(event);

        // Verify alert was called with correct information
        expect(mockAlert).toHaveBeenCalledWith(
            expect.stringContaining('Team Standup')
        );

        mockAlert.mockRestore();
    });

    it('displays correct event details', async () => {
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(mockEvents)
        });

        await act(async () => {
            render(<Schedule />);
        });

        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });

        // Check if event titles and descriptions are displayed
        expect(screen.getByText('Team Standup')).toBeInTheDocument();
        expect(screen.getByText('Project Review')).toBeInTheDocument();
    });
}); 