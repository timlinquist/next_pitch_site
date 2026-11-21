import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LocationMapEmbed from '../LocationMapEmbed';

describe('LocationMapEmbed', () => {
  const address = '123 Main St, Springfield, IL 62701';

  it('renders iframe with correct embed src', () => {
    render(<LocationMapEmbed address={address} />);
    const iframe = screen.getByTitle(/Map of 123 Main St/i);
    expect(iframe.tagName).toBe('IFRAME');
    expect(iframe.getAttribute('src')).toBe(
      `https://www.google.com/maps?q=${encodeURIComponent(address)}&output=embed`
    );
  });

  it('renders external Open in Google Maps link', () => {
    render(<LocationMapEmbed address={address} />);
    const link = screen.getByRole('link', { name: /Open in Google Maps/i });
    expect(link.getAttribute('href')).toBe(
      `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(address)}`
    );
    expect(link.getAttribute('target')).toBe('_blank');
    expect(link.getAttribute('rel')).toContain('noopener');
  });
});
