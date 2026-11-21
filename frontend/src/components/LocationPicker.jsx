import React from 'react';

const EMPTY = {
  venue_name: '',
  street: '',
  city: '',
  state: '',
  zip: '',
  country: 'US',
};

const LocationPicker = ({ value, onChange, existingLocations = [] }) => {
  const current = { ...EMPTY, ...(value || {}) };

  const update = (patch) => {
    onChange({ ...current, ...patch });
  };

  const handleSelectExisting = (e) => {
    const id = e.target.value;
    if (!id) return;
    const match = existingLocations.find((l) => String(l.id) === id);
    if (!match) return;
    onChange({
      venue_name: match.venue_name || '',
      street: match.street,
      city: match.city,
      state: match.state,
      zip: match.zip,
      country: match.country || 'US',
    });
  };

  const matchesExisting = existingLocations.some(
    (l) => l.street === current.street && l.zip === current.zip && current.street && current.zip
  );

  return (
    <div className="location-picker">
      <div className="form-group">
        <label htmlFor="location-existing">Use existing location (optional)</label>
        <select id="location-existing" onChange={handleSelectExisting} defaultValue="">
          <option value="">— Enter new address below —</option>
          {existingLocations.map((l) => (
            <option key={l.id} value={l.id}>
              {(l.venue_name || l.street)} — {l.street}, {l.city} {l.state} {l.zip}
            </option>
          ))}
        </select>
        <small>Selecting fills fields below. Edits create new entry unless street+zip match.</small>
      </div>

      <div className="form-group">
        <label htmlFor="location-venue">Venue name (optional)</label>
        <input
          id="location-venue"
          type="text"
          value={current.venue_name || ''}
          onChange={(e) => update({ venue_name: e.target.value })}
        />
      </div>

      <div className="form-group">
        <label htmlFor="location-street">Street</label>
        <input
          id="location-street"
          type="text"
          value={current.street}
          onChange={(e) => update({ street: e.target.value })}
          required
        />
      </div>

      <div className="form-group location-row">
        <div>
          <label htmlFor="location-city">City</label>
          <input
            id="location-city"
            type="text"
            value={current.city}
            onChange={(e) => update({ city: e.target.value })}
            required
          />
        </div>
        <div>
          <label htmlFor="location-state">State</label>
          <input
            id="location-state"
            type="text"
            value={current.state}
            onChange={(e) => update({ state: e.target.value })}
            required
          />
        </div>
        <div>
          <label htmlFor="location-zip">Zip</label>
          <input
            id="location-zip"
            type="text"
            value={current.zip}
            onChange={(e) => update({ zip: e.target.value })}
            required
          />
        </div>
      </div>

      <div className="form-group">
        <label htmlFor="location-country">Country</label>
        <input
          id="location-country"
          type="text"
          value={current.country || 'US'}
          onChange={(e) => update({ country: e.target.value })}
        />
      </div>

      {matchesExisting && (
        <div className="location-dedup-hint">
          ✓ Matches existing location — will reuse on save
        </div>
      )}
    </div>
  );
};

export default LocationPicker;
