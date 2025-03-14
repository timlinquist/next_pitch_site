import React, { createContext, useContext, useState, useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { getApiUrl } from '../utils/api';

const Auth0Context = createContext();

export const Auth0Provider = ({ children }) => {
    const auth0 = useAuth0();
    const [isAdmin, setIsAdmin] = useState(false);

    useEffect(() => {
        const checkAdminStatus = async () => {
            if (auth0.isAuthenticated) {
                try {
                    const response = await fetch(getApiUrl('users/me'), {
                        headers: {
                            'Authorization': `Bearer ${auth0.getAccessTokenSilently()}`
                        }
                    });
                    if (response.ok) {
                        const userData = await response.json();
                        setIsAdmin(userData.is_admin);
                    }
                } catch (error) {
                    console.error('Error checking admin status:', error);
                }
            }
        };

        checkAdminStatus();
    }, [auth0.isAuthenticated]);

    return (
        <Auth0Context.Provider value={{ ...auth0, isAdmin }}>
            {children}
        </Auth0Context.Provider>
    );
};

export const useAuth0Context = () => {
    const context = useContext(Auth0Context);
    if (!context) {
        throw new Error('useAuth0Context must be used within an Auth0Provider');
    }
    return context;
}; 