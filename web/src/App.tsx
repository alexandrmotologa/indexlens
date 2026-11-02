import { useState } from 'react';
import { Header, type TabKey } from './components/Header';
import { SlowQueryTable } from './components/SlowQueryTable';
import { ExplainVisualizer } from './components/ExplainVisualizer';
import { IndexAdvisorCard } from './components/IndexAdvisorCard';
import { LockContentionView } from './components/LockContentionView';
import { BloatWatchdog } from './components/BloatWatchdog';
import { CacheHitGauge } from './components/CacheHitGauge';
import { ConfigTuner } from './components/ConfigTuner';
import { SchemaHealthView } from './components/SchemaHealthView';
import { ConnectionPoolView } from './components/ConnectionPoolView';
import { useStats } from './hooks/useStats';

export function App() {
  const [activeTab, setActiveTab] = useState<TabKey>('queries');
  const [explainPrefill, setExplainPrefill] = useState('');

  const {
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
    refreshAll,
    fetchTuning,
    analyzePlan,
    executeQuery,
    comparePlans,
    simulateHypo,
    downloadReport,
  } = useStats();

  const handleExplainQuery = (query: string) => {
    setExplainPrefill(query);
    setActiveTab('explain');
  };

  return (
    <div className="min-h-screen bg-background text-gray-100 flex flex-col font-sans">
      <Header
        overview={overview}
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        onRefresh={refreshAll}
        onExportReport={downloadReport}
        isRefreshing={isRefreshing}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {activeTab === 'queries' && (
          <SlowQueryTable queries={queries} onExplainQuery={handleExplainQuery} />
        )}

        {activeTab === 'explain' && (
          <ExplainVisualizer
            initialResult={explainResult}
            initialDiff={planDiff}
            onAnalyzePlan={analyzePlan}
            onExecuteQuery={executeQuery}
            onComparePlans={comparePlans}
            prefilledQuery={explainPrefill}
          />
        )}

        {activeTab === 'advisor' && (
          <IndexAdvisorCard
            recommendations={recommendations}
            unusedIndexes={unusedIndexes}
            onSimulateHypo={simulateHypo}
          />
        )}

        {activeTab === 'locks' && (
          <LockContentionView tree={lockTree} />
        )}

        {activeTab === 'bloat' && (
          <BloatWatchdog
            bloatList={bloatList}
            workers={vacuumWorkers}
            wraparound={wraparound}
          />
        )}

        {activeTab === 'cache' && (
          <CacheHitGauge stats={cacheStats} />
        )}

        {activeTab === 'tuner' && (
          <ConfigTuner report={tuningReport} onProfileChange={fetchTuning} />
        )}

        {activeTab === 'schema' && (
          <SchemaHealthView report={schemaHealth} />
        )}

        {activeTab === 'pool' && (
          <ConnectionPoolView health={connectionPool} />
        )}
      </main>

      <footer className="border-t border-surface-border bg-surface/50 py-4 text-center text-xs text-gray-400 font-mono">
        IndexLens v1.1.0 • Local-First PostgreSQL Performance Cockpit • Non-Invasive Read-Only Telemetry
      </footer>
    </div>
  );
}

export default App;
