import React from 'react';

const About = () => {
    return (
        <div className="container">
            <div className="section">
                <h1>About Us</h1>
                <p>Next Pitch is a forward-thinking company dedicated to helping businesses and individuals perfect their pitch. We believe that every great idea deserves to be presented in the most compelling way possible.</p>
                <p>Our mission is to empower people to communicate their ideas effectively, whether they're pitching to investors, clients, or stakeholders. We combine proven techniques with innovative approaches to help you deliver your message with confidence and impact.</p>
            </div>

            <div className="section">
                <h2>Our Team</h2>
                <div className="team-grid">
                    <div className="team-member">
                        <img src="https://placehold.co/150x150" alt="Sarah Johnson" />
                        <h3>Sarah Johnson</h3>
                        <p>Founder & CEO</p>
                    </div>
                    <div className="team-member">
                        <img src="https://placehold.co/150x150" alt="Michael Chen" />
                        <h3>Michael Chen</h3>
                        <p>Head of Training</p>
                    </div>
                    <div className="team-member">
                        <img src="https://placehold.co/150x150" alt="Emma Rodriguez" />
                        <h3>Emma Rodriguez</h3>
                        <p>Content Director</p>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default About; 