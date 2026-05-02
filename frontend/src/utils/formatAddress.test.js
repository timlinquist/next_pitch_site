import { describe, it, expect } from 'vitest';
import { formatAddress } from './formatAddress';

describe('formatAddress', () => {
  it('formats US address without country', () => {
    expect(formatAddress({
      street: '123 Main St',
      city: 'Springfield',
      state: 'IL',
      zip: '62701',
      country: 'US',
    })).toBe('123 Main St, Springfield, IL 62701');
  });

  it('appends country when not US', () => {
    expect(formatAddress({
      street: '1 Queen St',
      city: 'Toronto',
      state: 'ON',
      zip: 'M5H',
      country: 'Canada',
    })).toBe('1 Queen St, Toronto, ON M5H, Canada');
  });

  it('returns empty string when location is null', () => {
    expect(formatAddress(null)).toBe('');
  });
});
