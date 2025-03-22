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

const MAX_NON_ADMIN_DURATION = 2 * 60 * 60 * 1000; // 2 hours in milliseconds

/**
 * Expands recurring events into their instances based on the current calendar view
 * @param {Array} events - Array of events from the API
 * @param {Date} viewStart - Start date of the current calendar view
 * @param {Date} viewEnd - End date of the current calendar view
 * @returns {Array} Array of all events including recurring instances
 */
const expandRecurringEvents = (events, viewStart, viewEnd) => {
    if (!events || !Array.isArray(events)) return [];
    if (!viewStart || !viewEnd) return events; // Return original events if view dates not set
    
    console.log('[expandRecurringEvents] Processing events:', events);
    console.log('[expandRecurringEvents] View range:', { viewStart, viewEnd });
    
    const expandedEvents = [];
    
    events.forEach(event => {
        // Add the original event
        expandedEvents.push(event);
        
        // Skip if no recurrence
        if (!event.recurrence || event.recurrence === 'none') {
            console.log('[expandRecurringEvents] Skipping non-recurring event:', event.id);
            return;
        }
        
        console.log('[expandRecurringEvents] Processing recurring event:', event.id, event.recurrence);
        
        const originalStart = new Date(event.start_time);
        const originalEnd = new Date(event.end_time);
        
        // Calculate duration in milliseconds
        const duration = originalEnd - originalStart;
        
        // Get the day of week (0-6) for weekly/biweekly recurrence
        const originalDayOfWeek = originalStart.getDay();
        
        // Get the date of month (1-31) for monthly recurrence
        const originalDateOfMonth = originalStart.getDate();
        
        // Start from the day after the original event
        let currentDate = new Date(originalStart);
        currentDate.setDate(currentDate.getDate() + 1);
        
        // Keep generating instances until we're past the view end
        while (currentDate <= viewEnd) {
            let shouldCreateInstance = false;
            
            switch (event.recurrence) {
                case 'weekly':
                    shouldCreateInstance = currentDate.getDay() === originalDayOfWeek;
                    break;
                    
                case 'biweekly':
                    // Check if we're on the same day of week and if the week number difference is even
                    const weekDiff = Math.floor((currentDate - originalStart) / (7 * 24 * 60 * 60 * 1000));
                    shouldCreateInstance = currentDate.getDay() === originalDayOfWeek && weekDiff % 2 === 0;
                    break;
                    
                case 'monthly':
                    shouldCreateInstance = currentDate.getDate() === originalDateOfMonth;
                    break;
            }
            
            if (shouldCreateInstance) {
                // Create a new instance of the event
                const instanceStart = new Date(currentDate);
                instanceStart.setHours(originalStart.getHours(), originalStart.getMinutes());
                
                const instanceEnd = new Date(instanceStart);
                instanceEnd.setTime(instanceStart.getTime() + duration);
                
                // Only add instances that fall within the view range
                if (instanceStart >= viewStart && instanceEnd <= viewEnd) {
                    console.log('[expandRecurringEvents] Creating instance:', {
                        originalId: event.id,
                        instanceStart,
                        instanceEnd
                    });
                    
                    expandedEvents.push({
                        ...event,
                        id: `${event.id}-${instanceStart.getTime()}`, // Unique ID for each instance
                        start_time: instanceStart.toISOString(),
                        end_time: instanceEnd.toISOString(),
                        isRecurringInstance: true,
                        originalEventId: event.id
                    });
                }
            }
            
            // Move to next day
            currentDate.setDate(currentDate.getDate() + 1);
        }
    });
    
    console.log('[expandRecurringEvents] Final expanded events:', expandedEvents);
    return expandedEvents;
};

