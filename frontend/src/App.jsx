import React, { useState, useEffect } from 'react';
import './styles/App.css';
import LoginModal from './components/LoginModal';
import Testimonials from './components/Testimonials';
import HeroSection from './components/HeroSection';
import QuotesSection from './components/QuotesSection';
import Footer from './components/Footer';

function App() {
  const [showLogin, setShowLogin] = useState(false);
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 50);
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <div className="app">
      {/* Animated background */}
      <div className="bg-gradient"></div>
      <div className="bg-noise"></div>
      <div className="floating-shapes">
        <div className="shape shape-1"></div>
        <div className="shape shape-2"></div>
        <div className="shape shape-3"></div>
        <div className="shape shape-4"></div>
      </div>

      {/* Navigation */}
      <nav className={`navbar ${scrolled ? 'scrolled' : ''}`}>
        <div className="nav-container">
          <div className="logo">
            <span className="logo-icon">⟡</span>
            <span className="logo-text">PayForwardNow</span>
          </div>
          <div className="nav-links">
            <a href="#mission" className="nav-link">Mission</a>
            <a href="#stories" className="nav-link">Stories</a>
            <a href="#impact" className="nav-link">Impact</a>
            <button 
              className="login-btn"
              onClick={() => setShowLogin(true)}
            >
              <span>Log in</span>
            </button>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <HeroSection onJoinClick={() => setShowLogin(true)} />

      {/* Quotes Section */}
      <QuotesSection />

      {/* Testimonials */}
      <Testimonials />

      {/* Call to Action */}
      <section className="cta-section">
        <div className="cta-container">
          <h2 className="cta-title">Ready to Make a Difference?</h2>
          <p className="cta-text">
            Join thousands who believe in the power of paying it forward.
            Every act of kindness creates a ripple that never ends.
          </p>
          <button 
            className="cta-button"
            onClick={() => setShowLogin(true)}
          >
            Start Your Journey
            <span className="btn-arrow">→</span>
          </button>
        </div>
      </section>

      {/* Footer */}
      <Footer />

      {/* Login Modal */}
      {showLogin && <LoginModal onClose={() => setShowLogin(false)} />}
    </div>
  );
}

export default App;
