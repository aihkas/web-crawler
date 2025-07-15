import React from 'react';
import './MainLayout.css';

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  return (
    <div className="main-layout">
      <header className="main-header">
        <h1>Web Page Analyzer</h1>
      </header>
      <main className="main-content">
        {children}
      </main>
      <footer className="main-footer">
        <p>&copy; 2025 Web Crawler Inc. All rights reserved.</p>
      </footer>
    </div>
  );
};

export default MainLayout;
