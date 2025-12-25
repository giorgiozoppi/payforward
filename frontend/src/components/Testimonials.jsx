import React, { useEffect, useRef } from 'react';

const testimonials = [
  {
    id: 1,
    name: "Sarah Mitchell",
    location: "Portland, Oregon",
    avatar: "👩‍💼",
    story: "When I lost my job during the pandemic, a stranger on PayForwardNow covered my rent for two months. Now employed again, I've helped 7 families with groceries. The chain continues.",
    impact: "Helped 7 families",
    date: "Member since 2021"
  },
  {
    id: 2,
    name: "Marcus Chen",
    location: "Toronto, Canada",
    avatar: "👨‍🎓",
    story: "A mentor from this community spent 50+ hours helping me prepare for interviews. I landed my dream job and now I mentor three young professionals. Paying it forward feels incredible.",
    impact: "Mentored 3 professionals",
    date: "Member since 2022"
  },
  {
    id: 3,
    name: "Elena Rodriguez",
    location: "Mexico City, Mexico",
    avatar: "👩‍⚕️",
    story: "My mother needed surgery we couldn't afford. The community raised funds in 48 hours. She's healthy now, and I volunteer 10 hours weekly at a free clinic. None shall remain behind.",
    impact: "100+ volunteer hours",
    date: "Member since 2020"
  },
  {
    id: 4,
    name: "James Okonkwo",
    location: "Lagos, Nigeria",
    avatar: "👨‍💻",
    story: "Someone donated a laptop so I could learn programming. Two years later, I run a coding bootcamp for underprivileged youth. 200 students and counting.",
    impact: "Trained 200+ students",
    date: "Member since 2021"
  },
  {
    id: 5,
    name: "Yuki Tanaka",
    location: "Tokyo, Japan",
    avatar: "👩‍🎨",
    story: "After the earthquake, strangers rebuilt our community center. Now I organize art therapy for disaster survivors. Kindness heals what words cannot express.",
    impact: "500+ therapy sessions",
    date: "Member since 2019"
  },
  {
    id: 6,
    name: "David Bergström",
    location: "Stockholm, Sweden",
    avatar: "👨‍🍳",
    story: "Started a pay-what-you-can café after receiving help during tough times. We've served 10,000+ meals. Everyone deserves dignity and a warm meal.",
    impact: "10,000+ meals served",
    date: "Member since 2020"
  }
];

const Testimonials = () => {
  const sectionRef = useRef(null);
  const cardsRef = useRef([]);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('visible');
          }
        });
      },
      { threshold: 0.1, rootMargin: '0px 0px -50px 0px' }
    );

    cardsRef.current.forEach((card) => {
      if (card) observer.observe(card);
    });

    return () => observer.disconnect();
  }, []);

  return (
    <section className="testimonials-section" id="stories" ref={sectionRef}>
      <div className="testimonials-bg">
        <div className="grid-overlay"></div>
      </div>

      <div className="testimonials-container">
        <div className="section-header">
          <span className="section-tag">Real Stories</span>
          <h2 className="section-title">Lives Transformed</h2>
          <p className="section-subtitle">
            Every story represents a chain of kindness that started with one person
            deciding to make a difference.
          </p>
        </div>

        <div className="testimonials-grid">
          {testimonials.map((testimonial, index) => (
            <div
              key={testimonial.id}
              className="testimonial-card"
              ref={(el) => (cardsRef.current[index] = el)}
              style={{ animationDelay: `${index * 0.1}s` }}
            >
              <div className="card-header">
                <div className="avatar-wrapper">
                  <span className="avatar">{testimonial.avatar}</span>
                  <span className="avatar-ring"></span>
                </div>
                <div className="user-info">
                  <h4 className="user-name">{testimonial.name}</h4>
                  <span className="user-location">
                    <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                      <path d="M6 1C4.067 1 2.5 2.567 2.5 4.5C2.5 7.25 6 11 6 11C6 11 9.5 7.25 9.5 4.5C9.5 2.567 7.933 1 6 1ZM6 6C5.172 6 4.5 5.328 4.5 4.5C4.5 3.672 5.172 3 6 3C6.828 3 7.5 3.672 7.5 4.5C7.5 5.328 6.828 6 6 6Z" fill="currentColor"/>
                    </svg>
                    {testimonial.location}
                  </span>
                </div>
              </div>

              <p className="card-story">{testimonial.story}</p>

              <div className="card-footer">
                <div className="impact-badge">
                  <span className="impact-icon">✨</span>
                  <span>{testimonial.impact}</span>
                </div>
                <span className="member-date">{testimonial.date}</span>
              </div>

              <div className="card-glow"></div>
            </div>
          ))}
        </div>

        <div className="testimonials-cta">
          <p>Join thousands of changemakers around the world</p>
          <button className="outline-btn">
            Read More Stories
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <path d="M3 8H13M13 8L9 4M13 8L9 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
        </div>
      </div>
    </section>
  );
};

export default Testimonials;
