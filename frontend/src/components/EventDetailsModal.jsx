import React, { useState } from 'react';
import { FaTimes, FaTrash } from 'react-icons/fa';

const formatRecurrence = (recurrence) => {
    if (!recurrence || recurrence === 'none') return 'One-time event';
    return recurrence.charAt(0).toUpperCase() + recurrence.slice(1);
};

const EventDetailsModal = ({ isOpen, onClose, event, onDelete }) => {
    const [showConfirm, setShowConfirm] = useState(false);

    if (!isOpen || !event) return null;

    const handleDeleteClick = (e) => {
        e.stopPropagation();
        setShowConfirm(true);
    };

    const handleConfirmDelete = (e) => {
        e.stopPropagation();
        console.log('Event being deleted:', event);
        console.log('Event ID being passed:', event.id);
        onDelete(event.id);
    };

    const handleCancelDelete = (e) => {
        e.stopPropagation();
        setShowConfirm(false);
    };

    return (
        <div className="modal-overlay">
            <div className="modal-content event-details-modal">
                <button className="close-button" onClick={onClose}>
                    <FaTimes />
                </button>
                <div className="event-header">
                    <h2>{event.title}</h2>
                    <button className="delete-button" onClick={handleDeleteClick}>
                        <FaTrash />
                    </button>
                </div>
                <div className="event-time">
                    <p>
                        <strong>Start:</strong> {new Date(event.start).toLocaleString()}
                    </p>
                    <p>
                        <strong>End:</strong> {new Date(event.end).toLocaleString()}
                    </p>
                    <p>
                        <strong>Recurrence:</strong> {formatRecurrence(event.recurrence)}
                    </p>
                </div>
                <div className="event-description">
                    <strong>Description:</strong>
                    <p>{event.description}</p>
                </div>
                {showConfirm && (
                    <div className="delete-confirmation">
                        <p>Are you sure you want to delete this event?</p>
                        <div className="confirmation-actions">
                            <button className="btn btn-secondary" onClick={handleCancelDelete}>
                                Cancel
                            </button>
                            <button className="btn btn-danger" onClick={handleConfirmDelete}>
                                Delete
                            </button>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default EventDetailsModal; 