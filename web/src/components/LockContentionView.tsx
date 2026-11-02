import React from 'react';
import { Lock, AlertCircle, CheckCircle2, Clock, User, ArrowRight } from 'lucide-react';
import type { LockContentionTree } from '../types/api';

interface LockContentionViewProps {
  tree: LockContentionTree | null;
}

export const LockContentionView: React.FC<LockContentionViewProps> = ({ tree }) => {
  if (!tree) {
    return (
      <div className="bg-surface p-12 rounded-xl border border-surface-border text-center text-gray-400">
        Loading live lock contention telemetry...
      </div>
    );
  }

  const formatDuration = (ms: number) => {
    if (ms >= 1000) return `${(ms / 1000).toFixed(1)} s`;
    return `${ms.toFixed(0)} ms`;
  };

  return (
    <div className="space-y-6">
      {/* Top Banner */}
      <div className="bg-surface p-6 rounded-xl border border-surface-border flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center space-x-2">
            <Lock className="w-5 h-5 text-amber-400" />
            <h2 className="text-lg font-bold text-white">Live Query Activity &amp; Lock Contention</h2>
          </div>
          <p className="text-xs text-gray-400 mt-1 max-w-2xl">
            Real-time inspection of active database processes, blocked waiting queries, and idle-in-transaction connections holding exclusive locks.
          </p>
        </div>

        <div className="flex items-center space-x-3 text-xs font-mono">
          <div className="bg-surface-elevated px-3 py-2 rounded-lg border border-surface-border text-center">
            <div className="text-gray-400">Active Connections</div>
            <div className="text-lg font-bold text-blue-400">{tree.activeProcesses.length}</div>
          </div>
          <div className="bg-surface-elevated px-3 py-2 rounded-lg border border-surface-border text-center">
            <div className="text-gray-400">Blocked Queries</div>
            <div className={`text-lg font-bold ${tree.totalBlockedQueries > 0 ? 'text-red-400' : 'text-emerald-400'}`}>
              {tree.totalBlockedQueries}
            </div>
          </div>
        </div>
      </div>

      {/* Lock Status Alert */}
      {!tree.hasContention ? (
        <div className="bg-emerald-500/10 border border-emerald-500/20 p-4 rounded-xl flex items-center space-x-3 text-emerald-400 text-xs font-medium">
          <CheckCircle2 className="w-5 h-5 flex-shrink-0" />
          <span>No lock contention detected. All active transactions are progressing without blocking dependencies.</span>
        </div>
      ) : (
        <div className="bg-red-500/10 border border-red-500/30 p-4 rounded-xl flex items-center space-x-3 text-red-400 text-xs font-medium">
          <AlertCircle className="w-5 h-5 flex-shrink-0" />
          <span>
            Lock contention active! {tree.totalBlockedQueries} quer{tree.totalBlockedQueries === 1 ? 'y is' : 'ies are'} currently blocked waiting for locks held by {tree.rootBlockers.length} root transaction(s).
          </span>
        </div>
      )}

      {/* Root Blockers Section */}
      {tree.rootBlockers.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">
            Root Blocker Transactions ({tree.rootBlockers.length})
          </h3>
          <div className="space-y-3">
            {tree.rootBlockers.map((b) => (
              <div
                key={b.pid}
                className="bg-surface p-4 rounded-xl border border-red-500/40 space-y-3"
              >
                <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-2 border-b border-surface-border pb-2 text-xs font-mono">
                  <div className="flex items-center space-x-2">
                    <span className="px-2 py-0.5 rounded bg-red-500 text-white font-bold text-[10px]">
                      BLOCKER PID {b.pid}
                    </span>
                    <span className="text-gray-300 flex items-center gap-1">
                      <User className="w-3.5 h-3.5 text-gray-400" /> {b.user}
                    </span>
                    <span className="text-gray-500">•</span>
                    <span className="text-amber-400 font-semibold">{b.state}</span>
                  </div>

                  <div className="flex items-center space-x-2 text-gray-400">
                    <Clock className="w-3.5 h-3.5" />
                    <span>Duration: <span className="text-white font-bold">{formatDuration(b.durationMs)}</span></span>
                  </div>
                </div>

                <div className="bg-background p-3 rounded-lg border border-surface-border font-mono text-xs text-blue-300">
                  {b.query}
                </div>

                {b.blockedPids && b.blockedPids.length > 0 && (
                  <div className="text-xs text-gray-300 flex items-center space-x-2 font-mono">
                    <ArrowRight className="w-4 h-4 text-red-400" />
                    <span>Holding locks blocking PIDs: </span>
                    <span className="text-red-400 font-bold">{b.blockedPids.join(', ')}</span>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* All Active Processes Table */}
      <div className="bg-surface rounded-xl border border-surface-border overflow-hidden p-4 space-y-3">
        <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider">
          Active Database Connections &amp; Wait Events ({tree.activeProcesses.length})
        </h3>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-surface-border text-left">
            <thead className="bg-surface-elevated text-xs font-semibold text-gray-400 uppercase tracking-wider">
              <tr>
                <th className="py-2.5 px-3 w-20">PID</th>
                <th className="py-2.5 px-3 w-28">User</th>
                <th className="py-2.5 px-3 w-32">State</th>
                <th className="py-2.5 px-3 w-28 text-right">Duration</th>
                <th className="py-2.5 px-3 w-36">Wait Event</th>
                <th className="py-2.5 px-3">Current Query</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border text-xs font-mono">
              {tree.activeProcesses.map((p) => {
                const isBlocked = (p.blockingPid ?? 0) > 0;
                return (
                  <tr
                    key={p.pid}
                    className={`transition-colors ${
                      isBlocked ? 'bg-red-500/5 hover:bg-red-500/10' : 'hover:bg-surface-elevated/40'
                    }`}
                  >
                    <td className="py-2.5 px-3 font-bold text-white">{p.pid}</td>
                    <td className="py-2.5 px-3 text-gray-300">{p.user}</td>
                    <td className="py-2.5 px-3">
                      <span
                        className={`px-1.5 py-0.5 rounded text-[10px] ${
                          p.state === 'idle in transaction'
                            ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                            : 'bg-blue-500/10 text-blue-400'
                        }`}
                      >
                        {p.state}
                      </span>
                    </td>
                    <td className="py-2.5 px-3 text-right text-gray-300">
                      {formatDuration(p.durationMs)}
                    </td>
                    <td className="py-2.5 px-3">
                      {p.waitEventType ? (
                        <span className="text-amber-400 font-semibold">
                          {p.waitEventType}:{p.waitEvent}
                        </span>
                      ) : (
                        <span className="text-gray-500">None</span>
                      )}
                    </td>
                    <td className="py-2.5 px-3 text-gray-300 truncate max-w-xs" title={p.query}>
                      {p.query}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
