import type { ConnectionPoolHealth } from '../types/api';
import { Network, Activity, Clock, AlertTriangle, ShieldCheck, Server, RefreshCw } from 'lucide-react';

interface ConnectionPoolViewProps {
  health: ConnectionPoolHealth | null;
}

export function ConnectionPoolView({ health }: ConnectionPoolViewProps) {
  if (!health) {
    return (
      <div className="flex h-64 items-center justify-center rounded-xl border border-border bg-card">
        <RefreshCw className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  const saturationColor =
    health.utilizationPercent >= 85
      ? 'text-rose-400 bg-rose-500/10 border-rose-500/30'
      : health.utilizationPercent >= 70
      ? 'text-amber-400 bg-amber-500/10 border-amber-500/30'
      : 'text-emerald-400 bg-emerald-500/10 border-emerald-500/30';

  const activePct = health.totalConnections > 0 ? (health.activeCount / health.totalConnections) * 100 : 0;
  const idlePct = health.totalConnections > 0 ? (health.idleCount / health.totalConnections) * 100 : 0;
  const idleInTxPct = health.totalConnections > 0 ? (health.idleInTxCount / health.totalConnections) * 100 : 0;

  return (
    <div className="space-y-6">
      {/* Overview Banner */}
      <div className="flex flex-col gap-4 rounded-xl border border-border bg-card p-6 shadow-sm md:flex-row md:items-center md:justify-between">
        <div>
          <h2 className="flex items-center gap-2 text-lg font-semibold text-foreground">
            <Network className="h-5 w-5 text-indigo-400" />
            Connection Pool & Saturation Health
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Monitor client session states, idle-in-transaction leaks, and PgBouncer proxy sizing.
          </p>
        </div>

        <div className="flex items-center gap-4">
          <div className="text-right">
            <span className="text-xs uppercase tracking-wider text-muted-foreground font-semibold">Saturation</span>
            <div className="text-2xl font-bold font-mono text-foreground">{health.utilizationPercent.toFixed(1)}%</div>
          </div>
          <div className={`flex h-12 px-3 items-center justify-center rounded-xl border font-mono text-sm font-bold ${saturationColor}`}>
            {health.status}
          </div>
        </div>
      </div>

      {/* Warnings Banner if any */}
      {health.warnings.length > 0 && (
        <div className="rounded-xl border border-amber-500/30 bg-amber-500/10 p-4">
          <div className="flex items-start gap-3">
            <AlertTriangle className="h-5 w-5 text-amber-400 shrink-0 mt-0.5" />
            <div className="space-y-1">
              <h4 className="text-sm font-semibold text-amber-400">Connection Saturation Warnings</h4>
              {health.warnings.map((w, i) => (
                <p key={i} className="text-xs text-amber-300/90 leading-relaxed">
                  • {w}
                </p>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Metrics Grid */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {/* Total Connections */}
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Total Connections</span>
            <Server className="h-4 w-4 text-indigo-400" />
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono text-foreground">{health.totalConnections}</span>
            <span className="text-xs text-muted-foreground font-mono">/ {health.maxConnections} max</span>
          </div>
          <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-muted">
            <div
              className={`h-full rounded-full transition-all duration-500 ${
                health.utilizationPercent >= 85 ? 'bg-rose-500' : health.utilizationPercent >= 70 ? 'bg-amber-500' : 'bg-emerald-500'
              }`}
              style={{ width: `${Math.min(100, health.utilizationPercent)}%` }}
            />
          </div>
        </div>

        {/* Active Connections */}
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Active Executing</span>
            <Activity className="h-4 w-4 text-emerald-400" />
          </div>
          <div className="mt-3 text-2xl font-bold font-mono text-emerald-400">{health.activeCount}</div>
          <p className="mt-2 text-xs text-muted-foreground">Actively executing backend queries</p>
        </div>

        {/* Idle Connections */}
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Idle Sockets</span>
            <Clock className="h-4 w-4 text-sky-400" />
          </div>
          <div className="mt-3 text-2xl font-bold font-mono text-sky-400">{health.idleCount}</div>
          <p className="mt-2 text-xs text-muted-foreground">Connected sockets waiting for requests</p>
        </div>

        {/* Idle in Transaction */}
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Idle In Transaction</span>
            <AlertTriangle className={`h-4 w-4 ${health.idleInTxCount > 0 ? 'text-rose-400' : 'text-muted-foreground'}`} />
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className={`text-2xl font-bold font-mono ${health.idleInTxCount > 0 ? 'text-rose-400' : 'text-foreground'}`}>
              {health.idleInTxCount}
            </span>
            {health.maxIdleInTxDuration > 0 && (
              <span className="text-xs font-mono text-rose-400">({health.maxIdleInTxDuration.toFixed(0)}s max)</span>
            )}
          </div>
          <p className="mt-2 text-xs text-muted-foreground">Holding locks & preventing vacuum cleanup</p>
        </div>
      </div>

      {/* State Composition Breakdown */}
      <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
        <h3 className="font-semibold text-foreground">Connection State Distribution</h3>
        <p className="text-xs text-muted-foreground mt-0.5">Ratio of working processes versus idle connections.</p>

        <div className="mt-4 flex h-4 w-full overflow-hidden rounded-full bg-muted">
          <div style={{ width: `${activePct}%` }} className="bg-emerald-500 transition-all duration-300" title={`Active: ${health.activeCount}`} />
          <div style={{ width: `${idlePct}%` }} className="bg-sky-500 transition-all duration-300" title={`Idle: ${health.idleCount}`} />
          <div style={{ width: `${idleInTxPct}%` }} className="bg-rose-500 transition-all duration-300" title={`Idle in Tx: ${health.idleInTxCount}`} />
        </div>

        <div className="mt-3 flex flex-wrap items-center gap-6 text-xs font-medium">
          <div className="flex items-center gap-2">
            <span className="h-3 w-3 rounded-full bg-emerald-500" />
            <span className="text-foreground">Active: {health.activeCount} ({activePct.toFixed(1)}%)</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="h-3 w-3 rounded-full bg-sky-500" />
            <span className="text-foreground">Idle: {health.idleCount} ({idlePct.toFixed(1)}%)</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="h-3 w-3 rounded-full bg-rose-500" />
            <span className="text-foreground">Idle in Transaction: {health.idleInTxCount} ({idleInTxPct.toFixed(1)}%)</span>
          </div>
        </div>
      </div>

      {/* PgBouncer Sweet Spot Guide */}
      <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
        <div className="flex items-start gap-4">
          <div className="rounded-lg bg-indigo-500/10 p-3 text-indigo-400 border border-indigo-500/20">
            <ShieldCheck className="h-6 w-6" />
          </div>
          <div className="space-y-2 flex-1">
            <h3 className="font-semibold text-foreground">PgBouncer Architecture & Sizing Recommendation</h3>
            <p className="text-xs text-muted-foreground leading-relaxed">
              PostgreSQL allocates a dedicated OS process for every direct connection. Beyond{' '}
              <span className="font-semibold text-indigo-400 font-mono">(CPU Cores × 2) + Disks</span>, connection overhead triggers severe CPU context switching stalls and L3 cache thrashing.
            </p>
            <div className="rounded-lg border border-border/80 bg-zinc-950 p-3 font-mono text-xs text-indigo-300">
              Optimal PostgreSQL direct backend capacity: <span className="text-emerald-400 font-bold">{health.optimalPoolSize} connections</span>.
            </div>
            <p className="text-xs text-muted-foreground">{health.pgBouncerAdvice}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
