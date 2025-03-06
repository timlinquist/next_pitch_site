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
                
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                
                const data = await response.json();
                console.log('Received data:', data);
                
                const formattedEvents = data.map(entry => ({
                    id: entry.id,
                    title: entry.title,
                    start: entry.start_time,
                    end: entry.end_time,
                    description: entry.description
                }));
                
                setEvents(formattedEvents);
                setLoading(false);
            } catch (err) {
                console.error('Error fetching schedule:', err);
                setError(err.message);
                setLoading(false);
            }
        };

        fetchScheduleEntries();
    }, []);

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
                    eventClick={(info) => {
                        alert(`Event: ${info.event.title}\nDescription: ${info.event.extendedProps.description}`);
                    }}
                    timeZone="local"
                    initialDate={new Date()}
                />
            </div>
        </div>
    );
};

export default Schedule; 