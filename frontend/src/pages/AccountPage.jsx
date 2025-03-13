import React, { useState, useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { Navigate } from 'react-router-dom';
import config from '../config';
import { getApiUrl } from '../utils/api';

const AccountPage = () => {
    const { isAuthenticated, user, loginWithRedirect, logout } = useAuth0();
    const [appointments, setAppointments] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchUpcomingAppointments = async () => {
            if (!user) return;
            
            try {
                const response = await fetch(getApiUrl(`appointments/upcoming?email=${encodeURIComponent(user.email)}`), {
                    method: 'GET',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                });

                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }

                const data = await response.json();
                setAppointments(data || []);
            } catch (err) {
                console.error('Error fetching upcoming appointments:', err);
                setError('Failed to load appointments. Please try again later.');
                setAppointments([]);
            } finally {
                setLoading(false);
            }
        };

        fetchUpcomingAppointments();
    }, [user]);

    if (!isAuthenticated) {
        return (
            <div className="container">
                <div className="section" style={{ position: 'relative' }}>
                    <button 
                        className="btn"
                        style={{
                            position: 'absolute',
                            top: '1rem',
                            right: '1rem',
                        }}
                        onClick={() => loginWithRedirect({
                            appState: { returnTo: '/account' }
                        })}
                    >
                        Login
                    </button>
                    <h1>My Account</h1>
                    <p>Please log in to view your account information.</p>
                </div>
            </div>
        );
    }

    return (
        <div className="container">
            <div className="section" style={{ position: 'relative' }}>
                <button 
                    className="btn"
                    style={{
                        position: 'absolute',
                        top: '1rem',
                        right: '1rem',
                    }}
                    onClick={() => logout({  })}
                >
                    Logout
                </button>
                <h1 
                    style={{
                        marginTop: '1rem'
                    }}
                >My Account
                </h1>
                <div className="user-info">
                    <h2>Profile Information</h2>
                    <p><strong>Name:</strong> {user.name}</p>
                    <p><strong>Email:</strong> {user.email}</p>
                </div>

                <div className="appointments">
                    <h2>Upcoming Appointments</h2>
                    {loading ? (
                        <p>Loading appointments...</p>
                    ) : error ? (
                        <p className="error">{error}</p>
                    ) : !appointments || appointments.length === 0 ? (
                        <p>No upcoming appointments</p>
                    ) : (
                        <ul className="appointment-list">
                            {appointments.map((appointment) => (
                                <li key={appointment.id} className="appointment-item">
                                    <h3>{appointment.title}</h3>
                                    <p><strong>Date:</strong> {new Date(appointment.start_time).toLocaleDateString()}</p>
                                    <p><strong>Time:</strong> {new Date(appointment.start_time).toLocaleTimeString()} - {new Date(appointment.end_time).toLocaleTimeString()}</p>
                                    {appointment.description && (
                                        <p><strong>Description:</strong> {appointment.description}</p>
                                    )}
                                </li>
                            ))}
                        </ul>
                    )}
                </div>
            </div>
        </div>
    );
};

export default AccountPage; 