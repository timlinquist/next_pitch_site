import React from 'react';
import { useAuth0 } from '@auth0/auth0-react';

const AuthRequired = ({ children, returnTo }) => {
    const { isAuthenticated, isLoading, loginWithRedirect } = useAuth0();

    if (isLoading) {
        return <div className="container">Loading...</div>;
    }

    if (!isAuthenticated) {
        return (
            <div className="container">
                <p>Please login or signup to continue</p>
                <button 
                    onClick={() => loginWithRedirect({ appState: { returnTo } })} 
                    className="btn"
                >
                    Log In
                </button>
            </div>
        );
    }

    return children;
};

export default AuthRequired; 