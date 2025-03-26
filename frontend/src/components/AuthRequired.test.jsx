import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useAuth0 } from '@auth0/auth0-react';
import AuthRequired from './AuthRequired';

// Mock the Auth0 hook
vi.mock('@auth0/auth0-react', () => ({
    useAuth0: vi.fn()
}));

describe('AuthRequired', () => {
    const mockLoginWithRedirect = vi.fn();
    const TestComponent = () => <div>Protected Content</div>;

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('shows loading state when authentication is loading', () => {
        useAuth0.mockReturnValue({
            isLoading: true,
            isAuthenticated: false,
            loginWithRedirect: mockLoginWithRedirect
        });

        render(
            <AuthRequired returnTo="/test">
                <TestComponent />
            </AuthRequired>
        );

        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('shows login button when user is not authenticated', () => {
        useAuth0.mockReturnValue({
            isLoading: false,
            isAuthenticated: false,
            loginWithRedirect: mockLoginWithRedirect
        });

        render(
            <AuthRequired returnTo="/test">
                <TestComponent />
            </AuthRequired>
        );

        expect(screen.getByText('Please login or signup to continue')).toBeInTheDocument();
        
        const loginButton = screen.getByRole('button', { name: /log in/i });
        expect(loginButton).toBeInTheDocument();
        
        fireEvent.click(loginButton);
        expect(mockLoginWithRedirect).toHaveBeenCalledWith({
            appState: { returnTo: '/test' }
        });
    });

    it('renders children when user is authenticated', () => {
        useAuth0.mockReturnValue({
            isLoading: false,
            isAuthenticated: true,
            loginWithRedirect: mockLoginWithRedirect
        });

        render(
            <AuthRequired returnTo="/test">
                <TestComponent />
            </AuthRequired>
        );

        expect(screen.getByText('Protected Content')).toBeInTheDocument();
        expect(screen.queryByText('Please login or signup to continue')).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /log in/i })).not.toBeInTheDocument();
    });
}); 