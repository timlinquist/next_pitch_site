import React from 'react';

const Services = () => {
    return (
        <div className="container">
            <div className="section">
                <h1>Our Services</h1>
                <p>We offer a range of services to help you perfect your pitch, from one-on-one coaching to group workshops. Choose the option that best fits your needs.</p>
            </div>

            <div className="services-grid">
                <div className="service-card">
                    <h3>One-on-One Coaching</h3>
                    <div className="price">$299/session</div>
                    <ul>
                        <li>Personalized feedback</li>
                        <li>Custom pitch development</li>
                        <li>Video recording & analysis</li>
                        <li>Follow-up support</li>
                    </ul>
                    <a href="/contact" className="btn">Book Now</a>
                </div>

                <div className="service-card">
                    <h3>Group Workshops</h3>
                    <div className="price">$499/workshop</div>
                    <ul>
                        <li>Interactive learning</li>
                        <li>Peer feedback</li>
                        <li>Practice sessions</li>
                        <li>Resource materials</li>
                    </ul>
                    <a href="/contact" className="btn">Book Now</a>
                </div>

                <div className="service-card">
                    <h3>Corporate Training</h3>
                    <div className="price">Custom pricing</div>
                    <ul>
                        <li>Team building</li>
                        <li>Custom curriculum</li>
                        <li>Progress tracking</li>
                        <li>Certification</li>
                    </ul>
                    <a href="/contact" className="btn">Contact Us</a>
                </div>
            </div>
        </div>
    );
};

export default Services; 