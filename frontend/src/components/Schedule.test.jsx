import React from 'react';
import { render, screen, fireEvent, waitFor, waitForElementToBeRemoved, act } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import Schedule, { formatEvents } from './Schedule';
import { useAuth0 } from '@auth0/auth0-react';
import { useAuth0Context } from '../contexts/Auth0Context';

// Mock the Auth0 hooks
vi.mock('@auth0/auth0-react', () => ({
    useAuth0: vi.fn()
}));

vi.mock('../contexts/Auth0Context', () => ({
    useAuth0Context: vi.fn()
}));

// Mock the FullCalendar component
vi.mock('@fullcalendar/react', () => ({
    default: ({ events, loading, select }) => {
        // Store the select handler globally so we can call it in our tests
        window.handleDateSelect = select;
        return (
            <div data-testid="full-calendar">
                {events?.map(event => (
                    <div 
                        key={event.id} 
                        data-testid="calendar-event"
                        className={event.className}
                    >
                        {event.title}
                    </div>
                ))}
                {loading && <div>Loading calendar...</div>}
            </div>
        );
    }
}));

// Mock the API calls
global.fetch = vi.fn();

describe('Schedule Component', () => {
    const mockUser = {
        email: 'test@example.com',
        name: 'Test User'
    };

    const mockEvents = [
        {
            id: 1,
            title: 'Test Event',
            start_time: new Date().toISOString(),
            end_time: new Date(Date.now() + 3600000).toISOString(),
            description: 'Test Description',
            user_email: 'test@example.com'
        }
    ];

    beforeEach(() => {
        // Reset all mocks
        vi.clearAllMocks();
        
        // Setup default Auth0 mock
        useAuth0.mockReturnValue({
            isAuthenticated: true,
            user: mockUser,
            loginWithRedirect: vi.fn(),
            getAccessTokenSilently: vi.fn().mockResolvedValue('mock-token')
        });

        // Setup default Auth0Context mock
        useAuth0Context.mockReturnValue({
            isAdmin: false
        });

        // Setup default fetch mock
        global.fetch.mockResolvedValue({
            ok: true,
            json: () => Promise.resolve(mockEvents)
        });
    });

    const renderWithProviders = (component) => {
        return render(
            <BrowserRouter>
                {component}
            </BrowserRouter>
        );
    };

    it('renders schedule component', async () => {
        renderWithProviders(<Schedule />);

        // Wait for loading to finish
        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });

        expect(screen.getByTestId('full-calendar')).toBeInTheDocument();
        expect(screen.getByTestId('calendar-event')).toBeInTheDocument();
    });

    it('shows login prompt when not authenticated', () => {
        useAuth0.mockReturnValue({
            isAuthenticated: false,
            loginWithRedirect: vi.fn(),
            user: null,
            getAccessTokenSilently: vi.fn()
        });

        renderWithProviders(<Schedule />);
        expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument();
        expect(screen.getByText('Please login or signup to schedule appointments')).toBeInTheDocument();
    });

    it('prevents non-admin users from creating long events', async () => {
        // Mock non-admin user
        useAuth0.mockReturnValue({
            isAuthenticated: true,
            user: { email: 'test@example.com' },
            getAccessTokenSilently: vi.fn().mockResolvedValue('mock-token')
        });

        useAuth0Context.mockReturnValue({
            isAdmin: false
        });

        renderWithProviders(<Schedule />);

        // Wait for loading to finish
        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });

        // Simulate selecting a time slot longer than 2 hours
        const selectInfo = {
            start: new Date('2025-03-13T10:00:00'),
            end: new Date('2025-03-13T13:00:00')
        };

        // Call the select handler directly
        window.handleDateSelect(selectInfo);

        // Wait for the error message to appear in an alert
        await waitFor(() => {
            const alert = screen.getByRole('alert');
            expect(alert).toHaveTextContent('Non-admin users cannot create events longer than 2 hours');
        });
    });

    it('prevents creating overlapping events', async () => {
        // Mock authenticated user
        useAuth0.mockReturnValue({
            isAuthenticated: true,
            user: { email: 'test@example.com' },
            getAccessTokenSilently: vi.fn().mockResolvedValue('mock-token')
        });
        useAuth0Context.mockReturnValue({ isAdmin: true });

        // Mock fetch responses
        global.fetch.mockImplementation((url) => {
            if (url.includes('users/me')) {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({ is_admin: true })
                });
            }
            return Promise.resolve({
                ok: true,
                json: () => Promise.resolve([{
                    id: 1,
                    title: 'Existing Event',
                    start_time: '2025-03-13T22:00:00.000Z',
                    end_time: '2025-03-13T23:00:00.000Z',
                    description: 'Existing Event Description',
                    user_email: 'other@example.com'
                }])
            });
        });

        renderWithProviders(<Schedule />);

        // Wait for loading to complete
        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });

        // Wait for the existing event to be displayed
        await waitFor(() => {
            expect(screen.getByText('Unavailable')).toBeInTheDocument();
        });

        // Simulate selecting an overlapping time slot
        act(() => {
            window.handleDateSelect({
                start: new Date('2025-03-13T22:30:00.000Z'),
                end: new Date('2025-03-13T23:30:00.000Z')
            });
        });

        // Wait for the error message to appear in an alert
        await waitFor(() => {
            expect(screen.getByRole('alert')).toBeInTheDocument();
            expect(screen.getByRole('alert')).toHaveTextContent('This time slot overlaps with an existing event');
        });
    });

    it('shows different styles for user events vs other events', async () => {
        const otherUserEvents = [
            {
                id: 1,
                title: 'Other User Event',
                start_time: new Date().toISOString(),
                end_time: new Date(Date.now() + 3600000).toISOString(),
                description: 'Other User Description',
                user_email: 'other@example.com'
            }
        ];

        global.fetch.mockResolvedValue({
            ok: true,
            json: () => Promise.resolve(otherUserEvents)
        });

        renderWithProviders(<Schedule />);

        // Wait for loading to finish
        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });

        const event = screen.getByTestId('calendar-event');
        expect(event).toHaveClass('other-event');
    });

    describe('formatEvents function', () => {
        const mockUser = {
            email: 'test@example.com',
            name: 'Test User'
        };

        it('handles null input', () => {
            const result = formatEvents(null, mockUser);
            expect(result).toEqual([]);
        });

        it('handles undefined input', () => {
            const result = formatEvents(undefined, mockUser);
            expect(result).toEqual([]);
        });

        it('handles empty array input', () => {
            const result = formatEvents([], mockUser);
            expect(result).toEqual([]);
        });

        it('handles array with null or undefined entries', () => {
            const result = formatEvents([null, undefined], mockUser);
            expect(result).toEqual([]);
        });

        it('formats a valid event entry', () => {
            const mockEvent = {
                id: 1,
                title: 'Test Event',
                start_time: '2024-03-24T10:00:00Z',
                end_time: '2024-03-24T11:00:00Z',
                user_email: 'test@example.com'
            };

            const result = formatEvents([mockEvent], mockUser);
            expect(result).toHaveLength(1);
            expect(result[0]).toMatchObject({
                id: 1,
                title: 'Test Event',
                className: 'user-event',
                extendedProps: expect.objectContaining({
                    isUnavailable: false
                })
            });
        });

        it('formats an event from another user as unavailable', () => {
            const mockEvent = {
                id: 1,
                title: 'Other User Event',
                start_time: '2024-03-24T10:00:00Z',
                end_time: '2024-03-24T11:00:00Z',
                user_email: 'other@example.com'
            };

            const result = formatEvents([mockEvent], mockUser);
            expect(result).toHaveLength(1);
            expect(result[0]).toMatchObject({
                id: 1,
                title: 'Unavailable',
                className: 'other-event',
                extendedProps: expect.objectContaining({
                    isUnavailable: true
                })
            });
        });
    });
});
