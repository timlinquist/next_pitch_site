import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { describe, it, expect, vi } from 'vitest';
import MechanicalAnalysisPage from './MechanicalAnalysisPage';
import MechanicalAnalysis from '../components/MechanicalAnalysis';
import { useAuth0 } from '@auth0/auth0-react';

// Mock the Auth0 hook
vi.mock('@auth0/auth0-react', () => ({
    useAuth0: vi.fn()
}));

// Mock the MechanicalAnalysis component
vi.mock('../components/MechanicalAnalysis', () => ({
    default: function MockMechanicalAnalysis() {
        return <div data-testid="mock-mechanical-analysis">Mechanical Analysis Component</div>;
    }
}));

describe('MechanicalAnalysisPage', () => {
    it('renders the page with the MechanicalAnalysis component', async () => {
        // Mock authenticated user
        useAuth0.mockReturnValue({
            isLoading: false,
            isAuthenticated: true,
            loginWithRedirect: vi.fn()
        });

        render(<MechanicalAnalysisPage />);
        
        // Wait for loading to complete and check that the mock component is present
        await waitFor(() => {
            expect(screen.getByTestId('mock-mechanical-analysis')).toBeInTheDocument();
        });
    });
}); 