/**
 * Expands recurring events into their instances based on the current calendar view
 * @param {Array} events - Array of events from the API
 * @param {Date} viewStart - Start date of the current calendar view
 * @param {Date} viewEnd - End date of the current calendar view
 * @returns {Array} Array of all events including recurring instances
 */
export const expandRecurringEvents = (events, viewStart, viewEnd) => {
    if (!events || !Array.isArray(events)) {
        return [];
    }

    if (!viewStart || !viewEnd) {
        return events;
    }

    // Convert view dates to UTC midnight
    const viewStartUTC = new Date(Date.UTC(
        viewStart.getUTCFullYear(),
        viewStart.getUTCMonth(),
        viewStart.getUTCDate(),
        0, 0, 0
    ));

    const viewEndUTC = new Date(Date.UTC(
        viewEnd.getUTCFullYear(),
        viewEnd.getUTCMonth(),
        viewEnd.getUTCDate(),
        23, 59, 59
    ));

    const expandedEvents = [];

    events.forEach(event => {
        // Always add the original event
        expandedEvents.push(event);

        // Skip if no recurrence or recurrence is 'none'
        if (!event.recurrence || event.recurrence === 'none' || !event.start_time) {
            return;
        }

        console.log('Processing event:', {
            id: event.id,
            recurrence: event.recurrence,
            start_time: event.start_time
        });

        const eventStart = new Date(event.start_time);
        const eventEnd = new Date(event.end_time);
        const duration = eventEnd - eventStart;

        // Store original hours and minutes in UTC
        const originalHours = eventStart.getUTCHours();
        const originalMinutes = eventStart.getUTCMinutes();
        const originalDate = eventStart.getUTCDate();

        // Create first instance based on recurrence type
        let instanceStart = new Date(Date.UTC(
            eventStart.getUTCFullYear(),
            eventStart.getUTCMonth(),
            eventStart.getUTCDate(),
            originalHours,
            originalMinutes
        ));

        switch (event.recurrence) {
            case 'weekly':
                instanceStart.setUTCDate(instanceStart.getUTCDate() + 7);
                break;
            case 'biweekly':
                instanceStart.setUTCDate(instanceStart.getUTCDate() + 14);
                break;
            case 'monthly':
                // For monthly recurrence, we need to handle month length differences
                // For example, if the original event is on January 31st, the next instance
                // should be on February 29th (in leap years) or February 28th (in non-leap years)
                const originalDate = new Date(event.start_time);
                
                // Create a new date for the next month while preserving the time
                instanceStart = new Date(Date.UTC(
                    originalDate.getUTCFullYear(),
                    originalDate.getUTCMonth() + 1,
                    1,  // Start with the 1st of the month
                    originalDate.getUTCHours(),
                    originalDate.getUTCMinutes()
                ));
                
                // Get the last day of the target month
                const lastDayOfMonth = new Date(Date.UTC(
                    instanceStart.getUTCFullYear(),
                    instanceStart.getUTCMonth() + 1,
                    0
                )).getUTCDate();
                
                // Use either the original day or the last day of the month, whichever is smaller
                const targetDay = Math.min(originalDate.getUTCDate(), lastDayOfMonth);
                instanceStart.setUTCDate(targetDay);
                break;
        }

        console.log('Instance details:', {
            instanceStart,
            inRange: instanceStart >= viewStartUTC && instanceStart <= viewEndUTC
        });

        // Only add the instance if it starts within the view range
        // Compare dates at midnight to ensure consistent behavior
        const instanceStartDay = new Date(Date.UTC(
            instanceStart.getUTCFullYear(),
            instanceStart.getUTCMonth(),
            instanceStart.getUTCDate(),
            0, 0, 0
        ));

        if (instanceStartDay >= viewStartUTC && instanceStartDay <= viewEndUTC) {
            const instanceEnd = new Date(instanceStart.getTime() + duration);
            expandedEvents.push({
                ...event,
                id: `${event.id}-${instanceStart.getTime()}`,
                start_time: instanceStart.toISOString(),
                end_time: instanceEnd.toISOString(),
                isRecurringInstance: true,
                originalEventId: event.id
            });
        }
    });

    console.log('Final expanded events:', expandedEvents);
    return expandedEvents;
}; 