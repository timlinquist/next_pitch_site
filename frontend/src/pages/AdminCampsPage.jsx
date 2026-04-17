import React, { useState, useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { getApiUrl } from '../utils/api';
import '../styles/camps.css';

const AdminCampsPage = () => {
    const { getAccessTokenSilently, user } = useAuth0();
    const [camps, setCamps] = useState([]);
    const [registrations, setRegistrations] = useState({});
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [isAdmin, setIsAdmin] = useState(false);
    const [showForm, setShowForm] = useState(false);
    const [editingCamp, setEditingCamp] = useState(null);
    const [expandedCamp, setExpandedCamp] = useState(null);
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        start_date: '',
        end_date: '',
        price_cents: '',
        max_capacity: '',
    });

    useEffect(() => {
        checkAdminAndFetch();
    }, []);

    const checkAdminAndFetch = async () => {
        try {
            const token = await getAccessTokenSilently();
            const userRes = await fetch(getApiUrl('users/me'), {
                headers: { Authorization: `Bearer ${token}` },
            });
            const userData = await userRes.json();
            if (!userData.is_admin) {
                setIsAdmin(false);
                setLoading(false);
                return;
            }
            setIsAdmin(true);
            await fetchCamps();
        } catch (err) {
            setError('Failed to verify admin status.');
        } finally {
            setLoading(false);
        }
    };

    const fetchCamps = async () => {
        try {
            const response = await fetch(getApiUrl('camps'));
            if (!response.ok) throw new Error('Failed to fetch camps');
            const data = await response.json();
            setCamps(data || []);
        } catch (err) {
            setError('Failed to load camps.');
        }
    };

    const fetchRegistrations = async (campId) => {
        try {
            const token = await getAccessTokenSilently();
            const response = await fetch(getApiUrl(`camps/${campId}/registrations`), {
                headers: { Authorization: `Bearer ${token}` },
            });
            if (!response.ok) throw new Error('Failed to fetch registrations');
            const data = await response.json();
            setRegistrations(prev => ({ ...prev, [campId]: data || [] }));
        } catch (err) {
            setError('Failed to load registrations.');
        }
    };

    const handleFormChange = (e) => {
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    const resetForm = () => {
        setFormData({ name: '', description: '', start_date: '', end_date: '', price_cents: '', max_capacity: '' });
        setEditingCamp(null);
        setShowForm(false);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            const token = await getAccessTokenSilently();
            const payload = {
                name: formData.name,
                description: formData.description,
                start_date: formData.start_date + 'T00:00:00Z',
                end_date: formData.end_date + 'T00:00:00Z',
                price_cents: parseInt(formData.price_cents),
                max_capacity: formData.max_capacity ? parseInt(formData.max_capacity) : null,
            };

            const url = editingCamp
                ? getApiUrl(`camps/${editingCamp.id}`)
                : getApiUrl('camps');

            const response = await fetch(url, {
                method: editingCamp ? 'PUT' : 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Failed to save camp');
            }

            resetForm();
            await fetchCamps();
        } catch (err) {
            setError(err.message);
        }
    };

    const handleEdit = (camp) => {
        setFormData({
            name: camp.name,
            description: camp.description || '',
            start_date: camp.start_date.split('T')[0],
            end_date: camp.end_date.split('T')[0],
            price_cents: String(camp.price_cents),
            max_capacity: camp.max_capacity ? String(camp.max_capacity) : '',
        });
        setEditingCamp(camp);
        setShowForm(true);
    };

    const handleDeactivate = async (campId) => {
        try {
            const token = await getAccessTokenSilently();
            const response = await fetch(getApiUrl(`camps/${campId}`), {
                method: 'DELETE',
                headers: { Authorization: `Bearer ${token}` },
            });
            if (!response.ok) throw new Error('Failed to deactivate camp');
            await fetchCamps();
        } catch (err) {
            setError(err.message);
        }
    };

    const toggleRegistrations = async (campId) => {
        if (expandedCamp === campId) {
            setExpandedCamp(null);
            return;
        }
        setExpandedCamp(campId);
        if (!registrations[campId]) {
            await fetchRegistrations(campId);
        }
    };

    if (loading) {
        return (
            <div className="container">
                <div className="section"><p>Loading...</p></div>
            </div>
        );
    }

    if (!isAdmin) {
        return (
            <div className="container">
                <div className="section">
                    <h1>Access Denied</h1>
                    <p>You do not have admin access.</p>
                </div>
            </div>
        );
    }

    return (
        <div className="container">
            <div className="section">
                <h1>Manage Camps</h1>
                {error && <p className="error">{error}</p>}

                <button className="btn" onClick={() => { resetForm(); setShowForm(!showForm); }} style={{ marginBottom: '1rem' }}>
                    {showForm ? 'Cancel' : 'Create New Camp'}
                </button>

                {showForm && (
                    <form onSubmit={handleSubmit} className="admin-camp-form">
                        <h3>{editingCamp ? 'Edit Camp' : 'New Camp'}</h3>
                        <div className="form-group">
                            <label htmlFor="camp-name">Name</label>
                            <input type="text" id="camp-name" name="name" value={formData.name} onChange={handleFormChange} required />
                        </div>
                        <div className="form-group">
                            <label htmlFor="camp-desc">Description</label>
                            <textarea id="camp-desc" name="description" value={formData.description} onChange={handleFormChange} />
                        </div>
                        <div className="form-group">
                            <label htmlFor="camp-start">Start Date</label>
                            <input type="date" id="camp-start" name="start_date" value={formData.start_date} onChange={handleFormChange} required />
                        </div>
                        <div className="form-group">
                            <label htmlFor="camp-end">End Date</label>
                            <input type="date" id="camp-end" name="end_date" value={formData.end_date} onChange={handleFormChange} required />
                        </div>
                        <div className="form-group">
                            <label htmlFor="camp-price">Price (cents)</label>
                            <input type="number" id="camp-price" name="price_cents" value={formData.price_cents} onChange={handleFormChange} required min="0" />
                        </div>
                        <div className="form-group">
                            <label htmlFor="camp-cap">Max Capacity</label>
                            <input type="number" id="camp-cap" name="max_capacity" value={formData.max_capacity} onChange={handleFormChange} min="1" placeholder="Leave blank for unlimited" />
                        </div>
                        <div className="modal-actions">
                            <button type="button" className="btn btn-secondary" onClick={resetForm}>Cancel</button>
                            <button type="submit" className="btn">{editingCamp ? 'Update' : 'Create'}</button>
                        </div>
                    </form>
                )}
            </div>

            <div className="section">
                <h2>All Camps</h2>
                {camps.length === 0 ? (
                    <p>No camps created yet.</p>
                ) : (
                    <div className="admin-camps-list">
                        {camps.map((camp) => (
                            <div key={camp.id} className={`admin-camp-item ${!camp.is_active ? 'inactive' : ''}`}>
                                <div className="admin-camp-header">
                                    <div>
                                        <h3>{camp.name} {!camp.is_active && <span className="badge-inactive">Inactive</span>}</h3>
                                        <p>
                                            {new Date(camp.start_date).toLocaleDateString()} - {new Date(camp.end_date).toLocaleDateString()}
                                            {' | '}${(camp.price_cents / 100).toFixed(2)}
                                            {camp.max_capacity && ` | Capacity: ${camp.registered_count}/${camp.max_capacity}`}
                                        </p>
                                    </div>
                                    <div className="admin-camp-actions">
                                        <button className="btn btn-small" onClick={() => toggleRegistrations(camp.id)}>
                                            {expandedCamp === camp.id ? 'Hide' : 'Registrations'}
                                        </button>
                                        {camp.is_active && (
                                            <>
                                                <button className="btn btn-small btn-secondary" onClick={() => handleEdit(camp)}>Edit</button>
                                                <button className="btn btn-small btn-danger" onClick={() => handleDeactivate(camp.id)}>Deactivate</button>
                                            </>
                                        )}
                                    </div>
                                </div>

                                {expandedCamp === camp.id && (
                                    <div className="registrations-panel">
                                        {!registrations[camp.id] ? (
                                            <p>Loading registrations...</p>
                                        ) : registrations[camp.id].length === 0 ? (
                                            <p>No registrations yet.</p>
                                        ) : (
                                            <table className="registrations-table">
                                                <thead>
                                                    <tr>
                                                        <th>Athlete</th>
                                                        <th>Age</th>
                                                        <th>Position</th>
                                                        <th>Parent Email</th>
                                                        <th>Phone</th>
                                                        <th>Status</th>
                                                        <th>Amount</th>
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    {registrations[camp.id].map((r) => (
                                                        <tr key={r.registration.id}>
                                                            <td>{r.athlete.name}</td>
                                                            <td>{r.athlete.age}</td>
                                                            <td>{r.athlete.position || '-'}</td>
                                                            <td>{r.athlete.parent_email}</td>
                                                            <td>{r.athlete.parent_phone || '-'}</td>
                                                            <td>
                                                                <span className={`badge-status badge-${r.registration.payment_status}`}>
                                                                    {r.registration.payment_status}
                                                                </span>
                                                            </td>
                                                            <td>${(r.registration.amount_cents / 100).toFixed(2)}</td>
                                                        </tr>
                                                    ))}
                                                </tbody>
                                            </table>
                                        )}
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};

export default AdminCampsPage;
