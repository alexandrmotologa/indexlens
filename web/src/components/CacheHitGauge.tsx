import React from 'react';
import { Database, HardDrive, CheckCircle2, AlertTriangle, Layers } from 'lucide-react';
import type { CacheHitStats } from '../types/api';

interface CacheHitGaugeProps {
  stats: CacheHitStats | null;
}

export const CacheHitGauge: React.FC<CacheHitGaugeProps> = ({ stats }) => {
  if (!stats) {
    return (
      <div className="bg-surface p-12 rounded-xl border border-surface-border text-center text-gray-400">
        Loading buffer cache metrics...
      </div>
    );
  }

  const isOptimal = stats.overallHitRatio >= 99.0;

  return (
    <div className="space-y-6">
      {/* Top Banner & Main Ratios */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Overall Ratio */}
        <div className="bg-surface p-6 rounded-xl border border-surface-border space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-gray-400 font-semibold uppercase tracking-wider">
              Overall Buffer Cache Hit Ratio
            </span>
            {isOptimal ? (
              <CheckCircle2 className="w-5 h-5 text-emerald-400" />
            ) : (
              <AlertTriangle className="w-5 h-5 text-amber-400" />
            )}
          </div>

          <div className="flex items-baseline space-x-2">
            <span
              className={`text-4xl font-mono font-bold ${
                isOptimal ? 'text-emerald-400' : 'text-amber-400'
              }`}
            >
              {stats.overallHitRatio.toFixed(1)}%
            </span>
            <span className="text-xs text-gray-400 font-mono">
              target: &gt; 99.0%
            </span>
          </div>

          {/* Progress Bar */}
          <div className="w-full bg-surface-elevated rounded-full h-2.5 overflow-hidden border border-surface-border">
            <div
              style={{ width: `${Math.min(stats.overallHitRatio, 100)}%` }}
              className={`h-full rounded-full transition-all ${
                isOptimal ? 'bg-emerald-500' : 'bg-amber-500'
              }`}
            />
          </div>

          <div className="text-[11px] text-gray-400 pt-1">
            {isOptimal
              ? 'Excellent working set retention in PostgreSQL shared_buffers.'
              : 'Disk buffer misses detected. Queries are waiting for storage I/O.'}
          </div>
        </div>

        {/* Heap Blocks Cache */}
        <div className="bg-surface p-6 rounded-xl border border-surface-border space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-gray-400 font-semibold uppercase tracking-wider">
              Heap Table Hit Ratio
            </span>
            <Database className="w-5 h-5 text-blue-400" />
          </div>

          <div className="text-4xl font-mono font-bold text-white">
            {stats.globalHeapHitRatio.toFixed(1)}%
          </div>

          <div className="grid grid-cols-2 gap-2 text-xs font-mono pt-2">
            <div className="bg-surface-elevated p-2 rounded border border-surface-border">
              <span className="text-gray-400 text-[10px] block">Blocks Hit:</span>
              <span className="text-emerald-400 font-bold">{stats.totalHeapHit.toLocaleString()}</span>
            </div>
            <div className="bg-surface-elevated p-2 rounded border border-surface-border">
              <span className="text-gray-400 text-[10px] block">Disk Reads:</span>
              <span className="text-amber-400 font-bold">{stats.totalHeapRead.toLocaleString()}</span>
            </div>
          </div>
        </div>

        {/* Index Blocks Cache */}
        <div className="bg-surface p-6 rounded-xl border border-surface-border space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-gray-400 font-semibold uppercase tracking-wider">
              Index Hit Ratio
            </span>
            <Layers className="w-5 h-5 text-purple-400" />
          </div>

          <div className="text-4xl font-mono font-bold text-white">
            {stats.globalIndexHitRatio.toFixed(1)}%
          </div>

          <div className="grid grid-cols-2 gap-2 text-xs font-mono pt-2">
            <div className="bg-surface-elevated p-2 rounded border border-surface-border">
              <span className="text-gray-400 text-[10px] block">Blocks Hit:</span>
              <span className="text-emerald-400 font-bold">{stats.totalIndexHit.toLocaleString()}</span>
            </div>
            <div className="bg-surface-elevated p-2 rounded border border-surface-border">
              <span className="text-gray-400 text-[10px] block">Disk Reads:</span>
              <span className="text-amber-400 font-bold">{stats.totalIndexRead.toLocaleString()}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Tables Below 99% Threshold */}
      <div className="bg-surface rounded-xl border border-surface-border overflow-hidden p-5 space-y-4">
        <div className="flex items-center justify-between border-b border-surface-border pb-3">
          <div className="flex items-center space-x-2">
            <HardDrive className="w-4 h-4 text-amber-400" />
            <h3 className="font-semibold text-sm text-white">
              Relations with Sub-Optimal Cache Retention (&lt; 99%)
            </h3>
          </div>
          <span className="text-xs text-gray-400 font-mono">
            {stats.tablesBelowThreshold.length} tables flagged
          </span>
        </div>

        {stats.tablesBelowThreshold.length === 0 ? (
          <div className="py-8 text-center text-xs text-gray-400">
            All tables are exceeding the 99% cache hit threshold.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-surface-border text-left">
              <thead className="bg-surface-elevated text-xs font-semibold text-gray-400 uppercase tracking-wider">
                <tr>
                  <th className="py-3 px-4">Relation Name</th>
                  <th className="py-3 px-4 text-right">Heap Hit %</th>
                  <th className="py-3 px-4 text-right">Heap Disk Reads</th>
                  <th className="py-3 px-4 text-right">Index Hit %</th>
                  <th className="py-3 px-4 text-right">Index Disk Reads</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border text-sm font-mono">
                {stats.tablesBelowThreshold.map((t) => (
                  <tr key={t.tableName} className="hover:bg-surface-elevated/40 transition-colors">
                    <td className="py-3 px-4 font-semibold text-blue-400">{t.tableName}</td>
                    <td className="py-3 px-4 text-right">
                      <span
                        className={
                          t.heapHitRatio < 80
                            ? 'text-red-400 font-bold'
                            : t.heapHitRatio < 95
                            ? 'text-amber-400 font-semibold'
                            : 'text-gray-300'
                        }
                      >
                        {t.heapHitRatio.toFixed(1)}%
                      </span>
                    </td>
                    <td className="py-3 px-4 text-right text-gray-300">
                      {t.heapBlksRead.toLocaleString()}
                    </td>
                    <td className="py-3 px-4 text-right">
                      <span
                        className={
                          t.indexHitRatio < 95
                            ? 'text-amber-400 font-semibold'
                            : 'text-gray-300'
                        }
                      >
                        {t.indexHitRatio.toFixed(1)}%
                      </span>
                    </td>
                    <td className="py-3 px-4 text-right text-gray-300">
                      {t.indexBlksRead.toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
