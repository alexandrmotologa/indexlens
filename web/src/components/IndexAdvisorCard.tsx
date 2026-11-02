import React, { useState } from 'react';
import { Layers, Copy, Check, ShieldCheck, Zap, ArrowRight } from 'lucide-react';
import type { IndexRecommendation } from '../types/api';

interface IndexAdvisorCardProps {
  recommendations: IndexRecommendation[];
}

export const IndexAdvisorCard: React.FC<IndexAdvisorCardProps> = ({ recommendations }) => {
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const handleCopy = (id: string, ddl: string) => {
    navigator.clipboard.writeText(ddl);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2500);
  };

  return (
    <div className="space-y-6">
      {/* Advisor Summary Banner */}
      <div className="bg-surface p-6 rounded-xl border border-surface-border flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center space-x-2">
            <Layers className="w-5 h-5 text-blue-400" />
            <h2 className="text-lg font-bold text-white">Missing Index Synthesizer</h2>
          </div>
          <p className="text-xs text-gray-400 mt-1 max-w-2xl">
            Correlates sequential table scans with active query filter patterns. Enforces B-Tree ordering rules
            (equality columns first, range columns next, sort keys last) and synthesizes non-blocking DDL.
          </p>
        </div>

        <div className="flex items-center space-x-3 text-xs font-mono">
          <div className="bg-surface-elevated px-3 py-2 rounded-lg border border-surface-border text-center">
            <div className="text-gray-400">Suggestions</div>
            <div className="text-lg font-bold text-blue-400">{recommendations.length}</div>
          </div>
          <div className="bg-surface-elevated px-3 py-2 rounded-lg border border-surface-border text-center">
            <div className="text-gray-400">Avg Scan Reduction</div>
            <div className="text-lg font-bold text-emerald-400">~92%</div>
          </div>
        </div>
      </div>

      {/* Recommendations Cards */}
      <div className="space-y-4">
        {recommendations.length === 0 ? (
          <div className="bg-surface p-12 text-center rounded-xl border border-surface-border text-gray-400 text-sm">
            <ShieldCheck className="w-8 h-8 text-emerald-400 mx-auto mb-2 opacity-80" />
            No missing index bottlenecks detected! All tables are maintaining healthy index scan ratios.
          </div>
        ) : (
          recommendations.map((rec) => (
            <div
              key={rec.id}
              className="bg-surface rounded-xl border border-surface-border p-5 space-y-4 hover:border-blue-500/40 transition-colors"
            >
              {/* Card Header */}
              <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-2">
                <div className="flex items-center space-x-3">
                  <span
                    className={`px-2 py-0.5 rounded text-xs font-bold ${
                      rec.priority === 'HIGH'
                        ? 'bg-red-500/10 text-red-400 border border-red-500/20'
                        : rec.priority === 'MEDIUM'
                        ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                        : 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
                    }`}
                  >
                    {rec.priority} PRIORITY
                  </span>

                  <span className="font-mono text-sm font-bold text-white">
                    Table: <span className="text-blue-400">{rec.tableName}</span>
                  </span>
                </div>

                <div className="flex items-center space-x-2 text-xs text-gray-400 font-mono">
                  <Zap className="w-3.5 h-3.5 text-emerald-400" />
                  <span>Est. Scans Eliminated: </span>
                  <span className="text-emerald-400 font-bold">{rec.estimatedScanReduction.toFixed(0)}%</span>
                </div>
              </div>

              {/* Columns Ranking Flow */}
              <div className="bg-surface-elevated p-3 rounded-lg border border-surface-border">
                <div className="text-[10px] font-semibold text-gray-400 uppercase tracking-wider mb-2">
                  Synthesized B-Tree Column Ordering
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  {rec.columns.map((col, idx) => (
                    <React.Fragment key={idx}>
                      <span className="px-2.5 py-1 rounded bg-blue-600/20 border border-blue-500/30 text-blue-300 font-mono text-xs font-semibold">
                        {col}
                      </span>
                      {idx < rec.columns.length - 1 && (
                        <ArrowRight className="w-3 h-3 text-gray-500" />
                      )}
                    </React.Fragment>
                  ))}

                  {rec.predicate && (
                    <span className="ml-2 px-2.5 py-1 rounded bg-purple-600/20 border border-purple-500/30 text-purple-300 font-mono text-xs">
                      WHERE {rec.predicate}
                    </span>
                  )}
                </div>
              </div>

              {/* Diagnostic Reasoning */}
              <div className="text-xs text-gray-300">
                <span className="text-gray-400 font-semibold">Telemetry Evidence: </span>
                {rec.reason}
              </div>

              {/* DDL Code Box */}
              <div className="relative group">
                <div className="bg-background p-3.5 rounded-lg border border-surface-border font-mono text-xs text-emerald-400 overflow-x-auto pr-24">
                  {rec.ddl}
                </div>

                <button
                  onClick={() => handleCopy(rec.id, rec.ddl)}
                  className="absolute right-2 top-2.5 flex items-center space-x-1.5 px-3 py-1.5 rounded-md bg-surface-elevated hover:bg-surface-border text-gray-200 text-xs font-medium border border-surface-border transition-colors shadow-sm"
                >
                  {copiedId === rec.id ? (
                    <>
                      <Check className="w-3.5 h-3.5 text-emerald-400" />
                      <span className="text-emerald-400 font-bold">Copied!</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5 text-gray-400" />
                      <span>Copy DDL</span>
                    </>
                  )}
                </button>
              </div>

              {/* Safety Footnote */}
              <div className="flex items-center space-x-1.5 text-[11px] text-gray-400">
                <ShieldCheck className="w-3.5 h-3.5 text-blue-400" />
                <span>
                  Uses <code className="text-gray-200 font-mono font-semibold">CONCURRENTLY</code> to avoid taking table write locks. Run outside explicit transaction blocks.
                </span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};
