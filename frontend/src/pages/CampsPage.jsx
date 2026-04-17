import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getApiUrl } from '../utils/api';
import '../styles/camps.css';

const CampsPage = () => {
    const [camps, setCamps] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchCamps = async () => {
            try {
                const response = await fetch(getApiUrl('camps'));
                if (!response.ok) {
                    throw new Error('Failed to fetch camps');
                }
                const data = await response.json();
                setCamps(data || []);
            } catch (err) {
                setError('Failed to load camps. Please try again later.');
            } finally {
                setLoading(false);
            }
        };

        fetchCamps();
    }, []);

    const formatDate = (dateStr) => {
        return new Date(dateStr).toLocaleDateString('en-US', {
            month: 'long',
            day: 'numeric',
            year: 'numeric'
        });
    };

    const formatPrice = (cents) => {
        return `$${(cents / 100).toFixed(2)}`;
    };

    if (loading) {
        return (
            <div className="container">
                <div className="section">
                    <h1>Upcoming Camps</h1>
                    <p>Loading camps...</p>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="container">
                <div className="section">
                    <h1>Upcoming Camps</h1>
                    <p className="error">{error}</p>
                </div>
            </div>
        );
    }

    return (
        <div className="container">
            <div className="section">
                <h1>Upcoming Camps</h1>
                <p>Register your athlete for one of our upcoming camps. Space is limited!</p>
            </div>

            {camps.length === 0 ? (
                <div className="section">
                    <p>No camps are currently scheduled. Check back soon!</p>
                </div>
            ) : (
                <div className="services-grid">
                    {camps.map((camp) => (
                        <div key={camp.id} className="service-card">
                            <h3>{camp.name}</h3>
                            <div className="price">{formatPrice(camp.price_cents)}</div>
                            <p className="duration">
                                {formatDate(camp.start_date)} - {formatDate(camp.end_date)}
                            </p>
                            <p className="description">{camp.description}</p>
                            {camp.spots_remaining !== null && camp.spots_remaining !== undefined && (
                                <p className="camp-spots">
                                    {camp.spots_remaining > 0
                                        ? `${camp.spots_remaining} spot${camp.spots_remaining !== 1 ? 's' : ''} remaining`
                                        : 'Full'}
                                </p>
                            )}
                            {camp.spots_remaining === 0 ? (
                                <button className="btn" disabled>Full</button>
                            ) : (
                                <Link to={`/camps/${camp.id}/register`} className="btn">
                                    Register Now
                                </Link>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default CampsPage;
