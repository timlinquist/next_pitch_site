import React, { useState, useEffect } from 'react';
import FullCalendar from '@fullcalendar/react';
import dayGridPlugin from '@fullcalendar/daygrid';
import timeGridPlugin from '@fullcalendar/timegrid';
import interactionPlugin from '@fullcalendar/interaction';
import '../styles/calendar.css';

const Schedule = () => {
    const [events, setEvents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchScheduleEntries = async () => {
            try {
                console.log('Fetching schedule entries...');
                const response = await fetch('http://localhost:8080/api/schedule', {
                    method: 'GET',
                    headers: {
                        'Accept': 'application/json',
                        'Content-Type': 'application/json',
                    },
                });
                
                console.log('Response status:', response.status);
                console.log('Response headers:', Object.fromEntries(response.headers.entries()));
                
                if (!response.ok) {
                    const errorText = await response.text();
                    console.error('Error response:', errorText);
                    throw new Error(`Failed to fetch schedule entries: ${response.status} - ${errorText}`);
                }
                
                const data = await response.json();
                console.log('Fetched data:', JSON.stringify(data, null, 2));
                
                if (!Array.isArray(data)) {
                    throw new Error('Received invalid data format from server');
                }
                
                // Transform the data to match FullCalendar's event format
                const calendarEvents = data.map(entry => {
                    try {
                        // Parse the dates and adjust for timezone
                        const startDate = new Date(entry.start_time);
                        const endDate = new Date(entry.end_time);
                        
                        // Log the parsed dates for debugging
                        console.log('Event:', entry.title);
                        console.log('Raw start:', entry.start_time);
                        console.log('Parsed start:', startDate.toISOString());
                        console.log('Raw end:', entry.end_time);
                        console.log('Parsed end:', endDate.toISOString());
                        
                        return {
                            id: entry.id,
                            title: entry.title,
                            description: entry.description,
                            start: startDate.toISOString(),
                            end: endDate.toISOString(),
                            backgroundColor: '#4CAF50',
                            borderColor: '#4CAF50',
                            extendedProps: {
                                description: entry.description
                            }
                        };
                    } catch (err) {
                        console.error('Error processing event:', entry, err);
                        return null;
                    }
                }).filter(Boolean); // Remove any null entries from failed parsing
                
                console.log('Transformed events:', JSON.stringify(calendarEvents, null, 2));
                setEvents(calendarEvents);
            } catch (err) {
                console.error('Error fetching schedule:', err);
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchScheduleEntries();
    }, []);

    const handleEventClick = (info) => {
        const event = info.event;
        alert(`
Event: ${event.title}
Time: ${event.start.toLocaleString()} - ${event.end.toLocaleString()}
Description: ${event.extendedProps.description || 'No description available'}
        `);
    };

    const handleEventsSet = (events) => {
        console.log('FullCalendar received events:', events);
    };

    if (loading) {
        return <div className="container">Loading schedule...</div>;
    }

    if (error) {
        return <div className="container">Error: {error}</div>;
    }

    return (
        <div className="container">
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
                    eventClick={handleEventClick}
                    eventsSet={handleEventsSet}
                    slotMinTime="09:00:00"
                    slotMaxTime="17:00:00"
                    allDaySlot={false}
                    weekends={false}
                    height="auto"
                    stickyHeaderDates={false}
                    expandRows={true}
                    handleWindowResize={true}
                    timeZone="local"
                    displayEventTime={true}
                    displayEventEnd={true}
                    eventTimeFormat={{
                        hour: '2-digit',
                        minute: '2-digit',
                        meridiem: false,
                        hour12: false
                    }}
                    initialDate={new Date()}
                />
            </div>
        </div>
    );
};

export default Schedule; 