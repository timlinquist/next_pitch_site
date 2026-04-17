import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { loadStripe } from '@stripe/stripe-js';
import { Elements, CardElement, useStripe, useElements } from '@stripe/react-stripe-js';
import { PayPalScriptProvider, PayPalButtons } from '@paypal/react-paypal-js';
import { getApiUrl } from '../utils/api';
import '../styles/camps.css';

const stripePromise = loadStripe(import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY);

const CARD_ELEMENT_OPTIONS = {
    style: {
        base: {
            fontSize: '16px',
            color: '#333',
            fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
            '::placeholder': { color: '#999' },
        },
        invalid: { color: '#dc3545' },
    },
};

const StripeCheckoutForm = ({ camp, athleteData, onSuccess, onError, setProcessing }) => {
    const stripe = useStripe();
    const elements = useElements();
    const [cardError, setCardError] = useState(null);

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!stripe || !elements) return;

        setProcessing(true);
        setCardError(null);

        try {
            // Create registration + payment intent
            const response = await fetch(getApiUrl('register'), {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    athlete: {
                        name: athleteData.name,
                        age: parseInt(athleteData.age),
                        years_played: parseInt(athleteData.yearsPlayed) || 0,
                        position: athleteData.position,
                    },
                    camp_id: camp.id,
                    parent_email: athleteData.parentEmail,
                    parent_phone: athleteData.parentPhone,
                    payment_method: 'stripe',
                }),
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Registration failed');
            }

            const { registration_id, client_secret } = await response.json();

            // Confirm card payment
            const result = await stripe.confirmCardPayment(client_secret, {
                payment_method: { card: elements.getElement(CardElement) },
            });

            if (result.error) {
                throw new Error(result.error.message);
            }

            // Confirm on backend
            await fetch(getApiUrl('register/stripe-confirm'), {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ registration_id }),
            });

            onSuccess();
        } catch (err) {
            setCardError(err.message);
            onError(err.message);
        } finally {
            setProcessing(false);
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <div className="card-element-wrapper">
                <CardElement options={CARD_ELEMENT_OPTIONS} />
            </div>
            {cardError && <p className="payment-error">{cardError}</p>}
            <button type="submit" className="btn register-btn" disabled={!stripe}>
                Pay {camp ? `$${(camp.price_cents / 100).toFixed(2)}` : ''}
            </button>
        </form>
    );
};

