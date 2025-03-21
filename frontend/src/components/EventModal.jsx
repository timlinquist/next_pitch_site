import React, { useState, useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { useLocation } from 'react-router-dom';

const RECURRENCE_OPTIONS = {
    NONE: 'none',
    WEEKLY: 'weekly',
    BIWEEKLY: 'biweekly',
    MONTHLY: 'monthly'
};

const EventModal = ({ isOpen, onClose, onSubmit, startTime, endTime, initialData }) => {
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const [recurrence, setRecurrence] = useState(RECURRENCE_OPTIONS.NONE);
    const { isAuthenticated, loginWithRedirect, user } = useAuth0();
    const location = useLocation();

    // Reset form when modal is opened/closed or initialData changes
    useEffect(() => {
        if (isOpen) {
            if (initialData) {
                setTitle(initialData.title ?? '');
                setDescription(initialData.description ?? '');
                setRecurrence(initialData.recurrence ?? RECURRENCE_OPTIONS.NONE);
            } else {
                setTitle('');
                setDescription('');
                setRecurrence(RECURRENCE_OPTIONS.NONE);
            }
        }
    }, [isOpen, initialData]);

    if (!isOpen) return null;

    const handleLogin = async () => {
        // Store the form data before redirecting
        const formData = {
            title,
            description,
            recurrence,
            selectedSlot: {
                start: startTime.toISOString(),
                end: endTime.toISOString()
            }
        };
        
        sessionStorage.setItem('pendingEventData', JSON.stringify(formData));
        
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
            user_email: user?.email,
            recurrence
        });
        setTitle('');
        setDescription('');
        setRecurrence(RECURRENCE_OPTIONS.NONE);
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
                    <div className="form-group">
                        <label htmlFor="recurrence">Recurrence:</label>
                        <select
                            id="recurrence"
                            value={recurrence}
                            onChange={(e) => setRecurrence(e.target.value)}
                        >
                            <option value={RECURRENCE_OPTIONS.NONE}>None</option>
                            <option value={RECURRENCE_OPTIONS.WEEKLY}>Weekly</option>
                            <option value={RECURRENCE_OPTIONS.BIWEEKLY}>Bi-weekly</option>
                            <option value={RECURRENCE_OPTIONS.MONTHLY}>Monthly</option>
                        </select>
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