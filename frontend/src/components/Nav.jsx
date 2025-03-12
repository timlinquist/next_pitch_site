import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth0 } from '@auth0/auth0-react';

const Nav = () => {
    const location = useLocation();
    const { isAuthenticated, loginWithRedirect, logout } = useAuth0();

    const isActive = (path) => {
        return location.pathname === path ? 'active' : '';
    };

    return (
        <nav className="navbar">
            <div className="nav-container">
                <Link to="/" className="nav-logo">
                    Next Pitch
                </Link>
                <div className="nav-links">
                    <Link to="/" className={`nav-link ${isActive('/')}`}>Home</Link>
                    <Link to="/about" className={`nav-link ${isActive('/about')}`}>About</Link>
                    <Link to="/services" className={`nav-link ${isActive('/services')}`}>Services</Link>
                    <Link to="/schedule" className={`nav-link ${isActive('/schedule')}`}>Schedule</Link>
                    <Link to="/contact" className={`nav-link ${isActive('/contact')}`}>Contact</Link>
                    {isAuthenticated ? (
                        <Link to="/account" className={`nav-link ${isActive('/account')}`}>
                            My Account
                        </Link>
                    ) : (
                        <button 
                            className={`nav-link ${isActive('/account')}`}
                            onClick={() => loginWithRedirect({
                                appState: { returnTo: '/account' }
                            })}
                        >
                            My Account
                        </button>
                    )}
                    {isAuthenticated ? (
                        <button 
                            className="nav-link" 
                            onClick={() => logout({ returnTo: window.location.origin })}
                        >
                            Logout
                        </button>
                    ) : (
                        <button 
                            className="nav-link" 
                            onClick={() => loginWithRedirect({
                                appState: { returnTo: location.pathname }
                            })}
                        >
                            Login
                        </button>
                    )}
                </div>
            </div>
        </nav>
    );
};

export default Nav; 