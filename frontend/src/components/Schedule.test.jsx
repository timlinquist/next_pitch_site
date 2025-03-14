import React from 'react';
import { render, screen, fireEvent, waitFor, waitForElementToBeRemoved, act } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import Schedule from './Schedule';
import { useAuth0 } from '@auth0/auth0-react';


// Mock the Auth0 hook
vi.mock('@auth0/auth0-react', () => ({
    useAuth0: vi.fn()
}));

// Mock the FullCalendar component
vi.mock('@fullcalendar/react', () => ({
    default: vi.fn(props => {
        // Store the select handler so we can call it in our tests
        window.handleDateSelect = props.select;
        return (
            <div data-testid="full-calendar">
                <div data-testid="calendar-event" className={props.events[0]?.className}>
                    {props.events[0]?.title}
                </div>
            </div>
        );
    })
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
            loginWithRedirect: vi.fn(),
            user: mockUser
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
            <Auth0Provider>
                <BrowserRouter>
                    {component}
                </BrowserRouter>
            </Auth0Provider>
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

    it('shows login prompt when not authenticated', async () => {
        useAuth0.mockReturnValue({
            isAuthenticated: false,
            loginWithRedirect: vi.fn(),
            user: null
        });

        renderWithProviders(<Schedule />);
        expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument();
        expect(screen.getByText('Please log in to schedule appointments')).toBeInTheDocument();
    });

    it('prevents non-admin users from creating long events', async () => {
        // Mock non-admin user
        useAuth0.mockReturnValue({
            isAuthenticated: true,
            user: { email: 'test@example.com' }
        });

        useAuth0Context.mockReturnValue({
            isAdmin: false
        });

        renderWithProviders(<Schedule />);

        // Wait for loading to finish
        await waitForElementToBeRemoved(() => screen.queryByText('Loading...'));

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
            user: { email: 'test@example.com' }
        });
        useAuth0Context.mockReturnValue({ isAdmin: true });

        // Mock existing event
        global.fetch = vi.fn().mockResolvedValueOnce({
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

        renderWithProviders(<Schedule />);

        // Wait for loading to complete
        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });

        // Wait for the existing event to be displayed
        await waitFor(() => {
            expect(screen.getByText('Existing Event')).toBeInTheDocument();
        });

        // Clear any previous error state
        await act(async () => {
            window.handleDateSelect({
                start: new Date('2025-03-13T21:00:00.000Z'),
                end: new Date('2025-03-13T22:00:00.000Z')
            });
        });

        // Close the modal if it's open
        const cancelButton = screen.queryByRole('button', { name: /cancel/i });
        if (cancelButton) {
            await act(async () => {
                fireEvent.click(cancelButton);
            });

            // Wait for the modal to close
            await waitFor(() => {
                expect(screen.queryByRole('button', { name: /cancel/i })).not.toBeInTheDocument();
            });
        }

        // Simulate selecting an overlapping time slot
        act(() => {
            window.handleDateSelect({
                start: new Date('2025-03-13T22:30:00.000Z'),
                end: new Date('2025-03-13T23:30:00.000Z')
            });
        });

        // Wait for the modal to close
        await waitFor(() => {
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
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
});
