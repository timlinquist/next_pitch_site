import { describe, it, expect } from 'vitest';
import { expandRecurringEvents } from './recurrence';

/**
 * Expands recurring events into their instances based on the current calendar view
 * @param {Array} events - Array of events from the API
 * @param {Date} viewStart - Start date of the current calendar view
 * @param {Date} viewEnd - End date of the current calendar view
 * @returns {Array} Array of all events including recurring instances
 */
describe('expandRecurringEvents', () => {
    // Helper function to create a date string for testing
    const createDateString = (year, month, day, hours = 0, minutes = 0) => {
        return new Date(year, month - 1, day, hours, minutes).toISOString();
    };

    // Helper function to create a test event
    const createTestEvent = (id, startTime, endTime, recurrence = 'none') => ({
        id,
        title: `Test Event ${id}`,
        start_time: startTime,
        end_time: endTime,
        description: 'Test Description',
        user_email: 'test@example.com',
        recurrence
    });

    it('should return empty array for invalid input', () => {
        expect(expandRecurringEvents(null, new Date(), new Date())).toEqual([]);
        expect(expandRecurringEvents(undefined, new Date(), new Date())).toEqual([]);
        expect(expandRecurringEvents([], new Date(), new Date())).toEqual([]);
    });

    it('should return original events if view dates are not set', () => {
        const events = [
            createTestEvent(1, createDateString(2024, 3, 15, 14, 0), createDateString(2024, 3, 15, 15, 0))
        ];
        expect(expandRecurringEvents(events, null, new Date())).toEqual(events);
        expect(expandRecurringEvents(events, new Date(), null)).toEqual(events);
    });

    it('should not expand non-recurring events', () => {
        const events = [
            createTestEvent(1, createDateString(2024, 3, 15, 14, 0), createDateString(2024, 3, 15, 15, 0))
        ];
        const viewStart = new Date(createDateString(2024, 3, 15));
        const viewEnd = new Date(createDateString(2024, 3, 22));
        
        const result = expandRecurringEvents(events, viewStart, viewEnd);
        expect(result).toHaveLength(1);
        expect(result[0]).toEqual(events[0]);
    });

    it('should expand weekly recurring events', () => {
        const events = [
            createTestEvent(
                1,
                createDateString(2024, 3, 15, 14, 0), // Friday at 2 PM
                createDateString(2024, 3, 15, 15, 0),
                'weekly'
            )
        ];
        const viewStart = new Date(createDateString(2024, 3, 15));
        const viewEnd = new Date(createDateString(2024, 3, 29)); // Two weeks

        const result = expandRecurringEvents(events, viewStart, viewEnd);
        
        // Should have original event plus one recurrence
        expect(result).toHaveLength(2);
        
        // Check original event
        expect(result[0]).toEqual(events[0]);
        
        // Check recurring instance
        expect(result[1]).toMatchObject({
            id: expect.stringContaining('1-'),
            title: events[0].title,
            recurrence: 'weekly',
            isRecurringInstance: true,
            originalEventId: 1
        });
        
        // Verify the recurring instance is on the next Friday
        const instanceDate = new Date(result[1].start_time);
        expect(instanceDate.getDay()).toBe(5); // Friday
        expect(instanceDate.getHours()).toBe(14);
        expect(instanceDate.getMinutes()).toBe(0);
    });

    it('should expand biweekly recurring events', () => {
        const events = [
            createTestEvent(
                1,
                createDateString(2024, 3, 15, 14, 0), // Friday at 2 PM
                createDateString(2024, 3, 15, 15, 0),
                'biweekly'
            )
        ];
        const viewStart = new Date(createDateString(2024, 3, 15));
        const viewEnd = new Date(createDateString(2024, 3, 29)); // Two weeks

        const result = expandRecurringEvents(events, viewStart, viewEnd);
        
        // Should have original event plus one recurrence (every other week)
        expect(result).toHaveLength(2);
        
        // Check original event
        expect(result[0]).toEqual(events[0]);
        
        // Check recurring instance
        expect(result[1]).toMatchObject({
            id: expect.stringContaining('1-'),
            title: events[0].title,
            recurrence: 'biweekly',
            isRecurringInstance: true,
            originalEventId: 1
        });
        
        // Verify the recurring instance is two weeks later
        const instanceDate = new Date(result[1].start_time);
        expect(instanceDate.getDay()).toBe(5); // Friday
        expect(instanceDate.getHours()).toBe(14);
        expect(instanceDate.getMinutes()).toBe(0);
        expect(instanceDate.getDate()).toBe(29); // Two weeks later
    });

    it('should expand monthly recurring events', () => {
        const events = [
            createTestEvent(
                1,
                createDateString(2024, 3, 15, 14, 0), // March 15 at 2 PM
                createDateString(2024, 3, 15, 15, 0),
                'monthly'
            )
        ];
        const viewStart = new Date(createDateString(2024, 3, 15));
        const viewEnd = new Date(createDateString(2024, 4, 15)); // One month

        const result = expandRecurringEvents(events, viewStart, viewEnd);
        
        // Should have original event plus one recurrence
        expect(result).toHaveLength(2);
        
        // Check original event
        expect(result[0]).toEqual(events[0]);
        
        // Check recurring instance
        expect(result[1]).toMatchObject({
            id: expect.stringContaining('1-'),
            title: events[0].title,
            recurrence: 'monthly',
            isRecurringInstance: true,
            originalEventId: 1
        });
        
        // Verify the recurring instance is on the same day next month
        const instanceDate = new Date(result[1].start_time);
        expect(instanceDate.getDate()).toBe(15);
        expect(instanceDate.getMonth()).toBe(3); // April (0-based)
        expect(instanceDate.getHours()).toBe(14);
        expect(instanceDate.getMinutes()).toBe(0);
    });

    it('should handle multiple recurring events', () => {
        const events = [
            createTestEvent(
                1,
                createDateString(2024, 3, 15, 14, 0),
                createDateString(2024, 3, 15, 15, 0),
                'weekly'
            ),
            createTestEvent(
                2,
                createDateString(2024, 3, 15, 16, 0),
                createDateString(2024, 3, 15, 17, 0),
                'monthly'
            )
        ];
        const viewStart = new Date(createDateString(2024, 3, 15));
        const viewEnd = new Date(createDateString(2024, 4, 15));

        const result = expandRecurringEvents(events, viewStart, viewEnd);
        
        // Should have 2 original events plus 1 weekly recurrence plus 1 monthly recurrence
        expect(result).toHaveLength(4);
        
        // Verify all events have unique IDs
        const ids = result.map(e => e.id);
        expect(new Set(ids).size).toBe(ids.length);
    });

    it('should not create instances outside the view range', () => {
        const events = [
            createTestEvent(
                1,
                createDateString(2024, 3, 15, 14, 0),
                createDateString(2024, 3, 15, 15, 0),
                'weekly'
            )
        ];
        const viewStart = new Date(createDateString(2024, 3, 15));
        const viewEnd = new Date(createDateString(2024, 3, 16)); // Only one day

        const result = expandRecurringEvents(events, viewStart, viewEnd);
        
        // Should only have the original event
        expect(result).toHaveLength(1);
        expect(result[0]).toEqual(events[0]);
    });

    it('should handle edge cases for monthly recurrence', () => {
        // Test with January 31st
        const events = [
            createTestEvent(
                1,
                createDateString(2024, 1, 31, 14, 0),
                createDateString(2024, 1, 31, 15, 0),
                'monthly'
            )
        ];
        const viewStart = new Date(createDateString(2024, 1, 31));
        const viewEnd = new Date(createDateString(2024, 3, 31));

        const result = expandRecurringEvents(events, viewStart, viewEnd);
        
        // Should have original event plus one recurrence (February 29th)
        expect(result).toHaveLength(2);
        
        // Check recurring instance
        const instanceDate = new Date(result[1].start_time);
        expect(instanceDate.getDate()).toBe(29); // February 29th (leap year)
        expect(instanceDate.getMonth()).toBe(1); // February (0-based)
    });
}); 