const Schedule = () => {
    const { user, isAuthenticated, loginWithRedirect, getAccessTokenSilently } = useAuth0();
    const [events, setEvents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isDetailsModalOpen, setIsDetailsModalOpen] = useState(false);
    const [selectedSlot, setSelectedSlot] = useState(null);
    const [selectedEvent, setSelectedEvent] = useState(null);
    const [deleteError, setDeleteError] = useState(null);
    const [isAdmin, setIsAdmin] = useState(false);
    const location = useLocation();
    const [initialEventData, setInitialEventData] = useState(null);
    const [viewDates, setViewDates] = useState({ start: null, end: null });

    const checkAdminStatus = async () => {
        try {
            const token = await getAccessTokenSilently();
            console.log('[Schedule] Got access token:', token);
            
            const response = await fetch(getApiUrl('users/me'), {
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                }
            });
            
            if (response.ok) {
                const userData = await response.json();
                setIsAdmin(userData.is_admin);
            } else {
                console.error('Failed to fetch user admin status:', response.status);
                setIsAdmin(false);
            }
        } catch (error) {
            console.error('Error checking admin status:', error);
            setIsAdmin(false);
        }
    };

    useEffect(() => {
        if (isAuthenticated) {
            checkAdminStatus();
            fetchScheduleEntries();
        } else {
            setLoading(false);
        }
    }, [isAuthenticated, getAccessTokenSilently]);

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
                        description: formData.description,
                        recurrence: formData.recurrence || 'none'
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

    // Add effect to refetch events when view dates change
    useEffect(() => {
        if (viewDates.start && viewDates.end) {
            console.log('[Schedule] View dates changed, refetching events');
            fetchScheduleEntries();
        }
    }, [viewDates]);

    const handleDatesSet = (dateInfo) => {
        console.log('[Schedule] View dates set:', dateInfo);
        setViewDates({
            start: dateInfo.start,
            end: dateInfo.end
        });
    };

    const formatEvents = (entries, currentUserEmail) => {
        if (!entries || !Array.isArray(entries)) {
            return [];
        }

        console.log('[formatEvents] Processing entries:', entries);
        console.log('[formatEvents] Current view dates:', viewDates);

        // Expand recurring events using the current view dates
        const expandedEntries = expandRecurringEvents(entries, viewDates.start, viewDates.end);

        const formattedEvents = expandedEntries.map(entry => ({
            id: entry.id,
            title: isAdmin || entry.user_email === currentUserEmail ? entry.title : 'Unavailable',
            start: entry.start_time,
            end: entry.end_time,
            description: isAdmin || entry.user_email === currentUserEmail ? entry.description : '',
            user_email: entry.user_email,
            recurrence: entry.recurrence,
            className: entry.user_email === currentUserEmail ? 'user-event' : 'other-event',
            editable: isAdmin,
            interactive: isAdmin || entry.user_email === currentUserEmail
        }));

        console.log('[formatEvents] Final formatted events:', formattedEvents);
        return formattedEvents;
    };

    const fetchScheduleEntries = async () => {
        try {
            console.log('[Schedule] Fetching schedule entries, user:', user?.email);
            const token = await getAccessTokenSilently();
            const response = await fetch(getApiUrl('schedule'), {
                method: 'GET',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
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

        const start = new Date(selectInfo.start);
        const end = new Date(selectInfo.end);
        const duration = end - start;

        if (!isAdmin && duration > MAX_NON_ADMIN_DURATION) {
            setError('Non-admin users cannot create events longer than 2 hours');
            return;
        }

        // Check for overlapping events - including other users' events
        const hasOverlap = events.some(event => {
            const eventStart = new Date(event.start);
            const eventEnd = new Date(event.end);
            return (start < eventEnd && end > eventStart);
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
        if (!isAuthenticated) {
            setError('Please log in to view event details');
            return;
        }

        const eventUserEmail = clickInfo.event.extendedProps.user_email;
        const isUsersEvent = eventUserEmail === user.email;

        // Allow admins to view any event, otherwise only show user's own events
        if (!isAdmin && !isUsersEvent) {
            return;
        }

        // Set the selected event for the details modal
        setSelectedEvent({
            id: clickInfo.event.id,
            title: clickInfo.event.title,
            start: clickInfo.event.start,
            end: clickInfo.event.end,
            description: clickInfo.event.extendedProps.description,
            user_email: eventUserEmail,
            recurrence: clickInfo.event.extendedProps.recurrence || 'none'
        });

        setIsDetailsModalOpen(true);
    };

    const handleEventSubmit = async (eventData) => {
        try {
            const token = await getAccessTokenSilently();
            const response = await fetch(getApiUrl('schedule'), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    title: eventData.title,
                    description: eventData.description,
                    start_time: eventData.start_time.toISOString(),
                    end_time: eventData.end_time.toISOString(),
                    user_email: user.email,
                    recurrence: eventData.recurrence || 'none'
                }),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to create event');
            }

            const newEvent = await response.json();
            const formattedEvent = formatEvents([newEvent], user.email)[0];

            setEvents(prevEvents => [...prevEvents, formattedEvent]);
            setIsModalOpen(false);
            setSelectedSlot(null);
            setInitialEventData(null);
        } catch (error) {
            console.error('Error creating event:', error);
            setError(error.message || 'Failed to create event');
        }
    };

    const handleEventDelete = async (eventId) => {
        try {
            const token = await getAccessTokenSilently();
            const response = await fetch(getApiUrl(`schedule/${eventId}`), {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to delete event');
            }

            // Convert eventId to string for consistent comparison
            const eventIdStr = String(eventId);
            setEvents(prevEvents => prevEvents.filter(event => String(event.id) !== eventIdStr));
            setIsDetailsModalOpen(false);
            setSelectedEvent(null);
            setDeleteError(null);
        } catch (error) {
            console.error('Error deleting event:', error);
            setDeleteError('Failed to delete event');
        }
    };

    if (loading) {
        return <div className="container">Loading...</div>;
    }

    if (!isAuthenticated) {
        return (
            <div className="container">
                <p>Please login or signup to schedule appointments</p>
                <button onClick={() => loginWithRedirect({ appState: { returnTo: '/schedule' } })} className="btn">Log In</button>
            </div>
        );
    }

    return (
        <div className="container">
            {error && (
                <div role="alert" className="error-message">
                    {error}
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
                    selectable={true}
                    select={handleDateSelect}
                    events={events}
                    eventClick={handleEventClick}
                    datesSet={handleDatesSet}
                    headerToolbar={{
                        left: 'prev,next today',
                        center: 'title',
                        right: 'dayGridMonth,timeGridWeek,timeGridDay'
                    }}
                    slotMinTime="08:00:00"
                    slotMaxTime="20:00:00"
                    allDaySlot={false}
                    selectConstraint={{
                        startTime: '08:00',
                        endTime: '20:00',
                        daysOfWeek: [0, 1, 2, 3, 4, 5, 6]
                    }}
                    businessHours={{
                        daysOfWeek: [0, 1, 2, 3, 4, 5, 6],
                        startTime: '08:00',
                        endTime: '20:00'
                    }}
                    selectOverlap={false}
                    loading={loading}
                />
            </div>
            {isModalOpen && selectedSlot && (
                <EventModal
                    isOpen={isModalOpen}
                    onClose={() => {
                        setIsModalOpen(false);
                        setSelectedSlot(null);
                        setInitialEventData(null);
                    }}
                    onSubmit={handleEventSubmit}
                    startTime={selectedSlot.start}
                    endTime={selectedSlot.end}
                    initialData={initialEventData}
                />
            )}
            {isDetailsModalOpen && selectedEvent && (
                <EventDetailsModal
                    isOpen={isDetailsModalOpen}
                    onClose={() => {
                        setIsDetailsModalOpen(false);
                        setSelectedEvent(null);
                    }}
                    event={selectedEvent}
                    onDelete={handleEventDelete}
                    deleteError={deleteError}
                    isAdmin={isAdmin}
                    isUsersEvent={selectedEvent.user_email === user?.email}
                />
            )}
        </div>
    );
};

export default Schedule; 