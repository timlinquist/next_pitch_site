import React from 'react';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import CampDetailPage from './CampDetailPage';

const mockCamp = {
  id: 1,
  name: 'Summer Pitching Camp',
  description: 'desc',
  slug: 'summer-pitching-camp',
  start_date: '2026-06-10T00:00:00Z',
  end_date: '2026-06-14T00:00:00Z',
  price: 299,
  spots_remaining: 10,
  location: {
    id: 1,
    venue_name: 'Springfield HS',
    street: '123 Main St',
    city: 'Springfield',
    state: 'IL',
    zip: '62701',
    country: 'US',
  },
};

describe('CampDetailPage', () => {
  beforeEach(() => {
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: () => Promise.resolve(mockCamp) })
    );
  });

  it('renders camp info and location block', async () => {
    render(
      <MemoryRouter initialEntries={['/camps/summer-pitching-camp']}>
        <Routes>
          <Route path="/camps/:slug" element={<CampDetailPage />} />
        </Routes>
      </MemoryRouter>
    );
    await waitFor(() => expect(screen.getByText('Summer Pitching Camp')).toBeInTheDocument());
    expect(screen.getByText(/Springfield HS/)).toBeInTheDocument();
    expect(screen.getByText(/123 Main St, Springfield, IL 62701/)).toBeInTheDocument();
    const registerLink = screen.getByRole('link', { name: /Register Now/i });
    expect(registerLink.getAttribute('href')).toBe('/camps/summer-pitching-camp/register');
  });

  it('shows Location TBA when location missing', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: () => Promise.resolve({ ...mockCamp, location: null }) })
    );
    render(
      <MemoryRouter initialEntries={['/camps/summer-pitching-camp']}>
        <Routes>
          <Route path="/camps/:slug" element={<CampDetailPage />} />
        </Routes>
      </MemoryRouter>
    );
    await waitFor(() => expect(screen.getByText(/Location TBA/i)).toBeInTheDocument());
  });
});
