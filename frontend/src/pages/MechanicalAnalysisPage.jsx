import React from 'react';
import MechanicalAnalysis from '../components/MechanicalAnalysis';
import AuthRequired from '../components/AuthRequired';

const MechanicalAnalysisPage = () => {
    return (
        <div className="container">
            <div className="section">
                <h1>Mechanical Analysis</h1>
                <p>Get a detailed analysis of your pitching mechanics from our expert coaches. Upload your videos below and receive comprehensive feedback to improve your performance.</p>
            </div>
            <AuthRequired returnTo="/mechanical-analysis">
                <MechanicalAnalysis />
            </AuthRequired>
        </div>
    );
};

export default MechanicalAnalysisPage; 