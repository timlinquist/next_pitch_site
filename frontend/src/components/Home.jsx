import React from 'react';
import Nav from './Nav';
import '../styles/nav.css';
import '../styles/common.css';
import '../styles/home.css';

const Home = () => {
    return (
        <>
            <Nav />
            <div className="home-container">
                <div className="home-background"></div>
                <div className="home-content">
                    <h1>Your Most Important Pitch ... is <span className="highlight">YOUR NEXT PITCH!</span></h1>
                </div>
            </div>
        </>
    );
};

export default Home; 