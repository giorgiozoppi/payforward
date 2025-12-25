import React, { useEffect, useRef } from 'react';

const HeroSection = ({ onJoinClick }) => {
  const heroRef = useRef(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('visible');
          }
        });
      },
      { threshold: 0.1 }
    );

    const elements = heroRef.current?.querySelectorAll('.animate-in');
    elements?.forEach((el) => observer.observe(el));

    return () => observer.disconnect();
  }, []);

  return (
    <section className="hero" ref={heroRef}>
      <div className="hero-content">
        <div className="hero-badge animate-in">
          <span className="badge-dot"></span>
          <span>Building a Movement of Kindness</span>
        </div>
        
        <h1 className="hero-title animate-in">
          <span className="title-line">Nobody Shall</span>
          <span className="title-line highlight">Remain Behind</span>
        </h1>
        
        <p className="hero-subtitle animate-in">
          A community where every act of kindness echoes infinitely.
          Give without expecting, receive without forgetting,
          and together we rise.
        </p>
        
        <div className="hero-actions animate-in">
          <button className="primary-btn" onClick={onJoinClick}>
            <span className="btn-content">
              <span className="btn-text">Join the Movement</span>
              <span className="btn-icon">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                  <path d="M4 10H16M16 10L11 5M16 10L11 15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </span>
            </span>
          </button>
          <button className="secondary-btn">
            <span>Watch Our Story</span>
            <span className="play-icon">▶</span>
          </button>
        </div>

        <div className="hero-stats animate-in">
          <div className="stat">
            <span className="stat-number">2.4M+</span>
            <span className="stat-label">Acts of Kindness</span>
          </div>
          <div className="stat-divider"></div>
          <div className="stat">
            <span className="stat-number">180+</span>
            <span className="stat-label">Countries</span>
          </div>
          <div className="stat-divider"></div>
          <div className="stat">
            <span className="stat-number">$12M</span>
            <span className="stat-label">Paid Forward</span>
          </div>
        </div>
      </div>

      <div className="hero-visual animate-in">
        <div className="visual-ring ring-1"></div>
        <div className="visual-ring ring-2"></div>
        <div className="visual-ring ring-3"></div>
        <div className="visual-center">
          <div className="center-icon">
            <svg viewBox="0 0 100 100" fill="none">
              <circle cx="50" cy="50" r="45" stroke="currentColor" strokeWidth="2" opacity="0.3"/>
              <path d="M50 20C45 30 35 35 25 35C25 55 35 70 50 80C65 70 75 55 75 35C65 35 55 30 50 20Z" fill="currentColor"/>
            </svg>
          </div>
        </div>
        <div className="floating-card card-1">
          <div className="card-avatar">👨‍🦱</div>
          <div className="card-content">
            <span className="card-action">Paid it forward</span>
            <span className="card-amount">$50</span>
          </div>
        </div>
        <div className="floating-card card-2">
          <div className="card-avatar">👩‍🦰</div>
          <div className="card-content">
            <span className="card-action">Received help</span>
            <span className="card-amount">Groceries</span>
          </div>
        </div>
        <div className="floating-card card-3">
          <div className="card-avatar">🧑‍🦳</div>
          <div className="card-content">
            <span className="card-action">Mentored</span>
            <span className="card-amount">3 hours</span>
          </div>
        </div>
      </div>

      <div className="scroll-indicator">
        <div className="scroll-mouse">
          <div className="scroll-wheel"></div>
        </div>
        <span>Scroll to explore</span>
      </div>
    </section>
  );
};

export default HeroSection;
