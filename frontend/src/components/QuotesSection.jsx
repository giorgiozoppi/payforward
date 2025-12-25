import React, { useEffect, useRef, useState } from 'react';

const quotes = [
  {
    text: "No one has ever become poor by giving.",
    author: "Anne Frank",
    role: "Author & Diarist"
  },
  {
    text: "We make a living by what we get, but we make a life by what we give.",
    author: "Winston Churchill",
    role: "Former UK Prime Minister"
  },
  {
    text: "The best way to find yourself is to lose yourself in the service of others.",
    author: "Mahatma Gandhi",
    role: "Civil Rights Leader"
  },
  {
    text: "Carry out a random act of kindness, with no expectation of reward.",
    author: "Princess Diana",
    role: "Humanitarian"
  },
  {
    text: "Remember there's no such thing as a small act of kindness.",
    author: "Scott Adams",
    role: "Creator of Dilbert"
  }
];

const QuotesSection = () => {
  const [activeQuote, setActiveQuote] = useState(0);
  const sectionRef = useRef(null);

  useEffect(() => {
    const interval = setInterval(() => {
      setActiveQuote((prev) => (prev + 1) % quotes.length);
    }, 6000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('visible');
          }
        });
      },
      { threshold: 0.2 }
    );

    if (sectionRef.current) {
      observer.observe(sectionRef.current);
    }

    return () => observer.disconnect();
  }, []);

  return (
    <section className="quotes-section" id="mission" ref={sectionRef}>
      <div className="quotes-bg">
        <div className="quote-mark left">"</div>
        <div className="quote-mark right">"</div>
      </div>
      
      <div className="quotes-container">
        <div className="section-header">
          <span className="section-tag">Words of Wisdom</span>
          <h2 className="section-title">Inspiration That Moves Us</h2>
        </div>

        <div className="quotes-carousel">
          {quotes.map((quote, index) => (
            <div 
              key={index}
              className={`quote-card ${index === activeQuote ? 'active' : ''}`}
            >
              <div className="quote-icon">
                <svg viewBox="0 0 40 40" fill="none">
                  <path d="M12 20H8C8 14 10 10 16 8V12C12 13 12 16 12 20ZM28 20H24C24 14 26 10 32 8V12C28 13 28 16 28 20Z" fill="currentColor"/>
                </svg>
              </div>
              <blockquote className="quote-text">
                {quote.text}
              </blockquote>
              <div className="quote-author">
                <div className="author-avatar">
                  {quote.author.charAt(0)}
                </div>
                <div className="author-info">
                  <span className="author-name">{quote.author}</span>
                  <span className="author-role">{quote.role}</span>
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className="quotes-dots">
          {quotes.map((_, index) => (
            <button
              key={index}
              className={`dot ${index === activeQuote ? 'active' : ''}`}
              onClick={() => setActiveQuote(index)}
              aria-label={`Go to quote ${index + 1}`}
            />
          ))}
        </div>

        <div className="philosophy-cards">
          <div className="philosophy-card">
            <div className="phil-icon">🌊</div>
            <h3>The Ripple Effect</h3>
            <p>Every act of kindness creates a ripple that extends far beyond what we can see.</p>
          </div>
          <div className="philosophy-card">
            <div className="phil-icon">🔄</div>
            <h3>Infinite Cycle</h3>
            <p>What goes around comes around. Give freely and receive abundantly.</p>
          </div>
          <div className="philosophy-card">
            <div className="phil-icon">🤝</div>
            <h3>United Purpose</h3>
            <p>Together we build a world where none shall remain behind.</p>
          </div>
        </div>
      </div>
    </section>
  );
};

export default QuotesSection;
