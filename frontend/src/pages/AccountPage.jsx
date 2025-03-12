import React, { useState, useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { Navigate } from 'react-router-dom';
import config from '../config';

const AccountPage = () => {
    const { isAuthenticated, user } = useAuth0();
    const [appointments, setAppointments] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchAppointments = async () => {
            if (!user?.email) return;
            
            try {
                const response = await fetch(`${config.apiUrl}/api/appointments/upcoming?email=${encodeURIComponent(user.email)}`);
                if (!response.ok) {
                    throw new Error('Failed to fetch appointments');
                }
                const data = await response.json();
                setAppointments(data);
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchAppointments();
    }, [user?.email]);

    if (!isAuthenticated) {
        return <Navigate to="/" />;
    }

    return (
        <div className="container">
            <div className="section">
                <h1>My Account</h1>
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
                    ) : appointments.length === 0 ? (
                        <p>No upcoming appointments</p>
                    ) : (
                        <ul className="appointment-list">
                            {appointments.map((appointment) => (
                                <li key={appointment.id} className="appointment-item">
                                    <p><strong>Date:</strong> {new Date(appointment.date).toLocaleDateString()}</p>
                                    <p><strong>Time:</strong> {new Date(appointment.date).toLocaleTimeString()}</p>
                                    <p><strong>Service:</strong> {appointment.service_type}</p>
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