/**
 * Expands recurring events into their instances based on the current calendar view
 * @param {Array} events - Array of events from the API
 * @param {Date} viewStart - Start date of the current calendar view
 * @param {Date} viewEnd - End date of the current calendar view
 * @returns {Array} Array of all events including recurring instances
 */
export const expandRecurringEvents = (events, viewStart, viewEnd) => {
    // Handle invalid inputs
    if (!events?.length || !viewStart || !viewEnd) {
        return events || [];
    }

    // Convert view dates to UTC midnight for consistent comparison
    const viewStartUTC = new Date(Date.UTC(
        viewStart.getUTCFullYear(),
        viewStart.getUTCMonth(),
        viewStart.getUTCDate()
    ));
    const viewEndUTC = new Date(Date.UTC(
        viewEnd.getUTCFullYear(),
        viewEnd.getUTCMonth(),
        viewEnd.getUTCDate(),
        23, 59, 59
    ));

    // Map of recurrence types to their offset calculators
    const recurrenceOffsets = {
        weekly: (date) => ({ days: 7 }),
        biweekly: (date) => ({ days: 14 }),
        monthly: (date) => ({ months: 1 })
    };

    return events.flatMap(event => {
        // Return non-recurring events as is
        if (!event.recurrence || event.recurrence === 'none' || !event.start_time) {
            return [event];
        }

        const eventStart = new Date(event.start_time);
        const offsetCalculator = recurrenceOffsets[event.recurrence];
        
        if (!offsetCalculator) {
            return [event];
        }

        // Calculate the next instance
        const offset = offsetCalculator(eventStart);
        const nextInstance = new Date(eventStart);
        
        if (offset.days) {
            nextInstance.setUTCDate(nextInstance.getUTCDate() + offset.days);
        } else if (offset.months) {
            // Handle month transitions properly
            const originalDay = nextInstance.getUTCDate();
            
            // Store original date components
            const originalMonth = nextInstance.getUTCMonth();
            const originalYear = nextInstance.getUTCFullYear();
            
            // Get last day of target month
            const lastDayOfTargetMonth = new Date(Date.UTC(
                originalYear,
                originalMonth + offset.months + 1,
                0
            )).getUTCDate();

            // Set to first day of month to avoid JavaScript's auto-adjustment
            nextInstance.setUTCDate(1);
            nextInstance.setUTCMonth(originalMonth + offset.months);
            
            // For dates beyond the last day of the target month,
            // use the last day of that month (e.g. Jan 31 -> Feb 29 in leap year)
            nextInstance.setUTCDate(Math.min(originalDay, lastDayOfTargetMonth));
        }

        // Only add the next instance if it falls within the view range
        const nextInstanceStartUTC = new Date(Date.UTC(
            nextInstance.getUTCFullYear(),
            nextInstance.getUTCMonth(),
            nextInstance.getUTCDate(),
            nextInstance.getUTCHours(),
            nextInstance.getUTCMinutes()
        ));

        if (nextInstanceStartUTC >= viewStartUTC && nextInstanceStartUTC <= viewEndUTC) {
            const duration = new Date(event.end_time) - new Date(event.start_time);
            const nextInstanceEnd = new Date(nextInstance.getTime() + duration);
            
            return [
                event,
                {
                    ...event,
                    instanceKey: `${event.id}_${nextInstance.toISOString()}`,
                    isRecurringInstance: true,
                    originalEventId: event.id,
                    start_time: nextInstance.toISOString(),
                    end_time: nextInstanceEnd.toISOString()
                }
            ];
        }

        return [event];
    });
}; 