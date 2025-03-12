import React, { useState, useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { useLocation } from 'react-router-dom';

const EventModal = ({ isOpen, onClose, onSubmit, startTime, endTime, initialData }) => {
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const { isAuthenticated, loginWithRedirect, user } = useAuth0();
    const location = useLocation();

    // Reset form when modal is opened/closed or initialData changes
    useEffect(() => {
        console.log('[EventModal] Effect triggered - isOpen:', isOpen, 'initialData:', initialData);
        
        if (isOpen) {
            if (initialData) {
                console.log('[EventModal] Setting form data from initialData:', initialData);
                setTitle(initialData.title ?? '');
                setDescription(initialData.description ?? '');
            } else {
                console.log('[EventModal] No initialData, clearing form');
                setTitle('');
                setDescription('');
            }
        }
    }, [isOpen, initialData]);

    // Log state changes
    useEffect(() => {
        console.log('[EventModal] Current form state - title:', title, 'description:', description);
    }, [title, description]);

    if (!isOpen) return null;

    const handleLogin = async () => {
        // Store the form data before redirecting
        const formData = {
            title,
            description,
            selectedSlot: {
                start: startTime.toISOString(),
                end: endTime.toISOString()
            }
        };
        
        console.log('[EventModal] Storing form data before login:', formData);
        sessionStorage.setItem('pendingEventData', JSON.stringify(formData));
        
        // Add a delay to see the logs
        console.log('[EventModal] Redirecting in 2 seconds...');
        await new Promise(resolve => setTimeout(resolve, 10000));
        
        // Redirect to login with state
        loginWithRedirect({
            appState: {
                returnTo: location.pathname,
                selectedSlot: {
                    start: startTime.toISOString(),
                    end: endTime.toISOString()
                }
            }
        });
    };

    const handleSubmit = (e) => {
        e.preventDefault();
        onSubmit({
            title,
            description,
            start_time: startTime,
            end_time: endTime,
            user_email: user?.email
        });
        setTitle('');
        setDescription('');
        onClose();
    };

    return (
        <div className="modal-overlay">
            <div className="modal-content">
                <h2>Schedule New Event</h2>
                <form onSubmit={handleSubmit}>
                    <div className="form-group">
                        <label htmlFor="title">Title:</label>
                        <input
                            type="text"
                            id="title"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            required
                        />
                    </div>
                    <div className="form-group">
                        <label htmlFor="description">Description:</label>
                        <textarea
                            id="description"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            required
                        />
                    </div>
                    {!isAuthenticated && (
                        <div className="error">
                            Please <button onClick={handleLogin} className="link-button">Login</button> to create an appointment
                        </div>
                    )}
                    <div className="modal-actions">
                        <button type="button" onClick={onClose} className="btn btn-secondary">
                            Cancel
                        </button>
                        <button type="submit" className="btn" disabled={!isAuthenticated}>
                            {isAuthenticated ? 'Create Event' : 'Login to Create Event'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default EventModal; 