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

const Schedule = () => {
    const [events, setEvents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isDetailsModalOpen, setIsDetailsModalOpen] = useState(false);
    const [selectedSlot, setSelectedSlot] = useState(null);
    const [selectedEvent, setSelectedEvent] = useState(null);
    const [deleteError, setDeleteError] = useState(null);
    const { isAuthenticated, loginWithRedirect, user } = useAuth0();
    const location = useLocation();
    const [initialEventData, setInitialEventData] = useState(null);

    useEffect(() => {
        // Remove the user check - we want to fetch entries regardless of auth status
        fetchScheduleEntries();
    }, []); // Remove user from dependencies since we don't need it for fetching

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
            classNames: entry.user_email === currentUserEmail ? [] : ['other-user-event']
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
        // Always set the selected slot and open the modal first
        setSelectedSlot({
            start: selectInfo.start,
            end: selectInfo.end
        });
        setIsModalOpen(true);
    };

    const handleEventClick = (clickInfo) => {
        setSelectedEvent({
            id: Number(clickInfo.event.id),
            title: clickInfo.event.title,
            description: clickInfo.event.extendedProps.description,
            start: clickInfo.event.start,
            end: clickInfo.event.end
        });
        setIsDetailsModalOpen(true);
    };

    const handleEventSubmit = async (eventData) => {
        try {
            if (!isAuthenticated) {
                // Store the form data in sessionStorage before redirecting
                const formData = {
                    ...eventData,
                    selectedSlot: {
                        start: selectedSlot.start.toISOString(),
                        end: selectedSlot.end.toISOString()
                    }
                };
                sessionStorage.setItem('pendingEventData', JSON.stringify(formData));
                
                console.log('[Schedule] Redirecting to login with form data stored');
                
                // Redirect to login with appState
                await loginWithRedirect({
                    appState: {
                        returnTo: location.pathname + location.search,
                        selectedSlot: {
                            start: selectedSlot.start.toISOString(),
                            end: selectedSlot.end.toISOString()
                        }
                    }
                });
                return;
            }

            if (!user?.email) {
                throw new Error('User must be authenticated to create events');
            }

            const eventWithEmail = {
                ...eventData,
                user_email: user.email
            };

            const response = await fetch(getApiUrl('schedule'), {
                method: 'POST',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(eventWithEmail)
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const newEvent = await response.json();
            const formattedEvent = formatEvents([{
                id: newEvent.id,
                title: newEvent.title,
                start_time: newEvent.start_time,
                end_time: newEvent.end_time,
                description: newEvent.description,
                user_email: user.email
            }], user.email)[0];

            setEvents([...events, formattedEvent]);
        } catch (err) {
            console.error('Error creating event:', err);
            alert('Failed to create event. Please try again.');
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

    if (error) {
        return <div className="container">Error loading schedule</div>;
    }

    return (
        <div className="container">
            <div className="section">
                <h1>Schedule a Consultation</h1>
                <p>View our availability and schedule a consultation to discuss your pitch needs. We offer flexible scheduling options to accommodate your timeline.</p>
            </div>
            {deleteError && (
                <div className="alert alert-error">
                    {deleteError}
                </div>
            )}
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
            <EventModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                onSubmit={handleEventSubmit}
                startTime={selectedSlot?.start}
                endTime={selectedSlot?.end}
                initialData={initialEventData}
            />
            <EventDetailsModal
                isOpen={isDetailsModalOpen}
                onClose={() => setIsDetailsModalOpen(false)}
                event={selectedEvent}
                onDelete={handleEventDelete}
            />
        </div>
    );
};

export default Schedule; 