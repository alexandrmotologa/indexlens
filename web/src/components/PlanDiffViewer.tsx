import React, { useState } from 'react';
import { GitCompare, ArrowRight, TrendingDown, TrendingUp, CheckCircle, RotateCcw } from 'lucide-react';
import type { PlanDiffResult } from '../types/api';

interface PlanDiffViewerProps {
  initialDiff: PlanDiffResult | null;
  onCompare: (baseline: string, current: string) => Promise<PlanDiffResult | null>;
}

export const PlanDiffViewer: React.FC<PlanDiffViewerProps> = ({ initialDiff, onCompare }) => {
  const [diff, setDiff] = useState<PlanDiffResult | null>(initialDiff);
  const [baseJSON, setBaseJSON] = useState('');
  const [currJSON, setCurrJSON] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  React.useEffect(() => {
    if (initialDiff) {
      setDiff(initialDiff);
    }
  }, [initialDiff]);

  const handleRunDiff = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await onCompare(baseJSON, currJSON);
      if (res) setDiff(res);
    } catch (e: any) {
      setError(e.message || 'Failed to compare execution plans');
    } finally {
      setIsLoading(false);
    }
  };

  const handleLoadDemo = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await onCompare('', '');
      if (res) setDiff(res);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Top Banner */}
      <div className="bg-surface p-6 rounded-xl border border-surface-border flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center space-x-2">
            <GitCompare className="w-5 h-5 text-blue-400" />
            <h2 className="text-lg font-bold text-white">Execution Plan Diff &amp; Regression Analyzer</h2>
          </div>
          <p className="text-xs text-gray-400 mt-1 max-w-2xl">
            Compare two execution plans side-by-side (Baseline vs Current) to isolate why a query regressed or verify latency gains after creating an index.
          </p>
        </div>

        <button
          onClick={handleLoadDemo}
          className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-surface-elevated hover:bg-surface-border text-xs font-medium text-gray-200 border border-surface-border transition-colors"
        >
          <RotateCcw className="w-3.5 h-3.5 text-blue-400" />
          <span>Load Sample Diff</span>
        </button>
      </div>

      {/* Input Boxes */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 bg-surface p-4 rounded-xl border border-surface-border">
        <div className="space-y-2">
          <label className="text-xs font-semibold text-gray-400 uppercase">
            Baseline Plan (Before)
          </label>
          <textarea
            value={baseJSON}
            onChange={(e) => setBaseJSON(e.target.value)}
            placeholder="Paste baseline EXPLAIN JSON plan..."
            rows={3}
            className="w-full bg-surface-elevated border border-surface-border rounded-lg p-2.5 text-xs font-mono text-gray-200 placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
        </div>

        <div className="space-y-2">
          <label className="text-xs font-semibold text-gray-400 uppercase">
            Current Plan (After)
          </label>
          <textarea
            value={currJSON}
            onChange={(e) => setCurrJSON(e.target.value)}
            placeholder="Paste current EXPLAIN JSON plan..."
            rows={3}
            className="w-full bg-surface-elevated border border-surface-border rounded-lg p-2.5 text-xs font-mono text-gray-200 placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
        </div>

        {error && <div className="col-span-2 text-xs text-red-400 font-medium">{error}</div>}

        <div className="col-span-2 flex justify-end">
          <button
            onClick={handleRunDiff}
            disabled={isLoading}
            className="flex items-center space-x-1.5 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium transition-colors shadow-lg shadow-blue-500/20 disabled:opacity-50"
          >
            <GitCompare className="w-3.5 h-3.5" />
            <span>{isLoading ? 'Comparing...' : 'Compare Execution Plans'}</span>
          </button>
        </div>
      </div>

      {/* Diff Result View */}
      {diff && (
        <div className="space-y-4">
          {/* Verdict Banner */}
          <div
            className={`p-4 rounded-xl border flex items-center justify-between ${
              diff.verdict === 'IMPROVED'
                ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400'
                : diff.verdict === 'REGRESSED'
                ? 'bg-red-500/10 border-red-500/30 text-red-400'
                : 'bg-blue-500/10 border-blue-500/30 text-blue-400'
            }`}
          >
            <div className="flex items-center space-x-3">
              {diff.verdict === 'IMPROVED' ? (
                <TrendingDown className="w-6 h-6 flex-shrink-0" />
              ) : diff.verdict === 'REGRESSED' ? (
                <TrendingUp className="w-6 h-6 flex-shrink-0" />
              ) : (
                <CheckCircle className="w-6 h-6 flex-shrink-0" />
              )}
              <div>
                <div className="font-bold text-sm tracking-wide">
                  PLAN {diff.verdict}: {diff.timeDeltaPct.toFixed(1)}% Execution Duration
                </div>
                <div className="text-xs text-gray-300 mt-0.5">{diff.summary}</div>
              </div>
            </div>

            <div className="font-mono text-right text-xs">
              <div>Baseline: {diff.baselineExecutionTime.toFixed(1)} ms</div>
              <div>Current: {diff.currentExecutionTime.toFixed(1)} ms</div>
            </div>
          </div>

          {/* Node Transformations Table */}
          {diff.nodeChanges.length > 0 && (
            <div className="bg-surface rounded-xl border border-surface-border overflow-hidden p-4 space-y-3">
              <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">
                Operation Transformations &amp; Timing Shifts
              </h3>

              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-surface-border text-left">
                  <thead className="bg-surface-elevated text-xs font-semibold text-gray-400 uppercase tracking-wider">
                    <tr>
                      <th className="py-2.5 px-3">Node</th>
                      <th className="py-2.5 px-3">Baseline Operation</th>
                      <th className="py-2.5 px-3 w-8"></th>
                      <th className="py-2.5 px-3">Current Operation</th>
                      <th className="py-2.5 px-3 text-right">Delta (ms)</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-surface-border text-xs font-mono">
                    {diff.nodeChanges.map((change, idx) => (
                      <tr key={idx} className="hover:bg-surface-elevated/40 transition-colors">
                        <td className="py-2.5 px-3 text-gray-400">{change.path}</td>
                        <td className="py-2.5 px-3 text-red-300">{change.baselineType}</td>
                        <td className="py-2.5 px-3 text-gray-500">
                          <ArrowRight className="w-3.5 h-3.5" />
                        </td>
                        <td className="py-2.5 px-3 text-emerald-300 font-semibold">{change.currentType}</td>
                        <td className="py-2.5 px-3 text-right font-bold">
                          <span className={change.timeDeltaMs < 0 ? 'text-emerald-400' : 'text-red-400'}>
                            {change.timeDeltaMs > 0 ? `+${change.timeDeltaMs.toFixed(1)}` : change.timeDeltaMs.toFixed(1)} ms
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
