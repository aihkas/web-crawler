import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { Analysis } from '../types';
import { getAnalysisById } from '../services/api';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import './DetailsPage.css';

const COLORS = ['#0088FE', '#00C49F'];

const DetailsPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [analysis, setAnalysis] = useState<Analysis | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    const fetchDetails = async () => {
      try {
        const data = await getAnalysisById(id);
        setAnalysis(data);
      } catch (err) {
        setError('Failed to fetch analysis details.');
      } finally {
        setIsLoading(false);
      }
    };
    fetchDetails();
  }, [id]);
  
  if (isLoading) return <p>Loading details...</p>;
  if (error) return <p style={{ color: 'red' }}>{error}</p>;
  if (!analysis) return <p>No analysis data found.</p>;

  const linkData = [
    { name: 'Internal Links', value: analysis.internal_link_count },
    { name: 'External Links', value: analysis.external_link_count },
  ];

  return (
    <div className="details-page">
      <Link to="/" className="back-link">&larr; Back to Dashboard</Link>
      <h1>Analysis for: <a href={analysis.url} target="_blank" rel="noopener noreferrer">{analysis.url}</a></h1>
      
      <div className="details-grid">
        <div className="card">
          <h3>Page Summary</h3>
          <p><strong>Title:</strong> {analysis.page_title || 'N/A'}</p>
          <p><strong>HTML Version:</strong> {analysis.html_version || 'N/A'}</p>
          <p><strong>Login Form:</strong> {analysis.has_login_form ? 'Yes' : 'No'}</p>
          <p><strong>Status:</strong> <span className={`status-${analysis.status}`}>{analysis.status}</span></p>
        </div>

        <div className="card">
          <h3>Link Distribution</h3>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie data={linkData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={80} fill="#8884d8" label>
                {linkData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>
      
      {analysis.inaccessible_links && analysis.inaccessible_links.length > 0 && (
        <div className="card">
          <h3>Inaccessible Links ({analysis.inaccessible_links.length})</h3>
          <ul className="broken-links-list">
            {analysis.inaccessible_links.map((link, index) => (
              <li key={index}>
                <span className="status-code">{link.status_code}</span>
                <a href={link.url} target="_blank" rel="noopener noreferrer">{link.url}</a>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export default DetailsPage;
