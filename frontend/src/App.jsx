import React, { useEffect } from 'react'
import { Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom'
import { useAuth0 } from '@auth0/auth0-react'
import Nav from './components/Nav.jsx'
import Home from './components/Home.jsx'
import About from './components/About.jsx'
import Services from './components/Services.jsx'
import Schedule from './components/Schedule.jsx'
import Contact from './components/Contact.jsx'
import MechanicalAnalysisPage from './pages/MechanicalAnalysisPage.jsx'
import AccountPage from './pages/AccountPage.jsx'
import './styles/nav.css'
import './styles/common.css'
import './App.css'

// Protected route component
const ProtectedRoute = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuth0();

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/" />;
  }

  return children;
};

function App() {
  const { isAuthenticated, isLoading } = useAuth0();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      console.log('[App] Location state:', location.state);
      
      // Check for state from Auth0 redirect
      if (location.state?.selectedSlot) {
        console.log('[App] Found selected slot in state:', location.state.selectedSlot);
        
        // Navigate to schedule with the slot data
        navigate('/schedule', {
          state: { selectedSlot: location.state.selectedSlot },
          replace: true
        });
      }
    }
  }, [isAuthenticated, isLoading, navigate, location]);

  return (
    <div className="app" style={{ width: '100%', minHeight: '100vh' }}>
      <Nav />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/about" element={<About />} />
        <Route path="/services" element={<Services />} />
        <Route path="/schedule" element={<Schedule />} />
        <Route path="/contact" element={<Contact />} />
        <Route path="/mechanical-analysis" element={<MechanicalAnalysisPage />} />
        <Route 
          path="/account" 
          element={
            <ProtectedRoute>
              <AccountPage />
            </ProtectedRoute>
          } 
        />
      </Routes>
    </div>
  );
}

export default App;
