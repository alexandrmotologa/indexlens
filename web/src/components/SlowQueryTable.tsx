import React, { useState, useMemo } from 'react';
import { Search, ChevronDown, ChevronRight, Copy, Check, Flame, ArrowUpDown, AlertTriangle, Activity } from 'lucide-react';
import type { QueryStat } from '../types/api';

interface SlowQueryTableProps {
  queries: QueryStat[];
  onExplainQuery: (query: string) => void;
}

export const SlowQueryTable: React.FC<SlowQueryTableProps> = ({ queries, onExplainQuery }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState<'total' | 'mean' | 'calls' | 'rows' | 'cache' | 'jitter'>('total');
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const filteredQueries = useMemo(() => {
    return queries
      .filter((q) => q.query.toLowerCase().includes(searchTerm.toLowerCase()))
      .sort((a, b) => {
        switch (sortBy) {
          case 'mean':
            return b.meanExecTime - a.meanExecTime;
          case 'calls':
            return b.calls - a.calls;
          case 'rows':
            return b.rows - a.rows;
          case 'cache':
            return a.cacheHitRatio - b.cacheHitRatio;
          case 'jitter':
            return (b.jitterScore || 0) - (a.jitterScore || 0);
          case 'total':
          default:
            return b.totalExecTime - a.totalExecTime;
        }
      });
  }, [queries, searchTerm, sortBy]);

  const handleCopy = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const formatTime = (ms: number) => {
    if (ms >= 1000) {
      return `${(ms / 1000).toFixed(2)} s`;
    }
    return `${ms.toFixed(1)} ms`;
  };

  const getJitterBadge = (q: QueryStat) => {
    const level = q.jitterLevel || (q.stddevExecTime / (q.meanExecTime || 1) > 0.75 ? 'SPIKY' : 'STABLE');
    const score = q.jitterScore !== undefined ? q.jitterScore.toFixed(2) : (q.stddevExecTime / (q.meanExecTime || 1)).toFixed(2);

    if (level === 'SPIKY') {
      return (
        <span className="inline-flex items-center gap-1 rounded bg-rose-500/10 px-2 py-0.5 text-[11px] font-semibold text-rose-400 border border-rose-500/20" title={`Coefficient of variation: ${score} (Spiky latency)`}>
          <Activity className="h-3 w-3" />
          SPIKY ({score})
        </span>
      );
    }
    if (level === 'MODERATE') {
      return (
        <span className="inline-flex items-center gap-1 rounded bg-amber-500/10 px-2 py-0.5 text-[11px] font-semibold text-amber-400 border border-amber-500/20" title={`Coefficient of variation: ${score}`}>
          MODERATE ({score})
        </span>
      );
    }
    return (
      <span className="inline-flex items-center gap-1 rounded bg-emerald-500/10 px-2 py-0.5 text-[11px] font-semibold text-emerald-400 border border-emerald-500/20" title={`Coefficient of variation: ${score}`}>
        STABLE ({score})
      </span>
    );
  };

  return (
    <div className="space-y-4">
      {/* Controls Bar */}
      <div className="flex flex-col sm:flex-row justify-between items-center gap-3 bg-surface p-4 rounded-xl border border-surface-border">
        <div className="relative w-full sm:w-96">
          <Search className="w-4 h-4 text-gray-400 absolute left-3 top-3" />
          <input
            type="text"
            placeholder="Filter queries by SQL text, table, or function..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full bg-surface-elevated border border-surface-border rounded-lg pl-9 pr-4 py-2 text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
        </div>

        <div className="flex items-center space-x-2 w-full sm:w-auto">
          <span className="text-xs text-gray-400 flex items-center gap-1">
            <ArrowUpDown className="w-3.5 h-3.5" /> Sort by:
          </span>
          <div className="flex bg-surface-elevated p-1 rounded-lg border border-surface-border text-xs">
            <button
              onClick={() => setSortBy('total')}
              className={`px-3 py-1 rounded transition-colors ${
                sortBy === 'total' ? 'bg-blue-600 text-white font-medium' : 'text-gray-400 hover:text-white'
              }`}
            >
              Total Time
            </button>
            <button
              onClick={() => setSortBy('mean')}
              className={`px-3 py-1 rounded transition-colors ${
                sortBy === 'mean' ? 'bg-blue-600 text-white font-medium' : 'text-gray-400 hover:text-white'
              }`}
            >
              Mean Latency
            </button>
            <button
              onClick={() => setSortBy('calls')}
              className={`px-3 py-1 rounded transition-colors ${
                sortBy === 'calls' ? 'bg-blue-600 text-white font-medium' : 'text-gray-400 hover:text-white'
              }`}
            >
              Calls
            </button>
            <button
              onClick={() => setSortBy('cache')}
              className={`px-3 py-1 rounded transition-colors ${
                sortBy === 'cache' ? 'bg-blue-600 text-white font-medium' : 'text-gray-400 hover:text-white'
              }`}
            >
              Cache Miss
            </button>
            <button
              onClick={() => setSortBy('jitter')}
              className={`px-3 py-1 rounded transition-colors ${
                sortBy === 'jitter' ? 'bg-blue-600 text-white font-medium' : 'text-gray-400 hover:text-white'
              }`}
            >
              Jitter / Spikes
            </button>
          </div>
        </div>
      </div>

      {/* Query List */}
      <div className="bg-surface rounded-xl border border-surface-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-surface-border text-left">
            <thead className="bg-surface-elevated text-xs font-semibold text-gray-400 uppercase tracking-wider">
              <tr>
                <th className="py-3 px-4 w-8"></th>
                <th className="py-3 px-4">Query Pattern</th>
                <th className="py-3 px-4 w-28 text-right">Calls</th>
                <th className="py-3 px-4 w-32 text-right">Mean Time</th>
                <th className="py-3 px-4 w-32 text-right">Jitter</th>
                <th className="py-3 px-4 w-32 text-right">Total Time</th>
                <th className="py-3 px-4 w-28 text-right">Cache Hit</th>
                <th className="py-3 px-4 w-28 text-center">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border text-sm">
              {filteredQueries.length === 0 ? (
                <tr>
                  <td colSpan={8} className="py-12 text-center text-gray-400">
                    No queries found matching filter criteria.
                  </td>
                </tr>
              ) : (
                filteredQueries.map((q) => {
                  const isExpanded = expandedId === q.queryId;
                  const hasAntiPatterns = q.antiPatterns && q.antiPatterns.length > 0;
                  return (
                    <React.Fragment key={q.queryId}>
                      <tr className="hover:bg-surface-elevated/50 transition-colors group">
                        <td className="py-3 px-4">
                          <button
                            onClick={() => setExpandedId(isExpanded ? null : q.queryId)}
                            className="text-gray-400 hover:text-white p-1 rounded"
                          >
                            {isExpanded ? (
                              <ChevronDown className="w-4 h-4" />
                            ) : (
                              <ChevronRight className="w-4 h-4" />
                            )}
                          </button>
                        </td>

                        <td className="py-3 px-4 max-w-md">
                          <div className="flex items-center space-x-2">
                            <span className="font-mono text-xs text-gray-200 truncate block">
                              {q.query}
                            </span>
                            {hasAntiPatterns && (
                              <span className="flex items-center gap-1 rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-amber-400 border border-amber-500/20 shrink-0" title="Anti-pattern detected">
                                <AlertTriangle className="h-3 w-3" />
                                Anti-Pattern
                              </span>
                            )}
                            <button
                              onClick={() => handleCopy(q.queryId, q.query)}
                              className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-white transition-opacity"
                              title="Copy SQL Query"
                            >
                              {copiedId === q.queryId ? (
                                <Check className="w-3.5 h-3.5 text-emerald-400" />
                              ) : (
                                <Copy className="w-3.5 h-3.5" />
                              )}
                            </button>
                          </div>
                        </td>

                        <td className="py-3 px-4 text-right font-mono text-gray-300">
                          {q.calls.toLocaleString()}
                        </td>

                        <td className="py-3 px-4 text-right font-mono font-medium">
                          <span
                            className={
                              q.meanExecTime > 100
                                ? 'text-red-400'
                                : q.meanExecTime > 20
                                ? 'text-amber-400'
                                : 'text-emerald-400'
                            }
                          >
                            {formatTime(q.meanExecTime)}
                          </span>
                        </td>

                        <td className="py-3 px-4 text-right font-mono">
                          {getJitterBadge(q)}
                        </td>

                        <td className="py-3 px-4 text-right font-mono text-gray-300">
                          {formatTime(q.totalExecTime)}
                        </td>

                        <td className="py-3 px-4 text-right font-mono">
                          <span
                            className={`inline-block px-2 py-0.5 rounded text-xs font-semibold ${
                              q.cacheHitRatio >= 99
                                ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                                : q.cacheHitRatio >= 80
                                ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                                : 'bg-red-500/10 text-red-400 border border-red-500/20'
                            }`}
                          >
                            {q.cacheHitRatio.toFixed(1)}%
                          </span>
                        </td>

                        <td className="py-3 px-4 text-center">
                          <button
                            onClick={() => onExplainQuery(q.query)}
                            className="inline-flex items-center space-x-1 px-2.5 py-1 rounded bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 text-xs font-medium border border-blue-500/30 transition-colors"
                          >
                            <Flame className="w-3.5 h-3.5" />
                            <span>Explain</span>
                          </button>
                        </td>
                      </tr>

                      {/* Expanded telemetry details */}
                      {isExpanded && (
                        <tr className="bg-surface-elevated/80 border-t border-b border-surface-border">
                          <td colSpan={8} className="py-4 px-6">
                            <div className="space-y-4">
                              <div className="bg-background/80 p-3 rounded-lg border border-surface-border">
                                <div className="text-xs font-semibold text-gray-400 uppercase mb-1">
                                  Full Query Statement
                                </div>
                                <pre className="font-mono text-xs text-blue-300 whitespace-pre-wrap">
                                  {q.query}
                                </pre>
                              </div>

                              {/* Anti-Patterns Alert Section */}
                              {hasAntiPatterns && (
                                <div className="rounded-lg border border-amber-500/30 bg-amber-500/10 p-3.5">
                                  <div className="flex items-center gap-2 text-xs font-semibold text-amber-400 mb-2">
                                    <AlertTriangle className="h-4 w-4" />
                                    <span>Detected SQL Anti-Pattern(s)</span>
                                  </div>
                                  <div className="space-y-3">
                                    {q.antiPatterns!.map((ap, i) => (
                                      <div key={i} className="rounded border border-amber-500/20 bg-background/50 p-2.5 text-xs">
                                        <div className="flex items-center justify-between mb-1">
                                          <span className="font-mono font-bold text-amber-300">{ap.type}</span>
                                          <span className="text-[10px] rounded bg-amber-500/20 px-1.5 py-0.5 text-amber-200">
                                            {ap.severity}
                                          </span>
                                        </div>
                                        <p className="text-gray-300 mb-1.5">{ap.message}</p>
                                        <div className="text-indigo-300 font-medium mb-1.5">
                                          💡 Advice: {ap.rewriteAdvice}
                                        </div>
                                        {ap.suggestedDdl && (
                                          <div className="flex items-center justify-between rounded bg-zinc-950 px-2.5 py-1.5 mt-1 font-mono text-[11px] text-emerald-400">
                                            <span className="truncate mr-2">{ap.suggestedDdl}</span>
                                            <button
                                              onClick={() => handleCopy(`ap-${q.queryId}-${i}`, ap.suggestedDdl!)}
                                              className="flex items-center gap-1 rounded bg-muted/60 hover:bg-muted px-2 py-0.5 text-xs text-gray-200"
                                            >
                                              {copiedId === `ap-${q.queryId}-${i}` ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                                              Copy
                                            </button>
                                          </div>
                                        )}
                                      </div>
                                    ))}
                                  </div>
                                </div>
                              )}

                              {/* Latency Percentiles & Jitter */}
                              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs font-mono">
                                <div className="bg-surface p-2.5 rounded border border-surface-border">
                                  <div className="text-gray-400">Latency Percentiles</div>
                                  <div className="text-gray-200 mt-1">
                                    p50: {q.p50ExecTime || q.meanExecTime.toFixed(1)}ms • p95: {q.p95ExecTime || (q.meanExecTime * 1.5).toFixed(1)}ms
                                  </div>
                                  <div className="text-muted-foreground text-[10px] mt-0.5">
                                    p90: {q.p90ExecTime || (q.meanExecTime * 1.3).toFixed(1)}ms • p99: {q.p99ExecTime || (q.meanExecTime * 2.0).toFixed(1)}ms
                                  </div>
                                </div>

                                <div className="bg-surface p-2.5 rounded border border-surface-border">
                                  <div className="text-gray-400">Min / Max / Stddev</div>
                                  <div className="text-gray-200 mt-1">
                                    {formatTime(q.minExecTime)} / {formatTime(q.maxExecTime)}
                                  </div>
                                  <div className="text-muted-foreground text-[10px] mt-0.5">
                                    Stddev: {q.stddevExecTime.toFixed(1)}ms
                                  </div>
                                </div>

                                <div className="bg-surface p-2.5 rounded border border-surface-border">
                                  <div className="text-gray-400">Buffer Cache (Hit / Read)</div>
                                  <div className="text-gray-200 mt-1">
                                    {q.sharedBlksHit.toLocaleString()} / {q.sharedBlksRead.toLocaleString()}
                                  </div>
                                </div>

                                <div className="bg-surface p-2.5 rounded border border-surface-border">
                                  <div className="text-gray-400">Temp Disk Blocks (W / R)</div>
                                  <div className={q.tempBlksWritten > 0 ? 'text-amber-400 font-bold mt-1' : 'text-gray-200 mt-1'}>
                                    {q.tempBlksWritten.toLocaleString()} / {q.tempBlksRead.toLocaleString()}
                                  </div>
                                </div>
                              </div>
                            </div>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
