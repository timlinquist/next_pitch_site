import React, { useState, useEffect } from 'react';
import FullCalendar from '@fullcalendar/react';
import dayGridPlugin from '@fullcalendar/daygrid';
import timeGridPlugin from '@fullcalendar/timegrid';
import interactionPlugin from '@fullcalendar/interaction';
import { useAuth0 } from '@auth0/auth0-react';
import { useLocation } from 'react-router-dom';
import EventModal from './EventModal';
import EventDetailsModal from './EventDetailsModal';
import '../styles/calendar.css';
import { getApiUrl } from '../utils/api';
import { useAuth0Context } from '../contexts/Auth0Context';

const MAX_NON_ADMIN_DURATION = 2 * 60 * 60 * 1000; // 2 hours in milliseconds

const Schedule = () => {
    const [events, setEvents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isDetailsModalOpen, setIsDetailsModalOpen] = useState(false);
    const [selectedSlot, setSelectedSlot] = useState(null);
    const [selectedEvent, setSelectedEvent] = useState(null);
    const [deleteError, setDeleteError] = useState(null);
    const { isAuthenticated, loginWithRedirect, user, isAdmin } = useAuth0();
    const location = useLocation();
    const [initialEventData, setInitialEventData] = useState(null);
    
    useEffect(() => {
        if (isAuthenticated) {
            fetchScheduleEntries();
        } else {
            setLoading(false);
        }
    }, [isAuthenticated]);

    useEffect(() => {
        // Handle redirect with selected slot
        if (isAuthenticated) {
            let slot = null;
            let formData = null;

            // Check location state first
            if (location.state?.selectedSlot) {
                slot = location.state.selectedSlot;
            }

            // Check session storage for form data
            const pendingEventData = sessionStorage.getItem('pendingEventData');
            
            if (pendingEventData) {
                try {
                    formData = JSON.parse(pendingEventData);
                    
                    // If we didn't get the slot from location state, use it from form data
                    if (!slot && formData.selectedSlot) {
                        slot = formData.selectedSlot;
                    }
                } catch (err) {
                    console.error('[Schedule] Error parsing pendingEventData:', err);
                }
            }

            // If we have a slot, set it and open the modal
            if (slot) {
                setSelectedSlot({
                    start: new Date(slot.start),
                    end: new Date(slot.end)
                });

                // If we have form data, set it as initial data BEFORE opening the modal
                if (formData) {
                    setInitialEventData({
                        title: formData.title,
                        description: formData.description
                    });
                }
                
                // Open the modal AFTER setting both slot and initial data
                setIsModalOpen(true);

                // Clear the stored data AFTER everything is set up
                if (pendingEventData) {
                    sessionStorage.removeItem('pendingEventData');
                }
            }
        }
    }, [location.state, isAuthenticated]);

    const formatEvents = (entries, currentUserEmail) => {
        if (!entries || !Array.isArray(entries)) {
            return [];
        }
        return entries.map(entry => ({
            id: entry.id,
            title: entry.title,
            start: entry.start_time,
            end: entry.end_time,
            description: entry.description,
            user_email: entry.user_email,
            className: entry.user_email === currentUserEmail ? 'user-event' : 'other-event'
        }));
    };

    const fetchScheduleEntries = async () => {
        try {
            console.log('[Schedule] Fetching schedule entries, user:', user?.email);
            const response = await fetch(getApiUrl('schedule'), {
                method: 'GET',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                },
            });
            
            if (!response.ok) {
                console.error('[Schedule] HTTP error response:', response.status, response.statusText);
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            console.log('[Schedule] Received data:', data);
            
            const formattedEvents = formatEvents(data, user?.email);
            console.log('[Schedule] Formatted events:', formattedEvents);
            
            setEvents(formattedEvents);
            setLoading(false);
        } catch (err) {
            console.error('Error fetching schedule entries:', err);
            setError(err.message);
            setLoading(false);
        }
    };

    const handleDateSelect = (selectInfo) => {
        if (!isAuthenticated) {
            setError('Please log in to create events');
            return;
        }

        const start = selectInfo.start;
        const end = selectInfo.end;
        const duration = end - start;

        if (!isAdmin && duration > MAX_NON_ADMIN_DURATION) {
            setError('Non-admin users cannot create events longer than 2 hours');
            return;
        }

        // Check for overlapping events
        const hasOverlap = events.some(event => {
            return (start < event.end && end > event.start);
        });

        if (hasOverlap) {
            setIsModalOpen(false);
            setSelectedSlot(null);
            setError('This time slot overlaps with an existing event');
            return;
        }

        // Clear any previous error state when opening the modal
        setError(null);
        setSelectedSlot({ start, end });
        setIsModalOpen(true);
    };

    const handleEventClick = (clickInfo) => {
        if (!isAuthenticated || clickInfo.event.extendedProps.user_email !== user.email) {
            return;
        }

        setSelectedSlot({
            start: clickInfo.event.start,
            end: clickInfo.event.end,
            title: clickInfo.event.title,
            description: clickInfo.event.extendedProps.description,
        });
        setIsModalOpen(true);
    };

    const handleEventSubmit = async (eventData) => {
        try {
            const response = await fetch(getApiUrl('schedule'), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    ...eventData,
                    user_email: user.email,
                }),
            });

            if (!response.ok) {
                throw new Error('Failed to create event');
            }

            const newEvent = await response.json();
            const formattedEvent = {
                id: newEvent.id,
                title: newEvent.title,
                start: newEvent.start_time,
                end: newEvent.end_time,
                description: newEvent.description,
                user_email: newEvent.user_email,
                className: 'user-event'
            };

            setEvents([...events, formattedEvent]);
            setIsModalOpen(false);
            setSelectedSlot(null);
            setError(null);
        } catch (err) {
            console.error('Error creating event:', err);
            setError('Failed to create event. Please try again.');
        }
    };

    const handleEventDelete = async (eventId) => {
        try {
            // Convert eventId to number for comparison
            const eventIdNum = Number(eventId);
            
            // Store the event to be deleted in case we need to restore it
            const eventToDelete = events.find(e => e.id === eventIdNum);
            
            // Optimistically remove the event from the UI
            setEvents(prevEvents => prevEvents.filter(e => e.id !== eventIdNum));
            setIsDetailsModalOpen(false);

            const response = await fetch(getApiUrl(`schedule/${eventIdNum}`), {
                method: 'DELETE',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            // Clear any previous delete errors
            setDeleteError(null);
        } catch (err) {
            console.error('Error deleting event:', err);
            // Restore the event in the UI using the current events state
            setEvents(prevEvents => [...prevEvents, eventToDelete]);
            setDeleteError('Failed to delete event. Please try again.');
        }
    };

    if (loading) {
        return <div className="container">Loading...</div>;
    }

    if (!isAuthenticated) {
        return (
            <div className="container">
                <p>Please log in to schedule appointments</p>
                <button onClick={() => loginWithRedirect({ appState: { returnTo: '/schedule' } })} className="btn">Log In</button>
            </div>
        );
    }

    return (
        <div className="container">
            {error && (
                <div role="alert" className="alert alert-error">
                    {error}
                </div>
            )}
            {deleteError && (
                <div role="alert" className="alert alert-error">
                    {deleteError}
                </div>
            )}
            <div className="section">
                <h1>Schedule a Consultation</h1>
                <p>View our availability and schedule a consultation to discuss your pitch needs. We offer flexible scheduling options to accommodate your timeline.</p>
            </div>
            <div className="calendar-container">
                <FullCalendar
                    plugins={[dayGridPlugin, timeGridPlugin, interactionPlugin]}
                    initialView="timeGridWeek"
                    headerToolbar={{
                        left: 'prev,next today',
                        center: 'title',
                        right: 'dayGridMonth,timeGridWeek,timeGridDay'
                    }}
                    events={events}
                    selectable={true}
                    selectMirror={true}
                    dayMaxEvents={true}
                    weekends={true}
                    select={handleDateSelect}
                    eventClick={handleEventClick}
                    timeZone="local"
                    initialDate={new Date()}
                />
            </div>
            {isModalOpen && (
                <EventModal
                    isOpen={isModalOpen}
                    onClose={() => {
                        setIsModalOpen(false);
                        setSelectedSlot(null);
                        setError(null);
                    }}
                    onSubmit={handleEventSubmit}
                    initialData={selectedSlot}
                />
            )}
        </div>
    );
};

export default Schedule; 