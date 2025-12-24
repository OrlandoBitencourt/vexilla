"use client";

import { apiClient } from "@/services/api";
import { FlagMetrics } from "@/types";
import { useEffect, useState } from "react";

interface AdminActionsProps {
  onInvalidate: () => void;
}

export default function AdminActions({ onInvalidate }: AdminActionsProps) {
  const [metrics, setMetrics] = useState<FlagMetrics | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadMetrics();
  }, []);

  const loadMetrics = async () => {
    try {
      const data = await apiClient.getMetrics();
      setMetrics(data);
    } catch (error) {
      console.error("Failed to load metrics:", error);
    }
  };

  const handleInvalidateAll = async () => {
    if (!confirm("Are you sure you want to invalidate ALL flags cache?")) {
      return;
    }

    setLoading(true);
    try {
      await apiClient.invalidateAllFlags();
      alert("All flags cache invalidated successfully!");
      onInvalidate();
      await loadMetrics();
    } catch (error) {
      alert(`Failed to invalidate flags: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const handleInvalidateFlag = async (flagKey: string) => {
    setLoading(true);
    try {
      await apiClient.invalidateFlag(flagKey);
      alert(`Flag '${flagKey}' invalidated successfully!`);
      onInvalidate();
      await loadMetrics();
    } catch (error) {
      alert(`Failed to invalidate flag: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <h2 className="text-2xl font-bold mb-4 text-gray-800">⚙️ Admin Operations</h2>

      <div className="space-y-4">
        {/* Metrics */}
        {metrics && (
          <div className="bg-gray-50 rounded-lg p-4">
            <h3 className="font-semibold text-gray-700 mb-3">📊 Metrics</h3>
            <div className="grid grid-cols-2 gap-3 text-sm">
              <div>
                <span className="text-gray-600">Cache Hits:</span>
                <span className="ml-2 font-bold text-green-600">{metrics.cache_hits}</span>
              </div>
              <div>
                <span className="text-gray-600">Cache Miss:</span>
                <span className="ml-2 font-bold text-red-600">{metrics.cache_miss}</span>
              </div>
              <div>
                <span className="text-gray-600">Flags Loaded:</span>
                <span className="ml-2 font-bold text-blue-600">{metrics.flags_loaded}</span>
              </div>
              <div>
                <span className="text-gray-600">Circuit Breaker:</span>
                <span className={`ml-2 font-bold ${
                  metrics.circuit_breaker === "closed" ? "text-green-600" : "text-red-600"
                }`}>
                  {metrics.circuit_breaker.toUpperCase()}
                </span>
              </div>
            </div>
            <div className="mt-3 text-xs text-gray-500">
              Last refresh: {new Date(metrics.last_refresh).toLocaleString()}
            </div>
          </div>
        )}

        {/* Actions */}
        <div className="space-y-2">
          <button
            onClick={handleInvalidateAll}
            disabled={loading}
            className="w-full px-4 py-3 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:bg-gray-400 disabled:cursor-not-allowed font-semibold transition-colors"
          >
            {loading ? "Processing..." : "🔄 Invalidate All Flags"}
          </button>

          <div className="grid grid-cols-2 gap-2">
            <button
              onClick={() => handleInvalidateFlag("api.checkout.v2")}
              disabled={loading}
              className="px-3 py-2 bg-orange-600 text-white text-sm rounded-md hover:bg-orange-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
            >
              Invalidate Checkout V2
            </button>

            <button
              onClick={() => handleInvalidateFlag("api.checkout.rollout")}
              disabled={loading}
              className="px-3 py-2 bg-orange-600 text-white text-sm rounded-md hover:bg-orange-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
            >
              Invalidate Rollout
            </button>
          </div>
        </div>

        <div className="text-xs text-gray-500 bg-yellow-50 p-3 rounded-md border border-yellow-200">
          <div className="font-semibold mb-1">⚠️ Warning:</div>
          <p>
            Invalidating cache forces the backend to reload flags from Flagr.
            Use this after changing flag configuration in Flagr UI.
          </p>
        </div>
      </div>
    </div>
  );
}
