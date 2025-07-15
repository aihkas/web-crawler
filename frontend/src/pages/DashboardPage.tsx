import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { Analysis } from '../types';
import { getAnalysisResults, submitUrlForAnalysis, deleteAnalyses } from '../services/api';
import { AnalysisTable } from '../components/AnalysisTable';
import { RowSelectionState } from '@tanstack/react-table';

const DashboardPage: React.FC = () => {
  const [analyses, setAnalyses] = useState<Analysis[]>([]);
  const [url, setUrl] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const [globalFilter, setGlobalFilter] = useState('');
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});

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

    // Filter data based on global filter input
  const filteredAnalyses = useMemo(() => {
    if (!globalFilter) return analyses;
    const lowercasedFilter = globalFilter.toLowerCase();
    return analyses.filter(analysis => 
      analysis.url.toLowerCase().includes(lowercasedFilter) ||
      (analysis.page_title && analysis.page_title.toLowerCase().includes(lowercasedFilter))
    );
  }, [analyses, globalFilter]);

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

    const handleDeleteSelected = async () => {
    const selectedIds = Object.keys(rowSelection).map(index => filteredAnalyses[parseInt(index)].id);
    if (selectedIds.length === 0) return;
    
    if (window.confirm(`Are you sure you want to delete ${selectedIds.length} item(s)?`)) {
        try {
            await deleteAnalyses(selectedIds);
            setRowSelection({}); // Clear selection
            fetchAnalyses(); // Refresh data
        } catch (err) {
            setError('Failed to delete items. Please try again.');
        }
    }
  };

  const selectedRowCount = Object.keys(rowSelection).length;

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
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
          <input
            type="text"
            value={globalFilter}
            onChange={(e) => setGlobalFilter(e.target.value)}
            placeholder="Search results..."
            style={{ padding: '0.5rem', minWidth: '300px' }}
          />
          {selectedRowCount > 0 && (
            <button onClick={handleDeleteSelected} style={{ background: '#e74c3c', color: 'white' }}>
              Delete Selected ({selectedRowCount})
            </button>
          )}
        </div>
        {isLoading ? (
          <p>Loading results...</p>
        ) : error ? (
          <p style={{ color: 'red' }}>{error}</p>
        ) : (
          <AnalysisTable 
            data={filteredAnalyses} 
            rowSelection={rowSelection}
            setRowSelection={setRowSelection}
          />
        )}
      </section>
    </div>
  );
};

export default DashboardPage;
