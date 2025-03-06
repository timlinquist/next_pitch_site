import React from 'react';
import FullCalendar from '@fullcalendar/react';
import dayGridPlugin from '@fullcalendar/daygrid';
import timeGridPlugin from '@fullcalendar/timegrid';
import interactionPlugin from '@fullcalendar/interaction';
import '../styles/calendar.css';

const Schedule = () => {
    const events = [
        {
            title: 'Available',
            start: '2024-03-06T10:00:00',
            end: '2024-03-06T11:00:00'
        },
        {
            title: 'Available',
            start: '2024-03-07T14:00:00',
            end: '2024-03-07T15:00:00'
        },
        {
            title: 'Available',
            start: '2024-03-08T11:00:00',
            end: '2024-03-08T12:00:00'
        }
    ];

    const handleEventClick = (info) => {
        alert('Would you like to schedule a consultation for ' + info.event.title + ' on ' + info.event.start.toLocaleString() + '?');
    };

    return (
        <div className="container">
            <div className="section">
                <h1>Schedule a Consultation</h1>
                <p>View our availability and schedule a consultation to discuss your pitch needs. We offer flexible scheduling options to accommodate your timeline.</p>
            </div>

            <div className="calendar-container">
                <FullCalendar
                    plugins={[dayGridPlugin, timeGridPlugin, interactionPlugin]}
                    initialView="dayGridMonth"
                    headerToolbar={{
                        left: 'prev,next today',
                        center: 'title',
                        right: 'dayGridMonth,timeGridWeek,timeGridDay'
                    }}
                    events={events}
                    eventClick={handleEventClick}
                    slotMinTime="09:00:00"
                    slotMaxTime="17:00:00"
                    allDaySlot={false}
                    weekends={false}
                    height="auto"
                    stickyHeaderDates={false}
                    expandRows={true}
                    handleWindowResize={true}
                />
            </div>
        </div>
    );
};

export default Schedule; 