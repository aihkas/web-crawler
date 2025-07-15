import React, { useState, useEffect, useCallback } from 'react';
import { Analysis } from '../types';
import { getAnalysisResults, submitUrlForAnalysis } from '../services/api';

const DashboardPage: React.FC = () => {
  const [analyses, setAnalyses] = useState<Analysis[]>([]);
  const [url, setUrl] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const fetchAnalyses = useCallback(async () => {
    try {
      const data = await getAnalysisResults();
      setAnalyses(data);
      setError(null);
    } catch (err) {
      setError('Failed to fetch analysis results. Is the backend running?');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAnalyses();
    // Poll for new data every 5 seconds
    const intervalId = setInterval(fetchAnalyses, 5000);
    return () => clearInterval(intervalId); // Cleanup on unmount
  }, [fetchAnalyses]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url) {
      setSubmitError('URL cannot be empty.');
      return;
    }
    setSubmitError(null);
    try {
      await submitUrlForAnalysis(url);
      setUrl(''); // Clear input on success
      // Immediately fetch results to show the 'queued' item
      fetchAnalyses();
    } catch (err) {
      setSubmitError('Failed to submit URL. Please try again.');
    }
  };

  return (
    <div>
      <section style={{ marginBottom: '2rem' }}>
        <h2>Submit a new URL for Analysis</h2>
        <form onSubmit={handleSubmit} style={{ display: 'flex', gap: '1rem' }}>
          <input
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://example.com"
            style={{ flexGrow: 1, padding: '0.5rem' }}
            required
          />
          <button type="submit" style={{ padding: '0.5rem 1rem' }}>Analyze</button>
        </form>
        {submitError && <p style={{ color: 'red' }}>{submitError}</p>}
      </section>

      <section>
        <h2>Analysis Results</h2>
        {isLoading ? (
          <p>Loading results...</p>
        ) : error ? (
          <p style={{ color: 'red' }}>{error}</p>
        ) : (
          "table"
        )}
      </section>
    </div>
  );
};

export default DashboardPage;
