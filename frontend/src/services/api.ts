import axios from 'axios';
import { Analysis } from '../types';

const api = axios.create({
  baseURL: process.env.REACT_APP_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${process.env.REACT_APP_API_TOKEN}`,
  },
});

export const getAnalysisResults = async (): Promise<Analysis[]> => {
  const response = await api.get('/results');
  return response.data;
};

export const submitUrlForAnalysis = async (url: string): Promise<any> => {
  const response = await api.post('/analyze', { url });
  return response.data;
};
