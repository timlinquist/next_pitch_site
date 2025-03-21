import { render, screen, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BrowserRouter } from 'react-router-dom';
import AccountPage from './AccountPage';
import { useAuth0 } from '@auth0/auth0-react';

// Mock the auth0 hook
vi.mock('@auth0/auth0-react', () => ({
    useAuth0: vi.fn(),
}));

// Mock fetch globally
global.fetch = vi.fn();

const renderWithProviders = (component) => {
    return render(
        <BrowserRouter>
            {component}
        </BrowserRouter>
    );
};

describe('AccountPage Component', () => {
    beforeEach(() => {
        // Reset all mocks before each test
        vi.clearAllMocks();
        
        // Setup default fetch mock
        global.fetch.mockResolvedValue({
            ok: true,
            json: () => Promise.resolve([]),
        });

        // Setup default Auth0 mock
        useAuth0.mockReturnValue({
            isAuthenticated: true,
            user: { email: 'test@example.com', name: 'Test User' },
            loginWithRedirect: vi.fn(),
            logout: vi.fn(),
            getAccessTokenSilently: vi.fn().mockResolvedValue('test-token'),
        });
    });

    it('shows login button when not authenticated', () => {
        useAuth0.mockReturnValue({
            isAuthenticated: false,
            user: null,
            loginWithRedirect: vi.fn(),
            logout: vi.fn(),
            getAccessTokenSilently: vi.fn().mockResolvedValue('test-token'),
        });

        renderWithProviders(<AccountPage />);
        expect(screen.getByText('Login')).toBeInTheDocument();
        expect(screen.getByText('Please log in to view your account information.')).toBeInTheDocument();
    });

    it('shows user information when authenticated', async () => {
        renderWithProviders(<AccountPage />);
        
        // Wait for loading to finish
        await waitFor(() => {
            expect(screen.queryByText('Loading appointments...')).not.toBeInTheDocument();
        });

        expect(screen.getByText('Test User')).toBeInTheDocument();
        expect(screen.getByText('test@example.com')).toBeInTheDocument();
    });

    it('shows loading state initially', async () => {
        // Mock a delayed fetch to ensure loading state is visible
        global.fetch.mockImplementationOnce(() => new Promise(resolve => setTimeout(resolve, 100)));

        renderWithProviders(<AccountPage />);
        
        // Check for loading state
        expect(screen.getByText('Loading appointments...')).toBeInTheDocument();
    });

    it('handles null appointments data gracefully', async () => {
        // Mock fetch to return null
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(null),
        });

        renderWithProviders(<AccountPage />);

        // Wait for loading to finish
        await waitFor(() => {
            expect(screen.queryByText('Loading appointments...')).not.toBeInTheDocument();
        });

        // Should show "No upcoming appointments" message
        await waitFor(() => {
            expect(screen.getByText(/no upcoming appointments/i)).toBeInTheDocument();
        });
    });

    it('handles empty appointments array', async () => {
        // Mock fetch to return empty array
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve([]),
        });

        renderWithProviders(<AccountPage />);

        // Wait for loading to finish
        await waitFor(() => {
            expect(screen.queryByText('Loading appointments...')).not.toBeInTheDocument();
        });

        // Should show "No upcoming appointments" message
        await waitFor(() => {
            expect(screen.getByText(/no upcoming appointments/i)).toBeInTheDocument();
        });
    });

    it('displays appointments when data is available', async () => {
        const mockAppointments = [
            {
                id: 1,
                title: 'Test Appointment',
                start_time: '2024-03-20T10:00:00Z',
                end_time: '2024-03-20T11:00:00Z',
                description: 'Test Description',
            },
        ];

        // Mock fetch to return appointments
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve(mockAppointments),
        });

        renderWithProviders(<AccountPage />);

        // Wait for loading to finish
        await waitFor(() => {
            expect(screen.queryByText('Loading appointments...')).not.toBeInTheDocument();
        });

        // Should show appointment details
        await waitFor(() => {
            expect(screen.getByText(/test appointment/i)).toBeInTheDocument();
            expect(screen.getByText(/test description/i)).toBeInTheDocument();
        });
    });

    it('handles fetch error gracefully', async () => {
        // Mock fetch to return error
        global.fetch.mockResolvedValueOnce({
            ok: false,
            status: 500,
        });

        renderWithProviders(<AccountPage />);

        // Wait for loading to finish
        await waitFor(() => {
            expect(screen.queryByText('Loading appointments...')).not.toBeInTheDocument();
        });

        // Should show error message
        expect(screen.getByText('Failed to load appointments. Please try again later.')).toBeInTheDocument();
    });
}); 