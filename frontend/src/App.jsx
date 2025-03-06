import React from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Nav from './components/Nav.jsx'
import Home from './components/Home.jsx'
import About from './components/About.jsx'
import Services from './components/Services.jsx'
import Schedule from './components/Schedule.jsx'
import Contact from './components/Contact.jsx'
import './styles/nav.css'
import './styles/common.css'
import './App.css'

function App() {
  return (
    <Router>
      <div className="app" style={{ width: '100%', minHeight: '100vh' }}>
        <Nav />
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/about" element={<About />} />
          <Route path="/services" element={<Services />} />
          <Route path="/schedule" element={<Schedule />} />
          <Route path="/contact" element={<Contact />} />
        </Routes>
      </div>
    </Router>
  )
}

export default App
