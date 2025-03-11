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
                    <p className="duration">&nbsp;</p>
                    <p className="description">
                        Get a detailed mechanical analysis of your pitching motion. Upload two videos: one from the front view and one from the side view. Our experts will analyze your mechanics and provide a comprehensive report with actionable insights.
                    </p>
                    <a href="/mechanical-analysis" className="btn">Upload Now</a>
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