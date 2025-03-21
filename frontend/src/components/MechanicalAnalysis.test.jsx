import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import MechanicalAnalysis from './MechanicalAnalysis';
import config from '../config';
import { useAuth0 } from '@auth0/auth0-react';

// Mock the Auth0 hook
vi.mock('@auth0/auth0-react', () => ({
    useAuth0: vi.fn()
}));

// Mock the config
vi.mock('../config', () => ({
    __esModule: true,
    default: {
        apiBaseUrl: 'http://test-api.com'
    }
}));

// Mock fetch
global.fetch = vi.fn();

describe('MechanicalAnalysis', () => {
    beforeEach(() => {
        // Clear all mocks before each test
        vi.clearAllMocks();

        // Setup default Auth0 mock
        useAuth0.mockReturnValue({
            getAccessTokenSilently: vi.fn().mockResolvedValue('test-token'),
            isAuthenticated: true,
            user: { email: 'test@example.com' }
        });
    });

    it('renders the component with upload forms', () => {
        render(<MechanicalAnalysis />);
        
        // Check for headings
        expect(screen.getByText('Mechanical Analysis')).toBeInTheDocument();
        expect(screen.getByText('Front View')).toBeInTheDocument();
        expect(screen.getByText('Side View')).toBeInTheDocument();
        
        // Check for file inputs
        expect(screen.getByTestId('front-video-input')).toBeInTheDocument();
        expect(screen.getByTestId('side-video-input')).toBeInTheDocument();
    });

    it('validates file size before upload', () => {
        render(<MechanicalAnalysis />);
        
        // Create a file larger than 10MB
        const largeFile = new File(['x'.repeat(11 * 1024 * 1024)], 'large.mp4', { type: 'video/mp4' });
        
        // Get the front video input and trigger change
        const frontInput = screen.getByTestId('front-video-input');
        fireEvent.change(frontInput, { target: { files: [largeFile] } });
        
        // Check for error message
        expect(screen.getByText('File too large. Maximum size is 10MB')).toBeInTheDocument();
        
        // Verify fetch was not called
        expect(fetch).not.toHaveBeenCalled();
    });

    it('validates file type before upload', () => {
        render(<MechanicalAnalysis />);
        
        // Create a non-video file
        const textFile = new File(['test'], 'test.txt', { type: 'text/plain' });
        
        // Get the front video input and trigger change
        const frontInput = screen.getByTestId('front-video-input');
        fireEvent.change(frontInput, { target: { files: [textFile] } });
        
        // Check for error message
        expect(screen.getByText('Invalid file type. Only video files are allowed')).toBeInTheDocument();
        
        // Verify fetch was not called
        expect(fetch).not.toHaveBeenCalled();
    });

    it('handles successful upload', async () => {
        render(<MechanicalAnalysis />);
        
        // Mock successful fetch response
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: () => Promise.resolve({ message: 'Upload successful' })
        });
        
        // Create a valid video file
        const videoFile = new File(['x'], 'test.mp4', { type: 'video/mp4' });
        
        // Get the front video input and trigger change
        const frontInput = screen.getByTestId('front-video-input');
        fireEvent.change(frontInput, { target: { files: [videoFile] } });
        
        // Wait for the upload to complete
        await waitFor(() => {
            expect(fetch).toHaveBeenCalledWith(
                `${config.apiBaseUrl}/video/upload`,
                expect.objectContaining({
                    method: 'POST',
                    headers: {
                        'Authorization': 'Bearer test-token'
                    },
                    body: expect.any(FormData)
                })
            );
        });

        // Verify the FormData contains the file
        const formData = fetch.mock.calls[0][1].body;
        expect(formData.get('video')).toEqual(videoFile);
    });

    it('handles upload error', async () => {
        render(<MechanicalAnalysis />);
        
        // Mock failed fetch response
        global.fetch.mockResolvedValueOnce({
            ok: false,
            json: () => Promise.resolve({ error: 'Upload failed' })
        });
        
        // Create a valid video file
        const videoFile = new File(['x'], 'test.mp4', { type: 'video/mp4' });
        
        // Get the front video input and trigger change
        const frontInput = screen.getByTestId('front-video-input');
        fireEvent.change(frontInput, { target: { files: [videoFile] } });
        
        // Wait for the error message
        await waitFor(() => {
            expect(screen.getByText('Upload failed')).toBeInTheDocument();
        });
        
        // Verify fetch was called with the correct parameters
        expect(fetch).toHaveBeenCalledWith(
            `${config.apiBaseUrl}/video/upload`,
            expect.objectContaining({
                method: 'POST',
                headers: {
                    'Authorization': 'Bearer test-token'
                },
                body: expect.any(FormData)
            })
        );
    });
}); 