import { useState, useEffect, useCallback } from 'react';
import type {
  SystemOverview,
  QueryStat,
  CacheHitStats,
  ExplainAnalysisResult,
  IndexRecommendation,
  BloatInfo,
  AutovacuumWorker,
  WraparoundInfo,
} from '../types/api';

export function useStats() {
  const [overview, setOverview] = useState<SystemOverview | null>(null);
  const [queries, setQueries] = useState<QueryStat[]>([]);
  const [cacheStats, setCacheStats] = useState<CacheHitStats | null>(null);
  const [recommendations, setRecommendations] = useState<IndexRecommendation[]>([]);
  const [bloatList, setBloatList] = useState<BloatInfo[]>([]);
  const [vacuumWorkers, setVacuumWorkers] = useState<AutovacuumWorker[]>([]);
  const [wraparound, setWraparound] = useState<WraparoundInfo | null>(null);
  const [explainResult, setExplainResult] = useState<ExplainAnalysisResult | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchOverview = useCallback(async () => {
    try {
      const res = await fetch('/api/overview');
      if (res.ok) setOverview(await res.json());
    } catch (e: any) {
      setError(e.message);
    }
  }, []);

  const fetchQueries = useCallback(async () => {
    try {
      const res = await fetch('/api/queries?limit=100&sort=total');
      if (res.ok) setQueries(await res.json());
    } catch (e: any) {
      setError(e.message);
    }
  }, []);

  const fetchCache = useCallback(async () => {
    try {
      const res = await fetch('/api/cache');
      if (res.ok) setCacheStats(await res.json());
    } catch (e: any) {
      setError(e.message);
    }
  }, []);

  const fetchRecommendations = useCallback(async () => {
    try {
      const res = await fetch('/api/indexes/advisor');
      if (res.ok) setRecommendations(await res.json());
    } catch (e: any) {
      setError(e.message);
    }
  }, []);

  const fetchBloatAndAutovacuum = useCallback(async () => {
    try {
      const [resBloat, resAuto] = await Promise.all([
        fetch('/api/bloat'),
        fetch('/api/autovacuum'),
      ]);
      if (resBloat.ok) setBloatList(await resBloat.json());
      if (resAuto.ok) {
        const data = await resAuto.json();
        setVacuumWorkers(data.workers || []);
        setWraparound(data.wraparound || null);
      }
    } catch (e: any) {
      setError(e.message);
    }
  }, []);

  const refreshAll = useCallback(async () => {
    setIsRefreshing(true);
    await Promise.all([
      fetchOverview(),
      fetchQueries(),
      fetchCache(),
      fetchRecommendations(),
      fetchBloatAndAutovacuum(),
    ]);
    setIsRefreshing(false);
  }, [fetchOverview, fetchQueries, fetchCache, fetchRecommendations, fetchBloatAndAutovacuum]);

  useEffect(() => {
    refreshAll();
    // Default initial explain demo plan
    fetch('/api/explain', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data) setExplainResult(data);
      })
      .catch(() => {});

    // Periodic telemetry polling every 10 seconds
    const interval = setInterval(refreshAll, 10000);
    return () => clearInterval(interval);
  }, [refreshAll]);

  const analyzePlan = async (jsonPlan: string): Promise<ExplainAnalysisResult | null> => {
    const res = await fetch('/api/explain', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ plan: jsonPlan }),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to analyze plan');
    }
    const data = await res.json();
    setExplainResult(data);
    return data;
  };

  const executeQuery = async (query: string): Promise<ExplainAnalysisResult | null> => {
    const res = await fetch('/api/explain', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query }),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to explain query');
    }
    const data = await res.json();
    setExplainResult(data);
    return data;
  };

  return {
    overview,
    queries,
    cacheStats,
    recommendations,
    bloatList,
    vacuumWorkers,
    wraparound,
    explainResult,
    isRefreshing,
    error,
    refreshAll,
    analyzePlan,
    executeQuery,
  };
}
