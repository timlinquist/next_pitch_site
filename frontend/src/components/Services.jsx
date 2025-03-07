import React from 'react';

const Services = () => {
    return (
        <div className="container">
            <div className="section">
                <h1>Our Services</h1>
                <p>We offer specialized training and analysis services to help you improve your pitching performance. Choose the option that best fits your needs.</p>
            </div>

            <div className="services-grid">
                <div className="service-card">
                    <h3>Individual Sessions</h3>
                    <div className="price">$50</div>
                    <p className="duration">30 or 60 minutes</p>
                    <p className="description">
                        Experience personalized, one-on-one instruction tailored to your specific needs. 
                        Our comprehensive approach integrates strength training, flexibility work, and mechanical refinements 
                        while incorporating essential breath work and mental performance techniques. Each session 
                        is designed to help you develop both the physical and mental aspects of your pitching game.
                    </p>
                    <a href="/schedule" className="btn">Book Now</a>
                </div>

                <div className="service-card">
                    <h3>Mechanical Analysis</h3>
                    <div className="price">$300</div>
                    <p className="description">
                        Submit your pitching footage for a detailed video analysis that will transform your technique. 
                        You'll receive a comprehensive review including custom-tailored drills, a targeted strength program, 
                        and specific mobility exercises designed to address your mechanical inefficiencies. Includes a 
                        detailed feedback report and follow-up consultation to ensure you're implementing the 
                        recommendations effectively.
                    </p>
                    <a href="/schedule" className="btn">Book Now</a>
                </div>

                <div className="service-card">
                    <h3>Team Training</h3>
                    <div className="price">$225</div>
                    <p className="duration">2 hours</p>
                    <p className="description">
                        Transform your pitching staff with our comprehensive team training session. We'll work with both 
                        players and coaches to develop a systematic approach to velocity and command. The session includes 
                        hands-on instruction, implementation strategies, and a detailed Q&A to ensure your coaching staff 
                        can effectively maintain and build upon the system long after we're gone.
                    </p>
                    <a href="/schedule" className="btn">Book Now</a>
                </div>
            </div>
        </div>
    );
};

export default Services; 