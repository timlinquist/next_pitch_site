import React from 'react';
import Nav from './Nav';
import '../styles/nav.css';
import '../styles/common.css';

const Home = () => {
    return (
        <>
            <Nav />
            <div className="container">
                <div className="section">
                    <h1>The Most Important Pitch is the ... Next Pitch</h1>
                </div>
            </div>
        </>
    );
};

export default Home; 