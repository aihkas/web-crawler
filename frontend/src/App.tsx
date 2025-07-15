import React from 'react';
import { Routes, Route } from 'react-router-dom';
import MainLayout from './components/layout/MainLayout';
import DashboardPage from './pages/DashboardPage';
import DetailsPage from './pages/DetailsPage';

function App() {
  return (
    <MainLayout>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/analysis/:id" element={<DetailsPage />} />
      </Routes>
    </MainLayout>
  );
}

export default App;
