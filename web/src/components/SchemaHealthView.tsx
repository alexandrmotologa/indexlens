import { useState } from 'react';
import type { SchemaHealthReport } from '../types/api';
import { ShieldCheck, AlertOctagon, Copy, Check, Key, Layers, ArrowRight, RefreshCw } from 'lucide-react';

interface SchemaHealthViewProps {
  report: SchemaHealthReport | null;
}

export function SchemaHealthView({ report }: SchemaHealthViewProps) {
  const [copiedIndex, setCopiedIndex] = useState<string | null>(null);

  const handleCopy = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedIndex(id);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  if (!report) {
    return (
      <div className="flex h-64 items-center justify-center rounded-xl border border-border bg-card">
        <RefreshCw className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  const scoreColor =
    report.healthScore >= 80
      ? 'text-emerald-400 border-emerald-500/30 bg-emerald-500/10'
      : report.healthScore >= 50
      ? 'text-amber-400 border-amber-500/30 bg-amber-500/10'
      : 'text-rose-400 border-rose-500/30 bg-rose-500/10';

  return (
    <div className="space-y-6">
      {/* Header and Health Score Banner */}
      <div className="flex flex-col gap-4 rounded-xl border border-border bg-card p-6 shadow-sm md:flex-row md:items-center md:justify-between">
        <div>
          <h2 className="flex items-center gap-2 text-lg font-semibold text-foreground">
            <ShieldCheck className="h-5 w-5 text-indigo-400" />
            Schema Health & Relational Anti-Pattern Auditor
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Identify unindexed foreign key cascades, missing primary keys, and heavy out-of-line TOAST relations.
          </p>
        </div>

        <div className="flex items-center gap-4">
          <div className="text-right">
            <span className="text-xs uppercase tracking-wider text-muted-foreground font-semibold">Schema Score</span>
            <div className="text-2xl font-bold font-mono text-foreground">{report.healthScore} / 100</div>
          </div>
          <div className={`flex h-12 w-12 items-center justify-center rounded-xl border font-mono text-lg font-bold ${scoreColor}`}>
            {report.healthScore}
          </div>
        </div>
      </div>

      {/* Unindexed Foreign Keys */}
      <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
        <div className="flex items-center justify-between border-b border-border pb-4">
          <div>
            <h3 className="flex items-center gap-2 font-semibold text-foreground">
              <AlertOctagon className="h-4 w-4 text-rose-400" />
              Unindexed Foreign Keys ({report.unindexedFks.length})
            </h3>
            <p className="text-xs text-muted-foreground mt-0.5">
              Foreign keys without child indexes trigger table-level ShareRowExclusiveLock or full sequential scans during parent deletes or updates.
            </p>
          </div>
        </div>

        {report.unindexedFks.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            ✓ Excellent! All foreign key constraints have supporting indexes on child tables.
          </div>
        ) : (
          <div className="mt-4 space-y-4">
            {report.unindexedFks.map((fk) => (
              <div
                key={fk.constraintName}
                className="rounded-lg border border-border/70 bg-muted/20 p-4 transition-colors hover:border-border"
              >
                <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                  <div className="flex items-center gap-2">
                    <span className="rounded bg-rose-500/10 px-2 py-0.5 font-mono text-xs font-semibold text-rose-400 border border-rose-500/20">
                      {fk.severity}
                    </span>
                    <span className="font-mono text-sm font-semibold text-foreground">{fk.childTable}</span>
                    <ArrowRight className="h-3.5 w-3.5 text-muted-foreground" />
                    <span className="font-mono text-sm text-muted-foreground">{fk.parentTable}</span>
                    <span className="text-xs text-muted-foreground">({fk.columns.join(', ')})</span>
                  </div>

                  <button
                    onClick={() => handleCopy(fk.constraintName, fk.suggestedDdl)}
                    className="flex items-center gap-1.5 self-start sm:self-auto rounded-md bg-indigo-600/90 hover:bg-indigo-600 px-3 py-1.5 text-xs font-semibold text-white shadow-sm transition-colors"
                  >
                    {copiedIndex === fk.constraintName ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
                    {copiedIndex === fk.constraintName ? 'Copied' : 'Copy DDL'}
                  </button>
                </div>

                <p className="mt-2 text-xs text-muted-foreground">{fk.reason}</p>

                <div className="mt-3 rounded bg-zinc-950 px-3 py-2">
                  <code className="font-mono text-xs text-emerald-400">{fk.suggestedDdl}</code>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Grid: Missing PKs & TOAST Bloat */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Missing PKs */}
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
          <div className="flex items-center gap-2 border-b border-border pb-4">
            <Key className="h-4 w-4 text-amber-400" />
            <div>
              <h3 className="font-semibold text-foreground">Tables Missing Primary Key ({report.missingPks.length})</h3>
              <p className="text-xs text-muted-foreground">Prevents logical replication and degrades tuple identity.</p>
            </div>
          </div>

          {report.missingPks.length === 0 ? (
            <div className="py-8 text-center text-sm text-muted-foreground">
              ✓ All analyzed tables have a declared primary key.
            </div>
          ) : (
            <div className="mt-4 space-y-3">
              {report.missingPks.map((pk) => (
                <div key={`${pk.schemaName}.${pk.tableName}`} className="rounded-lg border border-border/70 bg-muted/20 p-3.5">
                  <div className="flex items-center justify-between">
                    <div>
                      <span className="font-mono text-xs font-semibold text-foreground">
                        {pk.schemaName}.{pk.tableName}
                      </span>
                      <div className="text-[11px] text-muted-foreground mt-0.5">
                        ~{pk.estimatedRows.toLocaleString()} rows • {pk.totalSizePretty}
                      </div>
                    </div>
                    <button
                      onClick={() => handleCopy(pk.tableName, pk.suggestedDdl)}
                      className="flex items-center gap-1 rounded bg-muted hover:bg-muted/80 px-2.5 py-1 text-xs font-medium text-foreground transition-colors"
                    >
                      {copiedIndex === pk.tableName ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                      {copiedIndex === pk.tableName ? 'Copied' : 'Copy'}
                    </button>
                  </div>
                  <div className="mt-2 rounded bg-zinc-950 px-2.5 py-1.5 font-mono text-[11px] text-emerald-400">
                    {pk.suggestedDdl}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* TOAST Relations */}
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
          <div className="flex items-center gap-2 border-b border-border pb-4">
            <Layers className="h-4 w-4 text-purple-400" />
            <div>
              <h3 className="font-semibold text-foreground">TOAST Out-of-Line Storage ({report.toastRelations.length})</h3>
              <p className="text-xs text-muted-foreground">Oversized variable-length attributes (JSONB, TEXT, BYTEA).</p>
            </div>
          </div>

          {report.toastRelations.length === 0 ? (
            <div className="py-8 text-center text-sm text-muted-foreground">
              ✓ No disproportionate TOAST table bloat detected.
            </div>
          ) : (
            <div className="mt-4 space-y-3">
              {report.toastRelations.map((t) => (
                <div key={t.tableName} className="rounded-lg border border-border/70 bg-muted/20 p-3.5">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-xs font-semibold text-foreground">{t.tableName}</span>
                    <span className="font-mono text-xs font-bold text-purple-400">{t.toastRatio.toFixed(1)}% TOAST</span>
                  </div>

                  {/* Ratio bar */}
                  <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-muted">
                    <div
                      className="h-full bg-purple-500 rounded-full"
                      style={{ width: `${Math.min(100, t.toastRatio)}%` }}
                    />
                  </div>

                  <div className="mt-2 flex justify-between text-[11px] text-muted-foreground font-mono">
                    <span>Main: {t.mainSizePretty}</span>
                    <span>TOAST: {t.toastSizePretty}</span>
                  </div>

                  <p className="mt-2 text-[11px] text-muted-foreground leading-relaxed">{t.recommendation}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
