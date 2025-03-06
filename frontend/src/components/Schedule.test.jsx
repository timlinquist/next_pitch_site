import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Schedule from './Schedule';

// Mock the fetch function
global.fetch = vi.fn();

// Mock FullCalendar component
vi.mock('@fullcalendar/react', () => ({
    default: ({ select, eventClick, events }) => (
        <div role="grid" onClick={() => select({ start: new Date(), end: new Date() })}>
            {events.map((event, index) => (
                <div
                    key={index}
                    role="button"
                    onClick={() => eventClick({ event: { title: event.title, extendedProps: { description: event.description } } })}
                >
                    {event.title}
                </div>
            ))}
        </div>
    )
}));

// Mock FullCalendar plugins
vi.mock('@fullcalendar/daygrid', () => ({
    default: {}
}));

vi.mock('@fullcalendar/timegrid', () => ({
    default: {}
}));

vi.mock('@fullcalendar/interaction', () => ({
    default: {}
}));

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

    it('opens modal when clicking on calendar', async () => {
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

        // Simulate calendar selection
        const calendar = screen.getByRole('grid');
        await userEvent.click(calendar);

        // Verify modal is opened
        expect(screen.getByText('Schedule New Event')).toBeInTheDocument();
        expect(screen.getByLabelText('Title:')).toBeInTheDocument();
        expect(screen.getByLabelText('Description:')).toBeInTheDocument();
    });

    it('creates new event successfully', async () => {
        // Mock initial fetch
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(mockEvents)
        });

        // Mock successful event creation
        const newEvent = {
            id: 3,
            title: 'New Meeting',
            description: 'Test meeting',
            start_time: '2025-03-06T10:00:00-08:00',
            end_time: '2025-03-06T11:00:00-08:00',
            created_at: '2025-03-05T18:00:28.849713-08:00',
            updated_at: '2025-03-05T18:00:28.849713-08:00'
        };

        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(newEvent)
        });

        await act(async () => {
            render(<Schedule />);
        });

        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });

        // Open modal
        const calendar = screen.getByRole('grid');
        await userEvent.click(calendar);

        // Fill in form
        const titleInput = screen.getByLabelText('Title:');
        const descriptionInput = screen.getByLabelText('Description:');
        await userEvent.type(titleInput, 'New Meeting');
        await userEvent.type(descriptionInput, 'Test meeting');

        // Submit form
        const submitButton = screen.getByText('Create Event');
        await userEvent.click(submitButton);

        // Verify new event was created
        await waitFor(() => {
            expect(screen.getByText('New Meeting')).toBeInTheDocument();
        });
    });

    it('handles event creation error', async () => {
        // Mock initial fetch
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(mockEvents)
        });

        // Mock failed event creation
        global.fetch.mockRejectedValueOnce(new Error('Failed to create event'));

        // Mock window.alert
        const mockAlert = vi.spyOn(window, 'alert').mockImplementation(() => {});

        await act(async () => {
            render(<Schedule />);
        });

        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });

        // Open modal
        const calendar = screen.getByRole('grid');
        await userEvent.click(calendar);

        // Fill in form
        const titleInput = screen.getByLabelText('Title:');
        const descriptionInput = screen.getByLabelText('Description:');
        await userEvent.type(titleInput, 'New Meeting');
        await userEvent.type(descriptionInput, 'Test meeting');

        // Submit form
        const submitButton = screen.getByText('Create Event');
        await userEvent.click(submitButton);

        // Verify error alert was shown
        expect(mockAlert).toHaveBeenCalledWith('Failed to create event. Please try again.');

        mockAlert.mockRestore();
    });

    it('closes modal when clicking cancel', async () => {
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

        // Open modal
        const calendar = screen.getByRole('grid');
        await userEvent.click(calendar);

        // Verify modal is open
        expect(screen.getByText('Schedule New Event')).toBeInTheDocument();

        // Click cancel
        const cancelButton = screen.getByText('Cancel');
        await userEvent.click(cancelButton);

        // Verify modal is closed
        expect(screen.queryByText('Schedule New Event')).not.toBeInTheDocument();
    });
}); 