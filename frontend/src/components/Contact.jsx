import React from 'react';

const Contact = () => {
    const handleSubmit = async (e) => {
        e.preventDefault();
        
        const submitBtn = document.getElementById('submitBtn');
        const formMessage = document.getElementById('formMessage');
        
        // Disable submit button and show loading state
        submitBtn.disabled = true;
        submitBtn.textContent = 'Sending...';
        formMessage.style.display = 'none';
        
        const formData = {
            name: document.getElementById('name').value,
            email: document.getElementById('email').value,
            subject: document.getElementById('subject').value,
            message: document.getElementById('message').value
        };

        try {
            const response = await fetch('/api/contact', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(formData)
            });

            const data = await response.json();
            
            if (response.ok) {
                formMessage.textContent = 'Message sent successfully! We\'ll get back to you soon.';
                formMessage.className = 'form-message success';
                e.target.reset();
            } else {
                throw new Error(data.error || 'Failed to send message');
            }
        } catch (error) {
            formMessage.textContent = error.message || 'Failed to send message. Please try again.';
            formMessage.className = 'form-message error';
        } finally {
            formMessage.style.display = 'block';
            submitBtn.disabled = false;
            submitBtn.textContent = 'Send Message';
        }
    };

    return (
        <div className="container">
            <div className="section">
                <h1>Contact Us</h1>
                <p>Get in touch with us to learn more about our services or to schedule a consultation. We're here to help you perfect your pitch.</p>
            </div>

            <div className="section">
                <div className="contact-grid">
                    <div className="contact-info">
                        <h2>Get in Touch</h2>
                        <div className="info-item">
                            <h3>Address</h3>
                            <p>123 Pitch Street, Suite 100<br />San Francisco, CA 94105</p>
                        </div>
                        <div className="info-item">
                            <h3>Phone</h3>
                            <p>(555) 123-4567</p>
                        </div>
                        <div className="info-item">
                            <h3>Email</h3>
                            <p>info@nextpitch.com</p>
                        </div>
                        <div className="info-item">
                            <h3>Hours</h3>
                            <p>Monday - Friday<br />9:00 AM - 6:00 PM PST</p>
                        </div>
                    </div>

                    <div className="contact-form">
                        <h2>Send us a Message</h2>
                        <form id="contactForm" onSubmit={handleSubmit}>
                            <div className="form-group">
                                <label htmlFor="name">Name</label>
                                <input type="text" id="name" name="name" required />
                            </div>
                            <div className="form-group">
                                <label htmlFor="email">Email</label>
                                <input type="email" id="email" name="email" required />
                            </div>
                            <div className="form-group">
                                <label htmlFor="subject">Subject</label>
                                <input type="text" id="subject" name="subject" required />
                            </div>
                            <div className="form-group">
                                <label htmlFor="message">Message</label>
                                <textarea id="message" name="message" required></textarea>
                            </div>
                            <button type="submit" className="btn" id="submitBtn">Send Message</button>
                            <div id="formMessage" className="form-message"></div>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Contact; 