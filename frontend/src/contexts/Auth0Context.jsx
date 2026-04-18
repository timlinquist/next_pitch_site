import React, { createContext, useContext, useState, useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { getApiUrl } from '../utils/api';

const Auth0Context = createContext();

export const useAuth0Context = () => {
    const context = useContext(Auth0Context);
    if (!context) {
        throw new Error('useAuth0Context must be used within an Auth0ContextProvider');
    }
    return context;
};

export const Auth0ContextProvider = ({ children }) => {
    const { user, isAuthenticated, isLoading, getAccessTokenSilently } = useAuth0();
    const [backendUser, setBackendUser] = useState(null);
    const [userLoading, setUserLoading] = useState(true);

    useEffect(() => {
        const syncUser = async () => {
            if (isLoading) return;
            if (!isAuthenticated || !user) {
                setBackendUser(null);
                setUserLoading(false);
                return;
            }

            try {
                const token = await getAccessTokenSilently();
                const response = await fetch(getApiUrl('users/me'), {
                    headers: { Authorization: `Bearer ${token}` },
                });
                if (response.ok) {
                    setBackendUser(await response.json());
                }
            } catch (error) {
                console.error('Error syncing user:', error);
            } finally {
                setUserLoading(false);
            }
        };

        syncUser();
    }, [user, isAuthenticated, isLoading, getAccessTokenSilently]);

    return (
        <Auth0Context.Provider value={{
            backendUser,
            isAdmin: backendUser?.is_admin ?? false,
            userLoading,
        }}>
            {children}
        </Auth0Context.Provider>
    );
}; 