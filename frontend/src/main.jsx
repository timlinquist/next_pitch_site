import React from 'react'
import ReactDOM from 'react-dom/client'
import { Auth0Provider as Auth0ReactProvider } from '@auth0/auth0-react'
import { BrowserRouter, useNavigate } from 'react-router-dom'
import App from './App.jsx'
import './styles/common.css'

const Auth0ProviderWithNavigate = ({ children }) => {
  const navigate = useNavigate();

  const onRedirectCallback = (appState) => {
    console.log('[Auth0] Redirect callback triggered with appState:', appState);
    navigate(appState?.returnTo || '/', { 
      state: appState,
      replace: true 
    });
  };

  return (
    <Auth0ReactProvider
      domain={import.meta.env.VITE_AUTH0_CLIENT_DOMAIN}
      clientId={import.meta.env.VITE_AUTH0_CLIENT_ID}
      authorizationParams={{
        redirect_uri: window.location.origin,
        scope: 'openid profile email offline_access',
        audience: 'https://thenextpitch.org'
      }}
      useRefreshTokens={true}
      useRefreshTokensFallback={true}
      cacheLocation="localstorage"
      onRedirectCallback={onRedirectCallback}
    >
      {children}
    </Auth0ReactProvider>
  );
};

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter>
      <Auth0ProviderWithNavigate>
        <App />
      </Auth0ProviderWithNavigate>
    </BrowserRouter>
  </React.StrictMode>,
)
