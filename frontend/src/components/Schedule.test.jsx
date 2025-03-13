import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { BrowserRouter } from 'react-router-dom';
import Schedule from './Schedule';

// Mock FullCalendar components
vi.mock('@fullcalendar/react', () => ({
    default: ({ events }) => (
        <div data-testid="calendar">
            {events.map(event => (
                <div key={event.id} data-testid="calendar-event">
                    {event.title}
                </div>
            ))}
        </div>
    ),
}));

// Mock the auth0 hook
vi.mock('@auth0/auth0-react', () => ({
    useAuth0: () => ({
        isAuthenticated: true,
        user: { email: 'test@example.com' },
        loginWithRedirect: vi.fn(),
    }),
}));

// Mock the react-router-dom hook with proper exports
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useLocation: () => ({
            pathname: '/schedule',
            search: '',
            state: null,
        }),
    };
});

const renderWithProviders = (component) => {
    return render(
        <BrowserRouter>
            {component}
        </BrowserRouter>
    );
};

describe('Schedule Component', () => {
    it('should handle empty schedule entries', async () => {
        // Mock the fetch function to return empty data
        global.fetch = vi.fn(() =>
            Promise.resolve({
                ok: true,
                json: () => Promise.resolve(null),
            })
        );

        renderWithProviders(<Schedule />);

        // Wait for the loading state to finish
        await screen.findByText('Schedule a Consultation');

        // Verify that no events are rendered
        const events = screen.queryAllByTestId('calendar-event');
        expect(events).toHaveLength(0);
    });

    it('should handle null schedule entries', async () => {
        // Mock the fetch function to return null
        global.fetch = vi.fn(() =>
            Promise.resolve({
                ok: true,
                json: () => Promise.resolve(null),
            })
        );

        renderWithProviders(<Schedule />);

        // Wait for the loading state to finish
        await screen.findByText('Schedule a Consultation');

        // Verify that no events are rendered
        const events = screen.queryAllByTestId('calendar-event');
        expect(events).toHaveLength(0);
    });

    it('should handle valid schedule entries', async () => {
        const mockEvents = [
            {
                id: 1,
                title: 'Test Event',
                start_time: '2024-03-13T10:00:00Z',
                end_time: '2024-03-13T11:00:00Z',
                description: 'Test Description',
                user_email: 'test@example.com',
            },
        ];

        // Mock the fetch function to return valid data
        global.fetch = vi.fn(() =>
            Promise.resolve({
                ok: true,
                json: () => Promise.resolve(mockEvents),
            })
        );

        renderWithProviders(<Schedule />);

        // Wait for the loading state to finish
        await screen.findByText('Schedule a Consultation');

        // Verify that the event is rendered
        const events = await screen.findAllByTestId('calendar-event');
        expect(events).toHaveLength(1);
        expect(events[0]).toHaveTextContent('Test Event');
    });
}); 