const CampRegistrationPage = () => {
    const { slug } = useParams();
    const [camp, setCamp] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [paymentMethod, setPaymentMethod] = useState('stripe');
    const [processing, setProcessing] = useState(false);
    const [success, setSuccess] = useState(false);

    const [paypalRegistrationId, setPaypalRegistrationId] = useState(null);

    // Athlete form state
    const [athleteData, setAthleteData] = useState({
        name: '',
        age: '',
        yearsPlayed: '0',
        position: '',
        parentEmail: '',
        parentPhone: '',
    });

    useEffect(() => {
        const fetchCamp = async () => {
            try {
                const response = await fetch(getApiUrl(`camps/by-slug/${slug}`));
                if (!response.ok) throw new Error('Camp not found');
                const data = await response.json();
                setCamp(data);
            } catch (err) {
                setError('Failed to load camp details.');
            } finally {
                setLoading(false);
            }
        };

        fetchCamp();
    }, [slug]);

    const handleChange = (e) => {
        setAthleteData({ ...athleteData, [e.target.name]: e.target.value });
    };

    const isFormValid = () => {
        return athleteData.name && athleteData.age && athleteData.parentEmail;
    };

    const getAgeGroupFeedback = () => {
        if (!camp?.age_groups?.length || !athleteData.age) return null;
        const age = parseInt(athleteData.age);
        if (isNaN(age)) return null;

        const matched = camp.age_groups.find(g => age >= g.min_age && age <= g.max_age);
        if (!matched) {
            return { eligible: false, message: 'This age is not eligible for this camp' };
        }
        if (matched.spots_remaining <= 0) {
            return { eligible: false, message: `Ages ${matched.min_age}-${matched.max_age} group is full` };
        }
        return {
            eligible: true,
            message: `Ages ${matched.min_age}-${matched.max_age}: ${matched.spots_remaining} spot${matched.spots_remaining !== 1 ? 's' : ''} remaining`
        };
    };

    const handlePayPalCreateOrder = async () => {
        const response = await fetch(getApiUrl('register'), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                athlete: {
                    name: athleteData.name,
                    age: parseInt(athleteData.age),
                    years_played: parseInt(athleteData.yearsPlayed) || 0,
                    position: athleteData.position,
                },
                camp_id: camp.id,
                parent_email: athleteData.parentEmail,
                parent_phone: athleteData.parentPhone,
                payment_method: 'paypal',
            }),
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Registration failed');
        }

        const data = await response.json();
        setPaypalRegistrationId(data.registration_id);
        return data.paypal_order_id;
    };

    const handlePayPalApprove = async (data) => {
        setProcessing(true);
        try {
            const response = await fetch(getApiUrl('register/paypal-capture'), {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    registration_id: paypalRegistrationId,
                    paypal_order_id: data.orderID,
                }),
            });

            if (!response.ok) {
                throw new Error('Payment capture failed');
            }

            setSuccess(true);
        } catch (err) {
            setError(err.message);
        } finally {
            setProcessing(false);
        }
    };

    if (loading) {
        return (
            <div className="container">
                <div className="section"><p>Loading camp details...</p></div>
            </div>
        );
    }

    if (error && !camp) {
        return (
            <div className="container">
                <div className="section">
                    <h1>Registration</h1>
                    <p className="error">{error}</p>
                </div>
            </div>
        );
    }

    if (success) {
        return (
            <div className="container">
                <div className="section registration-success">
                    <h1>Registration Confirmed!</h1>
                    <p>Thank you for registering <strong>{athleteData.name}</strong> for <strong>{camp.name}</strong>.</p>
                    <p>A confirmation email has been sent to <strong>{athleteData.parentEmail}</strong>.</p>
                    <div className="success-details">
                        <p><strong>Camp:</strong> {camp.name}</p>
                        <p><strong>Dates:</strong> {new Date(camp.start_date).toLocaleDateString()} - {new Date(camp.end_date).toLocaleDateString()}</p>
                        <p><strong>Amount Paid:</strong> ${(camp.price_cents / 100).toFixed(2)}</p>
                    </div>
                </div>
            </div>
        );
    }

    const ageGroupFeedback = getAgeGroupFeedback();

    return (
        <div className="container">
            <div className="section">
                <h1>Register for {camp.name}</h1>
                <div className="camp-info-banner">
                    <p><strong>Dates:</strong> {new Date(camp.start_date).toLocaleDateString()} - {new Date(camp.end_date).toLocaleDateString()}</p>
                    <p><strong>Price:</strong> ${(camp.price_cents / 100).toFixed(2)}</p>
                    {camp.description && <p>{camp.description}</p>}
                    {camp.age_groups && camp.age_groups.length > 0 && (
                        <div className="age-group-spots" style={{ marginTop: '0.5rem' }}>
                            {camp.age_groups.map((g, i) => (
                                <p key={i} className="camp-spots" style={{ margin: '0.25rem 0' }}>
                                    Ages {g.min_age}-{g.max_age}: {g.spots_remaining > 0
                                        ? `${g.spots_remaining} spot${g.spots_remaining !== 1 ? 's' : ''} remaining`
                                        : 'Full'}
                                </p>
                            ))}
                        </div>
                    )}
                </div>
            </div>

            <div className="section">
                <h2>Athlete Information</h2>
                <div className="registration-form">
                    <div className="form-group">
                        <label htmlFor="name">Athlete Name</label>
                        <input
                            type="text"
                            id="name"
                            name="name"
                            value={athleteData.name}
                            onChange={handleChange}
                            required
                        />
                    </div>
                    <div className="form-group">
                        <label htmlFor="age">Age</label>
                        <input
                            type="number"
                            id="age"
                            name="age"
                            min="4"
                            max="25"
                            value={athleteData.age}
                            onChange={handleChange}
                            required
                        />
                        {ageGroupFeedback && (
                            <p className={ageGroupFeedback.eligible ? 'age-group-eligible' : 'age-group-ineligible'}>
                                {ageGroupFeedback.message}
                            </p>
                        )}
                    </div>
                    <div className="form-group">
                        <label htmlFor="yearsPlayed">Years Played</label>
                        <input
                            type="number"
                            id="yearsPlayed"
                            name="yearsPlayed"
                            min="0"
                            value={athleteData.yearsPlayed}
                            onChange={handleChange}
                        />
                    </div>
                    <div className="form-group">
                        <label htmlFor="position">Position</label>
                        <input
                            type="text"
                            id="position"
                            name="position"
                            placeholder="e.g. Pitcher, Catcher"
                            value={athleteData.position}
                            onChange={handleChange}
                        />
                    </div>
                    <div className="form-group">
                        <label htmlFor="parentEmail">Parent Email</label>
                        <input
                            type="email"
                            id="parentEmail"
                            name="parentEmail"
                            value={athleteData.parentEmail}
                            onChange={handleChange}
                            required
                        />
                    </div>
                    <div className="form-group">
                        <label htmlFor="parentPhone">Phone Number</label>
                        <input
                            type="tel"
                            id="parentPhone"
                            name="parentPhone"
                            value={athleteData.parentPhone}
                            onChange={handleChange}
                        />
                    </div>
                </div>
            </div>

            <div className="section">
                <h2>Payment</h2>
                {error && <p className="payment-error">{error}</p>}

                <div className="payment-method-tabs">
                    <button
                        className={`payment-tab ${paymentMethod === 'stripe' ? 'active' : ''}`}
                        onClick={() => setPaymentMethod('stripe')}
                        disabled={processing}
                    >
                        Credit Card
                    </button>
                    <button
                        className={`payment-tab ${paymentMethod === 'paypal' ? 'active' : ''}`}
                        onClick={() => setPaymentMethod('paypal')}
                        disabled={processing}
                    >
                        PayPal
                    </button>
                </div>

                <div className="payment-content">
                    {paymentMethod === 'stripe' ? (
                        <Elements stripe={stripePromise}>
                            <StripeCheckoutForm
                                camp={camp}
                                athleteData={athleteData}
                                onSuccess={() => setSuccess(true)}
                                onError={(msg) => setError(msg)}
                                setProcessing={setProcessing}
                            />
                        </Elements>
                    ) : (
                        <PayPalScriptProvider options={{ 'client-id': import.meta.env.VITE_PAYPAL_CLIENT_ID, currency: 'USD' }}>
                            <div className="paypal-button-wrapper">
                                {!isFormValid() ? (
                                    <p className="payment-note">Please fill in the required athlete information above.</p>
                                ) : (
                                    <PayPalButtons
                                        style={{ layout: 'vertical', shape: 'rect' }}
                                        disabled={!isFormValid() || processing}
                                        createOrder={handlePayPalCreateOrder}
                                        onApprove={handlePayPalApprove}
                                        onError={(err) => setError('PayPal payment failed. Please try again.')}
                                    />
                                )}
                            </div>
                        </PayPalScriptProvider>
                    )}
                </div>

                {processing && <p className="processing-message">Processing payment...</p>}
            </div>
        </div>
    );
};

export default CampRegistrationPage;
