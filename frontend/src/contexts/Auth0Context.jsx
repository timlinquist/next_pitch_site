import React, { createContext, useContext, useState, useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';

const Auth0Context = createContext();

export const useAuth0Context = () => {
    const context = useContext(Auth0Context);
    if (!context) {
        throw new Error('useAuth0Context must be used within an Auth0Provider');
    }
    return context;
};

export const Auth0ContextProvider = ({ children }) => {
    const { user, getAccessTokenSilently } = useAuth0();
    const [isAdmin, setIsAdmin] = useState(false);

    useEffect(() => {
        const checkAdminStatus = async () => {
            if (!user) {
                setIsAdmin(false);
                return;
            }

            try {
                const token = await getAccessTokenSilently();
                const response = await fetch(`${process.env.REACT_APP_API_URL}/users/is-admin?email=${encodeURIComponent(user.email)}`, {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });

                if (!response.ok) {
                    throw new Error('Failed to check admin status');
                }

                const { is_admin } = await response.json();
                setIsAdmin(is_admin);
            } catch (error) {
                console.error('Error checking admin status:', error);
                setIsAdmin(false);
            }
        };

        checkAdminStatus();
    }, [user, getAccessTokenSilently]);

    return (
        <Auth0Context.Provider value={{ isAdmin }}>
            {children}
        </Auth0Context.Provider>
    );
}; 