import React, { useState } from 'react';
import { Trash2, Copy, Check, ShieldAlert, Sparkles, Filter } from 'lucide-react';
import type { UnusedIndex } from '../types/api';

interface UnusedIndexListProps {
  unusedIndexes: UnusedIndex[];
}

export const UnusedIndexList: React.FC<UnusedIndexListProps> = ({ unusedIndexes }) => {
  const [filterType, setFilterType] = useState<'all' | 'zero' | 'redundant'>('all');
  const [copiedIndex, setCopiedIndex] = useState<string | null>(null);

  const totalWastedBytes = unusedIndexes.reduce((acc, curr) => acc + curr.indexSizeBytes, 0);

  const filtered = unusedIndexes.filter((idx) => {
    if (filterType === 'zero') return idx.indexScans === 0 && !idx.isRedundant;
    if (filterType === 'redundant') return idx.isRedundant;
    return true;
  });

  const handleCopy = (id: string, ddl: string) => {
    navigator.clipboard.writeText(ddl);
    setCopiedIndex(id);
    setTimeout(() => setCopiedIndex(null), 2500);
  };

  const formatBytes = (bytes: number) => {
    if (bytes >= 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <div className="space-y-6">
      {/* Banner */}
      <div className="bg-surface p-6 rounded-xl border border-surface-border flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center space-x-2">
            <Trash2 className="w-5 h-5 text-red-400" />
            <h2 className="text-lg font-bold text-white">Unused &amp; Redundant Index Cleanup</h2>
          </div>
          <p className="text-xs text-gray-400 mt-1 max-w-2xl">
            Detects indexes that incur write penalties on every INSERT/UPDATE/DELETE without serving read queries,
            as well as redundant indexes whose leading columns are already indexed by composite indexes.
          </p>
        </div>

        <div className="flex items-center space-x-3 text-xs font-mono">
          <div className="bg-surface-elevated px-3 py-2 rounded-lg border border-surface-border text-center">
            <div className="text-gray-400">Cleanup Candidates</div>
            <div className="text-lg font-bold text-amber-400">{unusedIndexes.length}</div>
          </div>
          <div className="bg-surface-elevated px-3 py-2 rounded-lg border border-surface-border text-center">
            <div className="text-gray-400">Recoverable Space</div>
            <div className="text-lg font-bold text-emerald-400">{formatBytes(totalWastedBytes)}</div>
          </div>
        </div>
      </div>

      {/* Filter Tabs */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2 text-xs">
          <Filter className="w-3.5 h-3.5 text-gray-400" />
          <button
            onClick={() => setFilterType('all')}
            className={`px-3 py-1.5 rounded-lg font-medium transition-colors ${
              filterType === 'all'
                ? 'bg-blue-600 text-white'
                : 'text-gray-400 hover:text-white hover:bg-surface-elevated'
            }`}
          >
            All Candidates ({unusedIndexes.length})
          </button>
          <button
            onClick={() => setFilterType('zero')}
            className={`px-3 py-1.5 rounded-lg font-medium transition-colors ${
              filterType === 'zero'
                ? 'bg-blue-600 text-white'
                : 'text-gray-400 hover:text-white hover:bg-surface-elevated'
            }`}
          >
            0 Scans Recorded
          </button>
          <button
            onClick={() => setFilterType('redundant')}
            className={`px-3 py-1.5 rounded-lg font-medium transition-colors ${
              filterType === 'redundant'
                ? 'bg-blue-600 text-white'
                : 'text-gray-400 hover:text-white hover:bg-surface-elevated'
            }`}
          >
            Redundant Indexes
          </button>
        </div>
      </div>

      {/* Index List */}
      <div className="space-y-4">
        {filtered.length === 0 ? (
          <div className="bg-surface p-12 text-center rounded-xl border border-surface-border text-gray-400 text-sm">
            <Sparkles className="w-8 h-8 text-emerald-400 mx-auto mb-2 opacity-80" />
            No unused or redundant indexes found. Your database indexing footprint is clean!
          </div>
        ) : (
          filtered.map((idx) => (
            <div
              key={idx.indexName}
              className="bg-surface rounded-xl border border-surface-border p-5 space-y-3 hover:border-red-500/40 transition-colors"
            >
              <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-2">
                <div className="flex items-center space-x-3">
                  <span
                    className={`px-2 py-0.5 rounded text-xs font-bold ${
                      idx.isRedundant
                        ? 'bg-purple-500/10 text-purple-400 border border-purple-500/20'
                        : 'bg-red-500/10 text-red-400 border border-red-500/20'
                    }`}
                  >
                    {idx.isRedundant ? 'REDUNDANT' : 'ZERO SCANS'}
                  </span>
                  <span className="font-mono text-sm font-bold text-white">
                    {idx.indexName} <span className="text-gray-400 font-normal">on {idx.tableName}</span>
                  </span>
                </div>

                <div className="text-xs font-mono text-gray-400">
                  Wasted Disk: <span className="text-amber-400 font-bold">{idx.indexSizePretty}</span>
                </div>
              </div>

              <div className="text-xs text-gray-300">
                <span className="text-gray-400 font-semibold">Diagnosis: </span>
                {idx.reason}
              </div>

              {/* DDL Drop Box */}
              <div className="relative group">
                <div className="bg-background p-3 rounded-lg border border-surface-border font-mono text-xs text-red-400 overflow-x-auto pr-24">
                  {idx.dropDdl}
                </div>

                <button
                  onClick={() => handleCopy(idx.indexName, idx.dropDdl)}
                  className="absolute right-2 top-2 flex items-center space-x-1.5 px-3 py-1 rounded-md bg-surface-elevated hover:bg-surface-border text-gray-200 text-xs font-medium border border-surface-border transition-colors shadow-sm"
                >
                  {copiedIndex === idx.indexName ? (
                    <>
                      <Check className="w-3.5 h-3.5 text-emerald-400" />
                      <span className="text-emerald-400 font-bold">Copied!</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5 text-gray-400" />
                      <span>Copy DROP</span>
                    </>
                  )}
                </button>
              </div>

              <div className="flex items-center space-x-1.5 text-[11px] text-gray-400">
                <ShieldAlert className="w-3.5 h-3.5 text-amber-400" />
                <span>
                  Uses <code className="text-gray-200 font-mono">CONCURRENTLY</code> so dropping the index does not lock ongoing read or write transactions.
                </span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};
