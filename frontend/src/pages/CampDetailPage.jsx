import React, { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { getApiUrl } from '../utils/api';
import { formatAddress } from '../utils/formatAddress';
import LocationMapEmbed from '../components/LocationMapEmbed';
import '../styles/camps.css';

const formatDate = (dateStr) =>
  new Date(dateStr).toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' });

const formatPrice = (dollars) => `$${Number(dollars).toFixed(2)}`;

const CampDetailPage = () => {
  const { slug } = useParams();
  const [camp, setCamp] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchCamp = async () => {
      try {
        const res = await fetch(getApiUrl(`camps/by-slug/${slug}`));
        if (!res.ok) throw new Error('not found');
        setCamp(await res.json());
      } catch (e) {
        setError('Camp not found.');
      } finally {
        setLoading(false);
      }
    };
    fetchCamp();
  }, [slug]);

  if (loading) {
    return <div className="container"><div className="section"><p>Loading...</p></div></div>;
  }
  if (error || !camp) {
    return <div className="container"><div className="section"><p className="error">{error || 'Not found'}</p></div></div>;
  }

  const isFull = camp.age_groups && camp.age_groups.length > 0
    ? camp.age_groups.every((g) => g.spots_remaining <= 0)
    : camp.spots_remaining === 0;

  const address = formatAddress(camp.location);

  return (
    <div className="container">
      <div className="section">
        <h1>{camp.name}</h1>
        <p className="duration">
          {formatDate(camp.start_date)} - {formatDate(camp.end_date)}
        </p>
        {camp.age_groups && camp.age_groups.length > 0 ? (
          <div className="age-group-spots">
            {camp.age_groups.map((g, i) => (
              <p key={i} className="camp-spots">
                Ages {g.min_age}-{g.max_age}: {formatPrice(g.price)}
                {' — '}
                {g.spots_remaining > 0
                  ? `${g.spots_remaining} spot${g.spots_remaining !== 1 ? 's' : ''} remaining`
                  : 'Full'}
              </p>
            ))}
          </div>
        ) : (
          <>
            {camp.price && <div className="price">{formatPrice(camp.price)}</div>}
            {camp.spots_remaining !== null && camp.spots_remaining !== undefined && (
              <p className="camp-spots">
                {camp.spots_remaining > 0
                  ? `${camp.spots_remaining} spot${camp.spots_remaining !== 1 ? 's' : ''} remaining`
                  : 'Full'}
              </p>
            )}
          </>
        )}
        <p className="description">{camp.description}</p>
      </div>

      <div className="section">
        <h2>Location</h2>
        {camp.location ? (
          <div className="camp-location">
            {camp.location.venue_name && <p className="venue-name">{camp.location.venue_name}</p>}
            <p className="address">{address}</p>
            <LocationMapEmbed address={address} />
          </div>
        ) : (
          <p>Location TBA</p>
        )}
      </div>

      <div className="section">
        {isFull ? (
          <button className="btn" disabled>Full</button>
        ) : (
          <Link to={`/camps/${camp.slug}/register`} className="btn">Register Now</Link>
        )}
      </div>
    </div>
  );
};

export default CampDetailPage;
