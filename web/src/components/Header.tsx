import React from 'react';
import { Database, Flame, Zap, ShieldAlert, Layers, RefreshCw, Lock, Download } from 'lucide-react';
import type { SystemOverview } from '../types/api';

interface HeaderProps {
  overview: SystemOverview | null;
  activeTab: 'queries' | 'explain' | 'advisor' | 'locks' | 'bloat' | 'cache';
  setActiveTab: (tab: 'queries' | 'explain' | 'advisor' | 'locks' | 'bloat' | 'cache') => void;
  onRefresh: () => void;
  onExportReport: () => void;
  isRefreshing: boolean;
}

export const Header: React.FC<HeaderProps> = ({
  overview,
  activeTab,
  setActiveTab,
  onRefresh,
  onExportReport,
  isRefreshing,
}) => {
  return (
    <header className="border-b border-surface-border bg-surface sticky top-0 z-30">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo & Database Name */}
          <div className="flex items-center space-x-3">
            <div className="w-9 h-9 rounded-lg bg-blue-600 flex items-center justify-center shadow-lg shadow-blue-500/20">
              <Database className="w-5 h-5 text-white" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <span className="font-bold text-lg text-white tracking-tight">IndexLens</span>
                {overview?.demoMode && (
                  <span className="px-2 py-0.5 text-xs font-semibold rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
                    DEMO MODE
                  </span>
                )}
                {!overview?.demoMode && overview?.connected && (
                  <span className="px-2 py-0.5 text-xs font-semibold rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                    LIVE
                  </span>
                )}
              </div>
              <div className="text-xs text-gray-400 font-mono">
                {overview?.databaseName || 'PostgreSQL Engine'}
              </div>
            </div>
          </div>

          {/* Navigation Tabs */}
          <nav className="flex space-x-1">
            <button
              onClick={() => setActiveTab('queries')}
              className={`flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'queries'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-300 hover:bg-surface-elevated hover:text-white'
              }`}
            >
              <Zap className="w-4 h-4" />
              <span>Slow Queries</span>
            </button>

            <button
              onClick={() => setActiveTab('explain')}
              className={`flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'explain'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-300 hover:bg-surface-elevated hover:text-white'
              }`}
            >
              <Flame className="w-4 h-4" />
              <span>Explain Flamegraph</span>
            </button>

            <button
              onClick={() => setActiveTab('advisor')}
              className={`flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'advisor'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-300 hover:bg-surface-elevated hover:text-white'
              }`}
            >
              <Layers className="w-4 h-4" />
              <span>Index Advisor</span>
              {overview && (overview.underIndexedTablesCount > 0 || overview.unusedIndexesCount > 0) && (
                <span className="ml-1 px-1.5 py-0.2 text-xs rounded-full bg-amber-500 text-black font-bold">
                  {overview.underIndexedTablesCount + overview.unusedIndexesCount}
                </span>
              )}
            </button>

            <button
              onClick={() => setActiveTab('locks')}
              className={`flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'locks'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-300 hover:bg-surface-elevated hover:text-white'
              }`}
            >
              <Lock className="w-4 h-4" />
              <span>Locks &amp; Activity</span>
              {overview && overview.activeLocksCount > 0 && (
                <span className="ml-1 px-1.5 py-0.2 text-xs rounded-full bg-red-500 text-white font-bold animate-pulse">
                  !
                </span>
              )}
            </button>

            <button
              onClick={() => setActiveTab('bloat')}
              className={`flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'bloat'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-300 hover:bg-surface-elevated hover:text-white'
              }`}
            >
              <ShieldAlert className="w-4 h-4" />
              <span>Bloat &amp; Autovacuum</span>
            </button>

            <button
              onClick={() => setActiveTab('cache')}
              className={`flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'cache'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-300 hover:bg-surface-elevated hover:text-white'
              }`}
            >
              <Database className="w-4 h-4" />
              <span>Buffer Cache</span>
            </button>
          </nav>

          {/* Export Report & Refresh */}
          <div className="flex items-center space-x-3">
            <button
              onClick={onExportReport}
              className="hidden sm:flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-surface-elevated hover:bg-surface-border text-xs font-medium text-gray-200 border border-surface-border transition-colors shadow-sm"
              title="Download Incident / Audit Markdown Report"
            >
              <Download className="w-3.5 h-3.5 text-blue-400" />
              <span>Export Report</span>
            </button>

            <button
              onClick={onRefresh}
              disabled={isRefreshing}
              className="p-2 text-gray-400 hover:text-white hover:bg-surface-elevated rounded-lg transition-colors border border-surface-border"
              title="Refresh Telemetry"
            >
              <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin text-blue-400' : ''}`} />
            </button>
          </div>
        </div>
      </div>
    </header>
  );
};
