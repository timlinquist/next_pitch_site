import React, { useState, useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { getApiUrl } from '../utils/api';
import '../styles/camps.css';

const generateSlug = (name) => {
    return name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
};

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
        price: '',
        max_capacity: '',
        slug: '',
        capacity_mode: 'simple',
        age_groups: [],
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
        const { name, value } = e.target;
        const updated = { ...formData, [name]: value };

        if (name === 'name' && !editingCamp) {
            updated.slug = generateSlug(value);
        }

        setFormData(updated);
    };

    const handleAgeGroupChange = (index, field, value) => {
        const updated = [...formData.age_groups];
        updated[index] = { ...updated[index], [field]: value };
        setFormData({ ...formData, age_groups: updated });
    };

    const addAgeGroup = () => {
        setFormData({
            ...formData,
            age_groups: [...formData.age_groups, { min_age: '', max_age: '', max_capacity: '', price: '' }],
        });
    };

    const removeAgeGroup = (index) => {
        const updated = formData.age_groups.filter((_, i) => i !== index);
        setFormData({ ...formData, age_groups: updated });
    };

    const validateAgeGroups = () => {
        const groups = formData.age_groups
            .map(g => ({ min_age: parseInt(g.min_age), max_age: parseInt(g.max_age), max_capacity: parseInt(g.max_capacity), price: parseFloat(g.price) }))
            .sort((a, b) => a.min_age - b.min_age);

        for (let i = 0; i < groups.length; i++) {
            if (isNaN(groups[i].min_age) || isNaN(groups[i].max_age) || isNaN(groups[i].max_capacity) || isNaN(groups[i].price)) {
                return 'All age group fields are required';
            }
            if (groups[i].price <= 0) {
                return 'Price must be greater than 0';
            }
            if (groups[i].min_age > groups[i].max_age) {
                return 'Min age cannot be greater than max age';
            }
            if (groups[i].max_capacity <= 0) {
                return 'Capacity must be greater than 0';
            }
            if (i > 0 && groups[i].min_age <= groups[i - 1].max_age) {
                return 'Age groups must not overlap';
            }
        }
        return null;
    };

    const resetForm = () => {
        setFormData({ name: '', description: '', start_date: '', end_date: '', price: '', max_capacity: '', slug: '', capacity_mode: 'simple', age_groups: [] });
        setEditingCamp(null);
        setShowForm(false);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError(null);

        if (formData.capacity_mode === 'age_range') {
            if (formData.age_groups.length === 0) {
                setError('At least one age group is required in age range mode');
                return;
            }
            const validationError = validateAgeGroups();
            if (validationError) {
                setError(validationError);
                return;
            }
        }

        try {
            const token = await getAccessTokenSilently();
            const payload = {
                name: formData.name,
                description: formData.description,
                start_date: formData.start_date + 'T00:00:00Z',
                end_date: formData.end_date + 'T00:00:00Z',
                price: formData.capacity_mode === 'simple' ? parseFloat(formData.price) : null,
                slug: formData.slug || null,
                max_capacity: formData.capacity_mode === 'simple' && formData.max_capacity
                    ? parseInt(formData.max_capacity)
                    : null,
                age_groups: formData.capacity_mode === 'age_range'
                    ? formData.age_groups.map(g => ({
                        min_age: parseInt(g.min_age),
                        max_age: parseInt(g.max_age),
                        max_capacity: parseInt(g.max_capacity),
                        price: parseFloat(g.price),
                    }))
                    : [],
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
        const hasAgeGroups = camp.age_groups && camp.age_groups.length > 0;
        setFormData({
            name: camp.name,
            description: camp.description || '',
            start_date: camp.start_date.split('T')[0],
            end_date: camp.end_date.split('T')[0],
            price: camp.price ? String(camp.price) : '',
            max_capacity: camp.max_capacity ? String(camp.max_capacity) : '',
            slug: camp.slug || '',
            capacity_mode: hasAgeGroups ? 'age_range' : 'simple',
            age_groups: hasAgeGroups
                ? camp.age_groups.map(g => ({
                    min_age: String(g.min_age),
                    max_age: String(g.max_age),
                    max_capacity: String(g.max_capacity),
                    price: String(g.price),
                }))
                : [],
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
                        <div className="form-group slug-field">
                            <label htmlFor="camp-slug">URL Slug</label>
                            <input type="text" id="camp-slug" name="slug" value={formData.slug} onChange={handleFormChange} placeholder="auto-generated-from-name" />
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
                        <div className="form-group capacity-mode-toggle">
                            <label>Capacity Mode</label>
                            <div className="capacity-mode-options">
                                <label>
                                    <input
                                        type="radio"
                                        name="capacity_mode"
                                        value="simple"
                                        checked={formData.capacity_mode === 'simple'}
                                        onChange={handleFormChange}
                                    />
                                    Simple
                                </label>
                                <label>
                                    <input
                                        type="radio"
                                        name="capacity_mode"
                                        value="age_range"
                                        checked={formData.capacity_mode === 'age_range'}
                                        onChange={handleFormChange}
                                    />
                                    Age Range
                                </label>
                            </div>
                        </div>

                        {formData.capacity_mode === 'simple' ? (
                            <>
                                <div className="form-group">
                                    <label htmlFor="camp-price">Price ($)</label>
                                    <input type="number" id="camp-price" name="price" value={formData.price} onChange={handleFormChange} required min="0.01" step="0.01" />
                                </div>
                                <div className="form-group">
                                    <label htmlFor="camp-cap">Max Capacity</label>
                                    <input type="number" id="camp-cap" name="max_capacity" value={formData.max_capacity} onChange={handleFormChange} min="1" placeholder="Leave blank for unlimited" />
                                </div>
                            </>
                        ) : (
                            <div className="age-group-editor">
                                <label>Age Groups</label>
                                {formData.age_groups.map((group, index) => (
                                    <div key={index} className="age-group-row">
                                        <input
                                            type="number"
                                            placeholder="Min Age"
                                            value={group.min_age}
                                            onChange={(e) => handleAgeGroupChange(index, 'min_age', e.target.value)}
                                            min="1"
                                            required
                                        />
                                        <input
                                            type="number"
                                            placeholder="Max Age"
                                            value={group.max_age}
                                            onChange={(e) => handleAgeGroupChange(index, 'max_age', e.target.value)}
                                            min="1"
                                            required
                                        />
                                        <input
                                            type="number"
                                            placeholder="Capacity"
                                            value={group.max_capacity}
                                            onChange={(e) => handleAgeGroupChange(index, 'max_capacity', e.target.value)}
                                            min="1"
                                            required
                                        />
                                        <input
                                            type="number"
                                            placeholder="Price ($)"
                                            value={group.price}
                                            onChange={(e) => handleAgeGroupChange(index, 'price', e.target.value)}
                                            min="0.01"
                                            step="0.01"
                                            required
                                        />
                                        <button type="button" className="btn btn-small btn-danger" onClick={() => removeAgeGroup(index)}>Remove</button>
                                    </div>
                                ))}
                                <button type="button" className="btn btn-small" onClick={addAgeGroup}>Add Age Group</button>
                            </div>
                        )}

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
                                        <h3>{camp.name} {camp.slug && <span className="camp-slug">/{camp.slug}</span>} {!camp.is_active && <span className="badge-inactive">Inactive</span>}</h3>
                                        <p>
                                            {new Date(camp.start_date).toLocaleDateString()} - {new Date(camp.end_date).toLocaleDateString()}
                                            {camp.age_groups && camp.age_groups.length > 0
                                                ? camp.age_groups.map((g, i) => (
                                                    <span key={i}>{' | '}Ages {g.min_age}-{g.max_age}: ${Number(g.price).toFixed(2)} ({g.registered_count}/{g.max_capacity})</span>
                                                ))
                                                : <>
                                                    {camp.price ? ` | $${Number(camp.price).toFixed(2)}` : ''}
                                                    {camp.max_capacity ? ` | Capacity: ${camp.registered_count}/${camp.max_capacity}` : ''}
                                                </>
                                            }
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
                                                            <td>${Number(r.registration.amount).toFixed(2)}</td>
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
