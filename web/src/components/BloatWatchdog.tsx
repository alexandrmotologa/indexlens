import React from 'react';
import { ShieldAlert, Activity, AlertTriangle, CheckCircle, Database } from 'lucide-react';
import type { AutovacuumWorker, BloatInfo, WraparoundInfo } from '../types/api';

interface BloatWatchdogProps {
  bloatList: BloatInfo[];
  workers: AutovacuumWorker[];
  wraparound: WraparoundInfo | null;
}

export const BloatWatchdog: React.FC<BloatWatchdogProps> = ({
  bloatList,
  workers,
  wraparound,
}) => {
  return (
    <div className="space-y-6">
      {/* Top Cards: Wraparound & Active Autovacuums */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* TxID Wraparound Horizon Card */}
        <div className="bg-surface p-5 rounded-xl border border-surface-border space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <ShieldAlert className="w-5 h-5 text-blue-400" />
              <h3 className="font-semibold text-sm text-white">Transaction ID (TxID) Wraparound Guard</h3>
            </div>

            {wraparound && (
              <span
                className={`px-2 py-0.5 rounded text-xs font-bold ${
                  wraparound.status === 'HEALTHY'
                    ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                    : wraparound.status === 'WARNING'
                    ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                    : 'bg-red-500/10 text-red-400 border border-red-500/20'
                }`}
              >
                {wraparound.status}
              </span>
            )}
          </div>

          {wraparound ? (
            <div className="space-y-3 font-mono text-xs">
              <div className="flex justify-between text-gray-400">
                <span>Database: <span className="text-white font-semibold">{wraparound.databaseName}</span></span>
                <span>Consumed: <span className="text-white font-semibold">{wraparound.percentConsumed.toFixed(1)}%</span></span>
              </div>

              {/* Progress Bar */}
              <div className="w-full bg-surface-elevated rounded-full h-3 overflow-hidden border border-surface-border">
                <div
                  style={{ width: `${Math.min(wraparound.percentConsumed, 100)}%` }}
                  className={`h-full transition-all rounded-full ${
                    wraparound.percentConsumed > 75
                      ? 'bg-red-500'
                      : wraparound.percentConsumed > 50
                      ? 'bg-amber-500'
                      : 'bg-blue-500'
                  }`}
                />
              </div>

              <div className="grid grid-cols-2 gap-3 pt-2 text-[11px]">
                <div className="bg-surface-elevated p-2.5 rounded border border-surface-border">
                  <div className="text-gray-400">Current TxID Age</div>
                  <div className="text-white font-bold text-sm mt-0.5">
                    {wraparound.currentTxidAge.toLocaleString()}
                  </div>
                </div>
                <div className="bg-surface-elevated p-2.5 rounded border border-surface-border">
                  <div className="text-gray-400">Freeze Max Limit</div>
                  <div className="text-white font-bold text-sm mt-0.5">
                    {wraparound.maxTxidAge.toLocaleString()}
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-xs text-gray-500">Checking transaction age status...</div>
          )}
        </div>

        {/* Active Autovacuum Progress Card */}
        <div className="bg-surface p-5 rounded-xl border border-surface-border space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Activity className="w-5 h-5 text-emerald-400" />
              <h3 className="font-semibold text-sm text-white">Running Autovacuum Workers</h3>
            </div>
            <span className="text-xs font-mono text-gray-400">
              {workers.length} worker{workers.length === 1 ? '' : 's'} active
            </span>
          </div>

          <div className="space-y-3">
            {workers.length === 0 ? (
              <div className="py-8 text-center text-xs text-gray-500 flex flex-col items-center justify-center">
                <CheckCircle className="w-6 h-6 text-gray-600 mb-1" />
                No autovacuum processes currently in progress.
              </div>
            ) : (
              workers.map((w) => (
                <div
                  key={w.pid}
                  className="bg-surface-elevated p-3 rounded-lg border border-surface-border space-y-2 text-xs font-mono"
                >
                  <div className="flex justify-between items-center">
                    <span className="font-bold text-white">
                      PID {w.pid} • Table: <span className="text-blue-400">{w.table}</span>
                    </span>
                    <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-[10px]">
                      {w.phase}
                    </span>
                  </div>

                  <div className="w-full bg-background rounded-full h-2 overflow-hidden border border-surface-border">
                    <div
                      style={{ width: `${Math.min(w.progressPercentage, 100)}%` }}
                      className="bg-emerald-500 h-full rounded-full transition-all"
                    />
                  </div>

                  <div className="flex justify-between text-[11px] text-gray-400">
                    <span>
                      Scanned: {w.heapBlksScanned.toLocaleString()} / {w.heapBlksTotal.toLocaleString()} blocks
                    </span>
                    <span className="text-emerald-400 font-semibold">{w.progressPercentage.toFixed(1)}%</span>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      {/* Table Bloat Estimates */}
      <div className="bg-surface rounded-xl border border-surface-border overflow-hidden space-y-3 p-4">
        <div className="flex items-center justify-between border-b border-surface-border pb-3">
          <div className="flex items-center space-x-2">
            <Database className="w-4 h-4 text-blue-400" />
            <h3 className="font-semibold text-sm text-white">Table Bloat & Dead Tuples Analysis</h3>
          </div>
          <span className="text-xs text-gray-400 font-mono">
            {bloatList.length} relations inspected
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-surface-border text-left">
            <thead className="bg-surface-elevated text-xs font-semibold text-gray-400 uppercase tracking-wider">
              <tr>
                <th className="py-3 px-4">Relation Name</th>
                <th className="py-3 px-4 text-right">Total Disk Size</th>
                <th className="py-3 px-4 text-right">Dead Tuples</th>
                <th className="py-3 px-4 text-right">Dead Ratio</th>
                <th className="py-3 px-4 text-right">Estimated Bloat</th>
                <th className="py-3 px-4 text-center">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border text-sm font-mono">
              {bloatList.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-gray-500">
                    No table bloat telemetry available.
                  </td>
                </tr>
              ) : (
                bloatList.map((b) => (
                  <tr key={b.tableName} className="hover:bg-surface-elevated/40 transition-colors">
                    <td className="py-3 px-4 font-semibold text-gray-200">
                      {b.tableName}
                    </td>

                    <td className="py-3 px-4 text-right text-gray-300">
                      {b.sizePretty}
                    </td>

                    <td className="py-3 px-4 text-right text-gray-300">
                      {b.deadTuples.toLocaleString()}
                    </td>

                    <td className="py-3 px-4 text-right">
                      <span
                        className={
                          b.deadTupleRatio > 25
                            ? 'text-red-400 font-bold'
                            : b.deadTupleRatio > 15
                            ? 'text-amber-400 font-semibold'
                            : 'text-gray-300'
                        }
                      >
                        {b.deadTupleRatio.toFixed(1)}%
                      </span>
                    </td>

                    <td className="py-3 px-4 text-right text-amber-400 font-semibold">
                      {b.bloatPretty} ({b.bloatRatio.toFixed(0)}%)
                    </td>

                    <td className="py-3 px-4 text-center">
                      {b.deadTupleRatio > 25 ? (
                        <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded text-[10px] font-bold bg-red-500/10 text-red-400 border border-red-500/20">
                          <AlertTriangle className="w-3 h-3" />
                          <span>HIGH BLOAT</span>
                        </span>
                      ) : b.deadTupleRatio > 15 ? (
                        <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded text-[10px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">
                          <span>MODERATE</span>
                        </span>
                      ) : (
                        <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                          <span>OPTIMAL</span>
                        </span>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
