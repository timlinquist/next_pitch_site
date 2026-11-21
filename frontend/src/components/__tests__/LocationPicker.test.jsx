import React, { useState } from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import LocationPicker from '../LocationPicker';

const existing = [
  {
    id: 1,
    venue_name: 'Springfield HS',
    street: '123 Main St',
    city: 'Springfield',
    state: 'IL',
    zip: '62701',
    country: 'US',
  },
];

function Harness({ initial = {}, onChange }) {
  const [value, setValue] = useState(initial);
  return (
    <LocationPicker
      value={value}
      onChange={(v) => {
        setValue(v);
        onChange && onChange(v);
      }}
      existingLocations={existing}
    />
  );
}

describe('LocationPicker', () => {
  it('fills fields when selecting an existing location', () => {
    const onChange = vi.fn();
    render(<Harness onChange={onChange} />);
    const select = screen.getByLabelText(/Use existing location/i);
    fireEvent.change(select, { target: { value: '1' } });
    expect(screen.getByLabelText(/Street/i).value).toBe('123 Main St');
    expect(screen.getByLabelText(/City/i).value).toBe('Springfield');
    expect(screen.getByLabelText(/^State/i).value).toBe('IL');
    expect(screen.getByLabelText(/Zip/i).value).toBe('62701');
    expect(onChange).toHaveBeenCalled();
  });

  it('keeps fields editable after selecting', () => {
    render(<Harness />);
    fireEvent.change(screen.getByLabelText(/Use existing location/i), { target: { value: '1' } });
    const streetInput = screen.getByLabelText(/Street/i);
    fireEvent.change(streetInput, { target: { value: '456 Different Rd' } });
    expect(streetInput.value).toBe('456 Different Rd');
  });

  it('shows dedup hint when street+zip match existing', () => {
    render(<Harness />);
    fireEvent.change(screen.getByLabelText(/Street/i), { target: { value: '123 Main St' } });
    fireEvent.change(screen.getByLabelText(/Zip/i), { target: { value: '62701' } });
    expect(screen.getByText(/Matches existing location/i)).toBeInTheDocument();
  });

  it('emits merged object on field change', () => {
    const onChange = vi.fn();
    render(<Harness onChange={onChange} />);
    fireEvent.change(screen.getByLabelText(/Street/i), { target: { value: '9 Elm' } });
    const lastCall = onChange.mock.calls.at(-1)[0];
    expect(lastCall.street).toBe('9 Elm');
  });
});
