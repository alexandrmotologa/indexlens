import { useState } from 'react';
import type { TuningReport, TuningProfile } from '../types/api';
import { Cpu, HardDrive, Database, Sliders, Copy, Check, RefreshCw, AlertTriangle, CheckCircle2 } from 'lucide-react';

interface ConfigTunerProps {
  report: TuningReport | null;
  onProfileChange: (profile: TuningProfile) => void;
}

export function ConfigTuner({ report, onProfileChange }: ConfigTunerProps) {
  const [copied, setCopied] = useState(false);
  const [ram, setRam] = useState(report?.profile.totalRamGb || 32);
  const [vcpus, setVcpus] = useState(report?.profile.vcpus || 8);
  const [disk, setDisk] = useState<'NVMe' | 'SSD' | 'HDD'>(report?.profile.diskType || 'NVMe');
  const [workload, setWorkload] = useState<'OLTP' | 'WEB' | 'DW' | 'MIXED'>(report?.profile.workloadType || 'OLTP');

  const handleApply = (newRam = ram, newCpu = vcpus, newDisk = disk, newWorkload = workload) => {
    onProfileChange({
      totalRamGb: newRam,
      vcpus: newCpu,
      diskType: newDisk,
      workloadType: newWorkload,
    });
  };

  const handleCopyScript = () => {
    if (report?.alterScript) {
      navigator.clipboard.writeText(report.alterScript);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (!report) {
    return (
      <div className="flex h-64 items-center justify-center rounded-xl border border-border bg-card">
        <RefreshCw className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Hardware Profile Controls */}
      <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
        <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-semibold text-foreground">
              <Sliders className="h-5 w-5 text-indigo-400" />
              PostgreSQL Configuration & Hardware Tuner
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              Tailor memory parameters, parallelism, and I/O heuristics specifically to your host's physical profile.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-400 border border-emerald-500/20">
              <CheckCircle2 className="h-3.5 w-3.5" />
              {report.optimalCount} Optimal
            </span>
            <span className="inline-flex items-center gap-1.5 rounded-full bg-amber-500/10 px-3 py-1 text-xs font-medium text-amber-400 border border-amber-500/20">
              <AlertTriangle className="h-3.5 w-3.5" />
              {report.attentionCount} Tuning Candidates
            </span>
          </div>
        </div>

        {/* Profile Sliders and Pickers */}
        <div className="mt-6 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {/* RAM */}
          <div className="rounded-lg border border-border/60 bg-muted/30 p-4">
            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
                <Database className="h-4 w-4 text-sky-400" />
                SYSTEM RAM
              </label>
              <span className="font-mono text-sm font-semibold text-foreground">{ram} GB</span>
            </div>
            <input
              type="range"
              min="4"
              max="128"
              step="4"
              value={ram}
              onChange={(e) => {
                const val = Number(e.target.value);
                setRam(val);
                handleApply(val, vcpus, disk, workload);
              }}
              className="mt-3 w-full accent-indigo-500"
            />
            <div className="mt-1 flex justify-between text-[10px] text-muted-foreground">
              <span>4 GB</span>
              <span>32 GB</span>
              <span>128 GB</span>
            </div>
          </div>

          {/* vCPUs */}
          <div className="rounded-lg border border-border/60 bg-muted/30 p-4">
            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
                <Cpu className="h-4 w-4 text-emerald-400" />
                VCPU CORES
              </label>
              <span className="font-mono text-sm font-semibold text-foreground">{vcpus} Cores</span>
            </div>
            <input
              type="range"
              min="2"
              max="64"
              step="2"
              value={vcpus}
              onChange={(e) => {
                const val = Number(e.target.value);
                setVcpus(val);
                handleApply(ram, val, disk, workload);
              }}
              className="mt-3 w-full accent-indigo-500"
            />
            <div className="mt-1 flex justify-between text-[10px] text-muted-foreground">
              <span>2</span>
              <span>16</span>
              <span>64</span>
            </div>
          </div>

          {/* Storage Type */}
          <div className="rounded-lg border border-border/60 bg-muted/30 p-4">
            <label className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
              <HardDrive className="h-4 w-4 text-purple-400" />
              STORAGE TYPE
            </label>
            <div className="mt-3 grid grid-cols-3 gap-1.5">
              {(['NVMe', 'SSD', 'HDD'] as const).map((t) => (
                <button
                  key={t}
                  onClick={() => {
                    setDisk(t);
                    handleApply(ram, vcpus, t, workload);
                  }}
                  className={`rounded-md px-2 py-1.5 text-xs font-medium transition-colors ${
                    disk === t
                      ? 'bg-indigo-600 text-white font-semibold shadow-sm'
                      : 'bg-muted/60 text-muted-foreground hover:bg-muted hover:text-foreground'
                  }`}
                >
                  {t}
                </button>
              ))}
            </div>
          </div>

          {/* Workload Type */}
          <div className="rounded-lg border border-border/60 bg-muted/30 p-4">
            <label className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
              <Database className="h-4 w-4 text-amber-400" />
              WORKLOAD TYPE
            </label>
            <div className="mt-3 grid grid-cols-2 gap-1.5">
              {(['OLTP', 'WEB', 'DW', 'MIXED'] as const).map((w) => (
                <button
                  key={w}
                  onClick={() => {
                    setWorkload(w);
                    handleApply(ram, vcpus, disk, w);
                  }}
                  className={`rounded-md px-2 py-1.5 text-xs font-medium transition-colors ${
                    workload === w
                      ? 'bg-indigo-600 text-white font-semibold shadow-sm'
                      : 'bg-muted/60 text-muted-foreground hover:bg-muted hover:text-foreground'
                  }`}
                >
                  {w}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Settings Comparison Table */}
      <div className="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
        <div className="border-b border-border bg-muted/20 px-6 py-4">
          <h3 className="font-semibold text-foreground">Kernel & Memory Parameters</h3>
          <p className="text-xs text-muted-foreground">
            Current configuration versus hardware-calibrated PostgreSQL targets.
          </p>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="border-b border-border bg-muted/40 text-xs uppercase text-muted-foreground">
              <tr>
                <th className="px-6 py-3 font-semibold">Parameter</th>
                <th className="px-6 py-3 font-semibold">Current</th>
                <th className="px-6 py-3 font-semibold">Recommended</th>
                <th className="px-6 py-3 font-semibold">Category</th>
                <th className="px-6 py-3 font-semibold">Restart Required</th>
                <th className="px-6 py-3 font-semibold">Reasoning</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {report.settings.map((s) => (
                <tr key={s.name} className="hover:bg-muted/20 transition-colors">
                  <td className="px-6 py-3.5 font-mono text-xs font-semibold text-foreground">
                    {s.name}
                  </td>
                  <td className="px-6 py-3.5 font-mono text-xs text-muted-foreground">
                    {s.currentValue}
                  </td>
                  <td className="px-6 py-3.5 font-mono text-xs font-bold text-indigo-400">
                    {s.recommendedValue}
                  </td>
                  <td className="px-6 py-3.5">
                    <span className="rounded-md bg-muted px-2 py-1 text-[11px] font-medium text-foreground">
                      {s.category}
                    </span>
                  </td>
                  <td className="px-6 py-3.5 text-xs">
                    {s.restartRequired ? (
                      <span className="text-rose-400 font-medium">Yes (Server restart)</span>
                    ) : (
                      <span className="text-emerald-400 font-medium">No (Reload only)</span>
                    )}
                  </td>
                  <td className="px-6 py-3.5 text-xs text-muted-foreground max-w-md">
                    {s.reason}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Copyable ALTER SYSTEM Script */}
      <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-semibold text-foreground">Deployment Script (`ALTER SYSTEM`)</h3>
            <p className="text-xs text-muted-foreground">
              Execute with superuser credentials. Changes persist into <code className="text-indigo-400">postgresql.auto.conf</code>.
            </p>
          </div>
          <button
            onClick={handleCopyScript}
            className="flex items-center gap-1.5 rounded-lg bg-indigo-600 px-4 py-2 text-xs font-semibold text-white shadow hover:bg-indigo-500 transition-colors"
          >
            {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
            {copied ? 'Copied to Clipboard' : 'Copy Tuning Script'}
          </button>
        </div>

        <div className="mt-4 rounded-lg border border-border/80 bg-zinc-950 p-4">
          <pre className="overflow-x-auto font-mono text-xs text-emerald-400 leading-relaxed">
            {report.alterScript}
          </pre>
        </div>
      </div>
    </div>
  );
}
