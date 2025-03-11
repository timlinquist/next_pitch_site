import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { describe, it, expect, vi } from 'vitest';
import MechanicalAnalysisPage from './MechanicalAnalysisPage';
import MechanicalAnalysis from '../components/MechanicalAnalysis';

// Mock the MechanicalAnalysis component
vi.mock('../components/MechanicalAnalysis', () => ({
    default: function MockMechanicalAnalysis() {
        return <div data-testid="mock-mechanical-analysis">Mechanical Analysis Component</div>;
    }
}));

describe('MechanicalAnalysisPage', () => {
    it('renders the page with the MechanicalAnalysis component', () => {
        render(<MechanicalAnalysisPage />);
        
        // Check that the page container is present
        expect(screen.getByTestId('mock-mechanical-analysis')).toBeInTheDocument();
    });
}); 