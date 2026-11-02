import { useState, useEffect, useCallback } from 'react';
import type {
  SystemOverview,
  QueryStat,
  CacheHitStats,
  ExplainAnalysisResult,
  IndexRecommendation,
  UnusedIndex,
  LockContentionTree,
  PlanDiffResult,
  HypoSimulationResult,
  BloatInfo,
  AutovacuumWorker,
  WraparoundInfo,
  TuningReport,
  TuningProfile,
  SchemaHealthReport,
  ConnectionPoolHealth,
} from '../types/api';

export function useStats() {
  const [overview, setOverview] = useState<SystemOverview | null>(null);
  const [queries, setQueries] = useState<QueryStat[]>([]);
  const [cacheStats, setCacheStats] = useState<CacheHitStats | null>(null);
  const [recommendations, setRecommendations] = useState<IndexRecommendation[]>([]);
  const [unusedIndexes, setUnusedIndexes] = useState<UnusedIndex[]>([]);
  const [lockTree, setLockTree] = useState<LockContentionTree | null>(null);
  const [bloatList, setBloatList] = useState<BloatInfo[]>([]);
  const [vacuumWorkers, setVacuumWorkers] = useState<AutovacuumWorker[]>([]);
  const [wraparound, setWraparound] = useState<WraparoundInfo | null>(null);
  const [tuningReport, setTuningReport] = useState<TuningReport | null>(null);
  const [schemaHealth, setSchemaHealth] = useState<SchemaHealthReport | null>(null);
  const [connectionPool, setConnectionPool] = useState<ConnectionPoolHealth | null>(null);
  const [explainResult, setExplainResult] = useState<ExplainAnalysisResult | null>(null);
  const [planDiff, setPlanDiff] = useState<PlanDiffResult | null>(null);
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
      const [resRecs, resUnused] = await Promise.all([
        fetch('/api/indexes/advisor'),
        fetch('/api/indexes/unused'),
      ]);
      if (resRecs.ok) setRecommendations(await resRecs.json());
      if (resUnused.ok) setUnusedIndexes(await resUnused.json());
    } catch (e: any) {
      setError(e.message);
    }
  }, []);

  const fetchLockContention = useCallback(async () => {
    try {
      const res = await fetch('/api/activity/locks');
      if (res.ok) setLockTree(await res.json());
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

  const fetchTuning = useCallback(async (customProfile?: TuningProfile) => {
    try {
      const profile = customProfile || {
        totalRamGb: 32,
        vcpus: 8,
        diskType: 'NVMe',
        workloadType: 'OLTP',
      };
      const res = await fetch('/api/config/tuning', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(profile),
      });
      if (res.ok) setTuningReport(await res.json());
    } catch (e: any) {
      setError(e.message);
    }
  }, []);

  const fetchSchemaHealth = useCallback(async () => {
    try {
      const res = await fetch('/api/schema/health');
      if (res.ok) setSchemaHealth(await res.json());
    } catch (e: any) {
      setError(e.message);
    }
  }, []);

  const fetchConnectionPool = useCallback(async () => {
    try {
      const res = await fetch('/api/connections/pool');
      if (res.ok) setConnectionPool(await res.json());
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
      fetchLockContention(),
      fetchBloatAndAutovacuum(),
      fetchTuning(),
      fetchSchemaHealth(),
      fetchConnectionPool(),
    ]);
    setIsRefreshing(false);
  }, [
    fetchOverview,
    fetchQueries,
    fetchCache,
    fetchRecommendations,
    fetchLockContention,
    fetchBloatAndAutovacuum,
    fetchTuning,
    fetchSchemaHealth,
    fetchConnectionPool,
  ]);

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

    // Default initial plan diff
    fetch('/api/explain/diff', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data) setPlanDiff(data);
      })
      .catch(() => {});

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

  const comparePlans = async (baseline: string, current: string): Promise<PlanDiffResult | null> => {
    const res = await fetch('/api/explain/diff', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ baseline, current }),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to compare plans');
    }
    const data = await res.json();
    setPlanDiff(data);
    return data;
  };

  const simulateHypo = async (ddl: string, query: string): Promise<HypoSimulationResult> => {
    const res = await fetch('/api/indexes/simulate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ddl, query }),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Failed to run index simulation');
    }
    return await res.json();
  };

  const downloadReport = () => {
    window.open('/api/report/markdown', '_blank');
  };

  return {
    overview,
    queries,
    cacheStats,
    recommendations,
    unusedIndexes,
    lockTree,
    bloatList,
    vacuumWorkers,
    wraparound,
    tuningReport,
    schemaHealth,
    connectionPool,
    explainResult,
    planDiff,
    isRefreshing,
    error,
    refreshAll,
    fetchTuning,
    fetchSchemaHealth,
    fetchConnectionPool,
    analyzePlan,
    executeQuery,
    comparePlans,
    simulateHypo,
    downloadReport,
  };
}
