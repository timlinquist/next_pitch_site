import React, { useState } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import '../styles/mechanical-analysis.css';
import config from '../config';

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB in bytes

const MechanicalAnalysis = () => {
    const { getAccessTokenSilently } = useAuth0();
    const [frontVideo, setFrontVideo] = useState(null);
    const [sideVideo, setSideVideo] = useState(null);
    const [frontProgress, setFrontProgress] = useState(0);
    const [sideProgress, setSideProgress] = useState(0);
    const [frontError, setFrontError] = useState('');
    const [sideError, setSideError] = useState('');
    const [frontComplete, setFrontComplete] = useState(false);
    const [sideComplete, setSideComplete] = useState(false);

    const validateFile = (file) => {
        if (file.size > MAX_FILE_SIZE) {
            throw new Error(`File too large. Maximum size is ${MAX_FILE_SIZE / (1024 * 1024)}MB`);
        }
        if (!file.type.startsWith('video/')) {
            throw new Error('Invalid file type. Only video files are allowed');
        }
        return true;
    };

    const handleVideoUpload = async (file, type) => {
        try {
            // Validate file before upload
            validateFile(file);

            const token = await getAccessTokenSilently();
            const formData = new FormData();
            formData.append('video', file);

            const response = await fetch(`${config.apiBaseUrl}/video/upload`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData,
                onUploadProgress: (progressEvent) => {
                    const progress = (progressEvent.loaded / progressEvent.total) * 100;
                    if (type === 'front') {
                        setFrontProgress(progress);
                    } else {
                        setSideProgress(progress);
                    }
                }
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Upload failed');
            }

            const result = await response.json();
            
            // Update completion status
            if (type === 'front') {
                setFrontComplete(true);
                setFrontError('');
            } else {
                setSideComplete(true);
                setSideError('');
            }

            console.log(`Upload successful: ${result.link}`);
        } catch (error) {
            console.error('Upload error:', error);
            if (type === 'front') {
                setFrontError(error.message);
                setFrontComplete(false);
            } else {
                setSideError(error.message);
                setSideComplete(false);
            }
        }
    };

    const handleFrontVideoChange = (e) => {
        const file = e.target.files[0];
        if (file) {
            setFrontVideo(file);
            handleVideoUpload(file, 'front');
        }
    };

    const handleSideVideoChange = (e) => {
        const file = e.target.files[0];
        if (file) {
            setSideVideo(file);
            handleVideoUpload(file, 'side');
        }
    };

    return (
        <div className="mechanical-analysis-container">
            <div className="upload-section">
                <div className="upload-form">
                    <h2>Front View</h2>
                    <p className="upload-description">Upload a video taken from directly in front of the pitcher</p>
                    <div className="file-input-wrapper">
                        <input
                            type="file"
                            accept="video/*"
                            onChange={handleFrontVideoChange}
                            className="file-input"
                            id="front-video"
                            data-testid="front-video-input"
                        />
                        <label htmlFor="front-video" className="file-label">
                            {frontVideo ? frontVideo.name : 'Choose Video'}
                        </label>
                    </div>
                    {frontProgress > 0 && (
                        <div className="progress-container">
                            <div className="progress-bar">
                                <div 
                                    className={`progress-fill ${frontComplete ? 'complete' : ''}`}
                                    style={{ width: `${frontProgress}%` }}
                                />
                            </div>
                            {frontComplete && (
                                <span className="complete-icon">✓</span>
                            )}
                        </div>
                    )}
                    {frontError && <p className="error-message">{frontError}</p>}
                </div>

                <div className="upload-form">
                    <h2>Side View</h2>
                    <p className="upload-description">Upload a video taken from the pitcher's side</p>
                    <div className="file-input-wrapper">
                        <input
                            type="file"
                            accept="video/*"
                            onChange={handleSideVideoChange}
                            className="file-input"
                            id="side-video"
                            data-testid="side-video-input"
                        />
                        <label htmlFor="side-video" className="file-label">
                            {sideVideo ? sideVideo.name : 'Choose Video'}
                        </label>
                    </div>
                    {sideProgress > 0 && (
                        <div className="progress-container">
                            <div className="progress-bar">
                                <div 
                                    className={`progress-fill ${sideComplete ? 'complete' : ''}`}
                                    style={{ width: `${sideProgress}%` }}
                                />
                            </div>
                            {sideComplete && (
                                <span className="complete-icon">✓</span>
                            )}
                        </div>
                    )}
                    {sideError && <p className="error-message">{sideError}</p>}
                </div>
            </div>
        </div>
    );
};

export default MechanicalAnalysis; 