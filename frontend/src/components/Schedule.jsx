import React, { useState, useEffect } from 'react';
import FullCalendar from '@fullcalendar/react';
import dayGridPlugin from '@fullcalendar/daygrid';
import timeGridPlugin from '@fullcalendar/timegrid';
import interactionPlugin from '@fullcalendar/interaction';
import EventModal from './EventModal';
import '../styles/calendar.css';

const Schedule = () => {
    const [events, setEvents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedSlot, setSelectedSlot] = useState(null);

    useEffect(() => {
        fetchScheduleEntries();
    }, []);

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

    const handleDateSelect = (selectInfo) => {
        setSelectedSlot({
            start: selectInfo.start,
            end: selectInfo.end
        });
        setIsModalOpen(true);
    };

    const handleEventSubmit = async (eventData) => {
        try {
            const response = await fetch('http://localhost:8080/api/schedule', {
                method: 'POST',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(eventData)
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const newEvent = await response.json();
            const formattedEvent = {
                id: newEvent.id,
                title: newEvent.title,
                start: newEvent.start_time,
                end: newEvent.end_time,
                description: newEvent.description
            };

            setEvents([...events, formattedEvent]);
        } catch (err) {
            console.error('Error creating event:', err);
            alert('Failed to create event. Please try again.');
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
                    eventClick={(info) => {
                        alert(`Event: ${info.event.title}\nDescription: ${info.event.extendedProps.description}`);
                    }}
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
            />
        </div>
    );
};

export default Schedule; 