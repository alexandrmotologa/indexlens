import React, { useState, useMemo } from 'react';
import {
  Flame,
  AlertTriangle,
  Play,
  RotateCcw,
  Clock,
  Database,
  Layers,
  HardDrive,
  Info,
  ChevronRight,
  ChevronDown,
  GitCompare,
  Variable,
} from 'lucide-react';
import type { ExplainAnalysisResult, ExplainPlanNode, PlanDiffResult } from '../types/api';
import { PlanDiffViewer } from './PlanDiffViewer';

interface ExplainVisualizerProps {
  initialResult: ExplainAnalysisResult | null;
  initialDiff: PlanDiffResult | null;
  onAnalyzePlan: (jsonPlan: string) => Promise<ExplainAnalysisResult | null>;
  onExecuteQuery: (query: string) => Promise<ExplainAnalysisResult | null>;
  onComparePlans: (baseline: string, current: string) => Promise<PlanDiffResult | null>;
  prefilledQuery?: string;
}

export const ExplainVisualizer: React.FC<ExplainVisualizerProps> = ({
  initialResult,
  initialDiff,
  onAnalyzePlan,
  onExecuteQuery,
  onComparePlans,
  prefilledQuery = '',
}) => {
  const [activeMode, setActiveMode] = useState<'visualize' | 'diff'>('visualize');
  const [result, setResult] = useState<ExplainAnalysisResult | null>(initialResult);
  const [inputQuery, setInputQuery] = useState(prefilledQuery);
  const [inputJSON, setInputJSON] = useState('');
  const [activeInputTab, setActiveInputTab] = useState<'query' | 'json'>('query');
  const [viewMode, setViewMode] = useState<'flamegraph' | 'tree'>('flamegraph');
  const [selectedNode, setSelectedNode] = useState<ExplainPlanNode | null>(null);
  const [paramValues, setParamValues] = useState<Record<string, string>>({
    '$1': '42',
    '$2': "'2026-01-01 00:00:00'",
    '$3': "'completed'",
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Sync with initialResult when it changes
  React.useEffect(() => {
    if (initialResult) {
      setResult(initialResult);
      if (!selectedNode && initialResult.rootNode) {
        setSelectedNode(initialResult.rootNode);
      }
    }
  }, [initialResult]);

  React.useEffect(() => {
    if (prefilledQuery) {
      setInputQuery(prefilledQuery);
      setActiveInputTab('query');
    }
  }, [prefilledQuery]);

  // Detect query parameters like $1, $2, $3
  const detectedParams = useMemo(() => {
    const matches = inputQuery.match(/\$[0-9]+/g);
    if (!matches) return [];
    return Array.from(new Set(matches)).sort();
  }, [inputQuery]);

  const handleRun = async () => {
    setIsLoading(true);
    setError(null);
    try {
      if (activeInputTab === 'query') {
        if (!inputQuery.trim()) {
          setError('Please enter a SQL query to explain');
          setIsLoading(false);
          return;
        }

        // Interpolate parameter values if present
        let finalQuery = inputQuery;
        for (const p of detectedParams) {
          const val = paramValues[p] || '1';
          finalQuery = finalQuery.split(p).join(val);
        }

        const res = await onExecuteQuery(finalQuery);
        if (res) {
          setResult(res);
          setSelectedNode(res.rootNode);
        }
      } else {
        if (!inputJSON.trim()) {
          setError('Please paste an EXPLAIN JSON plan');
          setIsLoading(false);
          return;
        }
        const res = await onAnalyzePlan(inputJSON);
        if (res) {
          setResult(res);
          setSelectedNode(res.rootNode);
        }
      }
    } catch (err: any) {
      setError(err.message || 'Failed to process execution plan');
    } finally {
      setIsLoading(false);
    }
  };

  const handleLoadDemo = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await onAnalyzePlan('');
      if (res) {
        setResult(res);
        setSelectedNode(res.rootNode);
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const getNodeColor = (node: ExplainPlanNode) => {
    if (node.nodeType === 'Seq Scan') return 'bg-red-500 hover:bg-red-600 border-red-400';
    if (node.nodeType === 'Sort' && (node.sortSpaceType === 'Disk' || node.tempWrittenBlocks > 0)) {
      return 'bg-amber-600 hover:bg-amber-700 border-amber-500';
    }
    if (node.percentageOfTotal >= 40) return 'bg-yellow-600 hover:bg-yellow-700 border-yellow-500';
    if (node.nodeType.includes('Index Scan')) return 'bg-emerald-600 hover:bg-emerald-700 border-emerald-500';
    if (node.nodeType.includes('Join')) return 'bg-indigo-600 hover:bg-indigo-700 border-indigo-500';
    return 'bg-blue-600 hover:bg-blue-700 border-blue-500';
  };

  // Recursive Flamegraph Renderer
  const renderFlamegraphRow = (node: ExplainPlanNode, depth = 0) => {
    const isSelected = selectedNode?.nodeId === node.nodeId;
    return (
      <div key={node.nodeId} className="w-full space-y-1">
        <div className="w-full flex items-center">
          <div
            onClick={() => setSelectedNode(node)}
            style={{ width: `${Math.max(node.percentageOfTotal, 8)}%` }}
            className={`cursor-pointer px-2 py-1.5 rounded text-xs font-mono truncate transition-all text-white border ${
              isSelected ? 'ring-2 ring-white shadow-lg' : ''
            } ${getNodeColor(node)}`}
            title={`${node.nodeType} (${node.actualTotalTime.toFixed(1)}ms - ${node.percentageOfTotal}%)`}
          >
            <span className="font-bold">{node.nodeType}</span>
            {node.relationName && <span className="ml-1 opacity-80 font-normal">on {node.relationName}</span>}
            <span className="ml-2 opacity-90 text-[10px]">{node.actualTotalTime.toFixed(1)}ms</span>
          </div>
        </div>

        {node.plans && node.plans.length > 0 && (
          <div className="pl-4 border-l border-surface-border space-y-1">
            {node.plans.map((child) => renderFlamegraphRow(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  // Recursive Tree Node Renderer
  const TreeNode: React.FC<{ node: ExplainPlanNode; depth?: number }> = ({ node, depth = 0 }) => {
    const [collapsed, setCollapsed] = useState(false);
    const hasChildren = node.plans && node.plans.length > 0;
    const isSelected = selectedNode?.nodeId === node.nodeId;

    return (
      <div className="space-y-1">
        <div
          onClick={() => setSelectedNode(node)}
          className={`flex items-center space-x-2 py-1.5 px-3 rounded-lg text-xs font-mono cursor-pointer transition-colors border ${
            isSelected
              ? 'bg-blue-600/30 border-blue-500 text-white'
              : 'hover:bg-surface-elevated border-transparent text-gray-300'
          }`}
        >
          {hasChildren ? (
            <button
              onClick={(e) => {
                e.stopPropagation();
                setCollapsed(!collapsed);
              }}
              className="text-gray-400 hover:text-white"
            >
              {collapsed ? <ChevronRight className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
            </button>
          ) : (
            <span className="w-3.5 h-3.5" />
          )}

          <span
            className={`px-1.5 py-0.5 rounded text-[11px] font-semibold text-white ${
              node.nodeType === 'Seq Scan'
                ? 'bg-red-600'
                : node.nodeType.includes('Index')
                ? 'bg-emerald-600'
                : 'bg-surface-border'
            }`}
          >
            {node.nodeType}
          </span>

          {node.relationName && (
            <span className="text-gray-200">
              on <span className="text-blue-400 font-semibold">{node.relationName}</span>
            </span>
          )}

          <span className="text-gray-500">•</span>
          <span className="text-amber-400">{node.actualTotalTime.toFixed(1)} ms</span>
          <span className="text-gray-500">({node.percentageOfTotal}%)</span>

          {node.rowsRemovedByFilter ? (
            <span className="text-red-400 text-[10px]">
              (-{node.rowsRemovedByFilter.toLocaleString()} rows)
            </span>
          ) : null}
        </div>

        {hasChildren && !collapsed && (
          <div className="pl-6 border-l border-surface-border space-y-1">
            {node.plans!.map((child) => (
              <TreeNode key={child.nodeId} node={child} depth={depth + 1} />
            ))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="space-y-6">
      {/* Mode Switcher Header: Visualizer vs Plan Diff */}
      <div className="flex space-x-2 border-b border-surface-border pb-3">
        <button
          onClick={() => setActiveMode('visualize')}
          className={`flex items-center space-x-2 px-4 py-2 rounded-lg text-xs font-semibold transition-colors ${
            activeMode === 'visualize'
              ? 'bg-blue-600 text-white'
              : 'text-gray-400 hover:text-white hover:bg-surface-elevated'
          }`}
        >
          <Flame className="w-4 h-4" />
          <span>Execution Flamegraph &amp; Bottlenecks</span>
        </button>

        <button
          onClick={() => setActiveMode('diff')}
          className={`flex items-center space-x-2 px-4 py-2 rounded-lg text-xs font-semibold transition-colors ${
            activeMode === 'diff'
              ? 'bg-blue-600 text-white'
              : 'text-gray-400 hover:text-white hover:bg-surface-elevated'
          }`}
        >
          <GitCompare className="w-4 h-4" />
          <span>Plan Comparison &amp; Diff</span>
        </button>
      </div>

      {activeMode === 'diff' ? (
        <PlanDiffViewer initialDiff={initialDiff} onCompare={onComparePlans} />
      ) : (
        <>
          {/* Plan / Query Input Header */}
          <div className="bg-surface p-4 rounded-xl border border-surface-border space-y-3">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-2 border-b border-surface-border pb-3">
              <div className="flex space-x-2">
                <button
                  onClick={() => setActiveInputTab('query')}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
                    activeInputTab === 'query'
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-400 hover:text-white hover:bg-surface-elevated'
                  }`}
                >
                  SQL Statement (Live EXPLAIN)
                </button>
                <button
                  onClick={() => setActiveInputTab('json')}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
                    activeInputTab === 'json'
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-400 hover:text-white hover:bg-surface-elevated'
                  }`}
                >
                  Paste EXPLAIN JSON
                </button>
              </div>

              <button
                onClick={handleLoadDemo}
                className="flex items-center space-x-1.5 text-xs text-gray-400 hover:text-blue-400 transition-colors"
              >
                <RotateCcw className="w-3.5 h-3.5" />
                <span>Load Demo Plan</span>
              </button>
            </div>

            <div>
              {activeInputTab === 'query' ? (
                <div className="space-y-3">
                  <textarea
                    value={inputQuery}
                    onChange={(e) => setInputQuery(e.target.value)}
                    placeholder="Enter SELECT / UPDATE query to run EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)..."
                    rows={2}
                    className="w-full bg-surface-elevated border border-surface-border rounded-lg p-3 text-xs font-mono text-gray-200 placeholder-gray-500 focus:outline-none focus:border-blue-500"
                  />

                  {/* Parameter Interpolator */}
                  {detectedParams.length > 0 && (
                    <div className="bg-surface-elevated/60 p-3 rounded-lg border border-surface-border space-y-2">
                      <div className="text-xs text-gray-400 flex items-center space-x-1.5 font-medium">
                        <Variable className="w-3.5 h-3.5 text-blue-400" />
                        <span>Query Parameters Detected (Injected for EXPLAIN):</span>
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {detectedParams.map((p) => (
                          <div key={p} className="flex items-center space-x-1.5 bg-background px-2.5 py-1 rounded border border-surface-border text-xs font-mono">
                            <span className="text-blue-400 font-bold">{p}:</span>
                            <input
                              type="text"
                              value={paramValues[p] || ''}
                              onChange={(e) => setParamValues({ ...paramValues, [p]: e.target.value })}
                              placeholder="value"
                              className="bg-transparent text-gray-200 focus:outline-none w-28 text-xs font-mono"
                            />
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <textarea
                  value={inputJSON}
                  onChange={(e) => setInputJSON(e.target.value)}
                  placeholder='Paste JSON plan starting with [{"Plan": ...}] or {"Plan": ...}'
                  rows={2}
                  className="w-full bg-surface-elevated border border-surface-border rounded-lg p-3 text-xs font-mono text-gray-200 placeholder-gray-500 focus:outline-none focus:border-blue-500"
                />
              )}
            </div>

            {error && <div className="text-xs text-red-400 font-medium">{error}</div>}

            <div className="flex justify-end">
              <button
                onClick={handleRun}
                disabled={isLoading}
                className="flex items-center space-x-1.5 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium transition-colors shadow-lg shadow-blue-500/20 disabled:opacity-50"
              >
                <Play className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
                <span>{isLoading ? 'Analyzing...' : 'Generate Flamegraph'}</span>
              </button>
            </div>
          </div>

          {result && (
            <>
              {/* Plan Diagnostics Summary Ribbon */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                <div className="bg-surface p-4 rounded-xl border border-surface-border flex items-center space-x-3">
                  <div className="p-2.5 rounded-lg bg-blue-500/10 text-blue-400">
                    <Clock className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="text-xs text-gray-400">Execution Time</div>
                    <div className="text-lg font-mono font-bold text-white">
                      {result.totalExecutionTime.toFixed(1)} ms
                    </div>
                  </div>
                </div>

                <div className="bg-surface p-4 rounded-xl border border-surface-border flex items-center space-x-3">
                  <div className="p-2.5 rounded-lg bg-emerald-500/10 text-emerald-400">
                    <Database className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="text-xs text-gray-400">Cache Hit Ratio</div>
                    <div className="text-lg font-mono font-bold text-white">
                      {result.cacheHitRatio.toFixed(1)}%
                    </div>
                  </div>
                </div>

                <div className="bg-surface p-4 rounded-xl border border-surface-border flex items-center space-x-3">
                  <div className="p-2.5 rounded-lg bg-amber-500/10 text-amber-400">
                    <HardDrive className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="text-xs text-gray-400">Disk Read Blocks</div>
                    <div className="text-lg font-mono font-bold text-white">
                      {result.totalBuffersRead.toLocaleString()}
                    </div>
                  </div>
                </div>

                <div className="bg-surface p-4 rounded-xl border border-surface-border flex items-center space-x-3">
                  <div className="p-2.5 rounded-lg bg-purple-500/10 text-purple-400">
                    <Layers className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="text-xs text-gray-400">Planning Time</div>
                    <div className="text-lg font-mono font-bold text-white">
                      {result.planningTime.toFixed(2)} ms
                    </div>
                  </div>
                </div>
              </div>

              {/* Bottlenecks Alert Banner */}
              {result.bottlenecks.length > 0 && (
                <div className="space-y-2">
                  <div className="text-xs font-semibold text-gray-400 uppercase tracking-wider flex items-center space-x-1.5">
                    <AlertTriangle className="w-4 h-4 text-amber-400" />
                    <span>Detected Bottlenecks ({result.bottlenecks.length})</span>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                    {result.bottlenecks.map((b, idx) => (
                      <div
                        key={idx}
                        className={`p-3.5 rounded-xl border ${
                          b.severity === 'CRITICAL'
                            ? 'bg-red-500/10 border-red-500/30'
                            : 'bg-amber-500/10 border-amber-500/30'
                        }`}
                      >
                        <div className="flex items-center justify-between mb-1.5">
                          <span
                            className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                              b.severity === 'CRITICAL'
                                ? 'bg-red-500 text-white'
                                : 'bg-amber-500 text-black'
                            }`}
                          >
                            {b.type}
                          </span>
                          <span className="text-[10px] text-gray-400 font-mono">
                            Impact: {b.impactScore.toFixed(0)}%
                          </span>
                        </div>
                        <div className="text-xs font-medium text-gray-200 mb-1">{b.message}</div>
                        <div className="text-[11px] text-gray-400 italic">💡 {b.recommendation}</div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Visual Plan Explorer */}
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Visualizer Area */}
                <div className="lg:col-span-2 bg-surface p-4 rounded-xl border border-surface-border space-y-4">
                  <div className="flex items-center justify-between border-b border-surface-border pb-3">
                    <div className="flex items-center space-x-2">
                      <Flame className="w-4 h-4 text-blue-400" />
                      <span className="font-semibold text-sm text-white">Execution Plan Map</span>
                    </div>

                    <div className="flex bg-surface-elevated p-1 rounded-lg border border-surface-border text-xs">
                      <button
                        onClick={() => setViewMode('flamegraph')}
                        className={`px-3 py-1 rounded transition-colors ${
                          viewMode === 'flamegraph'
                            ? 'bg-blue-600 text-white font-medium'
                            : 'text-gray-400 hover:text-white'
                        }`}
                      >
                        Flamegraph
                      </button>
                      <button
                        onClick={() => setViewMode('tree')}
                        className={`px-3 py-1 rounded transition-colors ${
                          viewMode === 'tree'
                            ? 'bg-blue-600 text-white font-medium'
                            : 'text-gray-400 hover:text-white'
                        }`}
                      >
                        Hierarchy Tree
                      </button>
                    </div>
                  </div>

                  {/* Flamegraph or Tree View */}
                  <div className="min-h-[280px] overflow-x-auto p-2">
                    {viewMode === 'flamegraph' ? (
                      <div className="space-y-2">
                        {renderFlamegraphRow(result.rootNode)}
                      </div>
                    ) : (
                      <div className="space-y-1">
                        <TreeNode node={result.rootNode} />
                      </div>
                    )}
                  </div>

                  {/* Legend */}
                  <div className="flex flex-wrap items-center gap-3 pt-3 border-t border-surface-border text-[11px] text-gray-400">
                    <div className="flex items-center space-x-1.5">
                      <span className="w-2.5 h-2.5 rounded bg-red-500" />
                      <span>Seq Scan</span>
                    </div>
                    <div className="flex items-center space-x-1.5">
                      <span className="w-2.5 h-2.5 rounded bg-emerald-600" />
                      <span>Index Scan</span>
                    </div>
                    <div className="flex items-center space-x-1.5">
                      <span className="w-2.5 h-2.5 rounded bg-amber-600" />
                      <span>Disk Spill</span>
                    </div>
                    <div className="flex items-center space-x-1.5">
                      <span className="w-2.5 h-2.5 rounded bg-yellow-600" />
                      <span>Time Sink (&gt;40%)</span>
                    </div>
                    <div className="flex items-center space-x-1.5">
                      <span className="w-2.5 h-2.5 rounded bg-blue-600" />
                      <span>Standard Node</span>
                    </div>
                  </div>
                </div>

                {/* Node Inspector Sidebar */}
                <div className="bg-surface p-4 rounded-xl border border-surface-border space-y-4">
                  <div className="flex items-center space-x-2 border-b border-surface-border pb-3">
                    <Info className="w-4 h-4 text-blue-400" />
                    <span className="font-semibold text-sm text-white">Node Inspector</span>
                  </div>

                  {selectedNode ? (
                    <div className="space-y-4 text-xs font-mono">
                      <div>
                        <div className="text-gray-400 text-[10px] uppercase">Node Type</div>
                        <div className="text-base font-bold text-white mt-0.5">{selectedNode.nodeType}</div>
                        {selectedNode.relationName && (
                          <div className="text-blue-400 mt-0.5">
                            Table: <span className="font-semibold">{selectedNode.relationName}</span>
                          </div>
                        )}
                        {selectedNode.indexName && (
                          <div className="text-emerald-400 mt-0.5">
                            Index: <span className="font-semibold">{selectedNode.indexName}</span>
                          </div>
                        )}
                      </div>

                      <div className="grid grid-cols-2 gap-2 bg-surface-elevated p-3 rounded-lg border border-surface-border">
                        <div>
                          <div className="text-gray-400 text-[10px]">Total Time</div>
                          <div className="text-white font-bold text-sm mt-0.5">
                            {selectedNode.actualTotalTime.toFixed(1)} ms
                          </div>
                        </div>
                        <div>
                          <div className="text-gray-400 text-[10px]">Exclusive Time</div>
                          <div className="text-amber-400 font-bold text-sm mt-0.5">
                            {selectedNode.exclusiveTime.toFixed(1)} ms ({selectedNode.percentageOfTotal}%)
                          </div>
                        </div>
                      </div>

                      <div className="grid grid-cols-2 gap-2 bg-surface-elevated p-3 rounded-lg border border-surface-border">
                        <div>
                          <div className="text-gray-400 text-[10px]">Actual Rows</div>
                          <div className="text-white font-bold mt-0.5">
                            {selectedNode.actualRows.toLocaleString()}
                          </div>
                        </div>
                        <div>
                          <div className="text-gray-400 text-[10px]">Plan Rows</div>
                          <div className="text-gray-300 mt-0.5">
                            {selectedNode.planRows.toLocaleString()}
                          </div>
                        </div>
                      </div>

                      <div className="bg-surface-elevated p-3 rounded-lg border border-surface-border space-y-1.5">
                        <div className="text-gray-400 text-[10px] uppercase font-semibold">Shared Buffers</div>
                        <div className="flex justify-between">
                          <span className="text-gray-400">Cache Hit:</span>
                          <span className="text-emerald-400 font-bold">{selectedNode.sharedHitBlocks.toLocaleString()}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-400">Disk Read:</span>
                          <span className={selectedNode.sharedReadBlocks > 0 ? 'text-amber-400 font-bold' : 'text-gray-300'}>
                            {selectedNode.sharedReadBlocks.toLocaleString()}
                          </span>
                        </div>
                        {selectedNode.tempWrittenBlocks > 0 && (
                          <div className="flex justify-between">
                            <span className="text-red-400 font-semibold">Temp Spill Blocks:</span>
                            <span className="text-red-400 font-bold">{selectedNode.tempWrittenBlocks.toLocaleString()}</span>
                          </div>
                        )}
                      </div>

                      {selectedNode.filter && (
                        <div className="bg-surface-elevated p-3 rounded-lg border border-surface-border">
                          <div className="text-gray-400 text-[10px] uppercase font-semibold">Filter Predicate</div>
                          <div className="text-blue-300 text-[11px] mt-1 break-words">{selectedNode.filter}</div>
                          {selectedNode.rowsRemovedByFilter ? (
                            <div className="text-red-400 text-[11px] mt-1">
                              Discarded: {selectedNode.rowsRemovedByFilter.toLocaleString()} rows
                            </div>
                          ) : null}
                        </div>
                      )}

                      {selectedNode.sortMethod && (
                        <div className="bg-surface-elevated p-3 rounded-lg border border-surface-border">
                          <div className="text-gray-400 text-[10px] uppercase font-semibold">Sort Method</div>
                          <div className={selectedNode.sortSpaceType === 'Disk' ? 'text-red-400 font-bold mt-1' : 'text-gray-200 mt-1'}>
                            {selectedNode.sortMethod} ({selectedNode.sortSpaceUsed} kB)
                          </div>
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="text-center py-12 text-gray-500 text-xs">
                      Click on any node in the flamegraph or tree to inspect its performance metrics.
                    </div>
                  )}
                </div>
              </div>
            </>
          )}
        </>
      )}
    </div>
  );
};
