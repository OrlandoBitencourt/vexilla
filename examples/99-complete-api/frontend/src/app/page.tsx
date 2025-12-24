"use client";

import { useEffect, useState } from "react";
import UserSimulator from "@/components/UserSimulator";
import FlagStatus from "@/components/FlagStatus";
import RolloutIndicator from "@/components/RolloutIndicator";
import CheckoutDemo from "@/components/CheckoutDemo";
import AdminActions from "@/components/AdminActions";
import { apiClient } from "@/services/api";
import { UserContext, FlagsSnapshot, CheckoutResponse } from "@/types";

const DEFAULT_CONTEXT: UserContext = {
  userId: "user-1",
  cpf: "12345678909",
  role: "user",
  country: "BR",
};

export default function Home() {
  const [context, setContext] = useState<UserContext>(DEFAULT_CONTEXT);
  const [flagsData, setFlagsData] = useState<FlagsSnapshot | null>(null);
  const [checkoutData, setCheckoutData] = useState<CheckoutResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [checkoutLoading, setCheckoutLoading] = useState(false);
  const [apiStatus, setApiStatus] = useState<"connected" | "disconnected" | "checking">("checking");

  useEffect(() => {
    checkHealth();
  }, []);

  useEffect(() => {
    loadFlags();
  }, [context]);

  const checkHealth = async () => {
    try {
      await apiClient.healthCheck();
      setApiStatus("connected");
    } catch (error) {
      setApiStatus("disconnected");
    }
  };

  const loadFlags = async () => {
    setLoading(true);
    try {
      const data = await apiClient.getFlagsSnapshot(context);
      setFlagsData(data);
    } catch (error) {
      console.error("Failed to load flags:", error);
      setApiStatus("disconnected");
    } finally {
      setLoading(false);
    }
  };

  const handleCheckout = async () => {
    setCheckoutLoading(true);
    setCheckoutData(null);
    try {
      const data = await apiClient.checkout(context);
      setCheckoutData(data);
    } catch (error) {
      console.error("Checkout failed:", error);
      alert(`Checkout failed: ${error}`);
    } finally {
      setCheckoutLoading(false);
    }
  };

  const handleContextChange = (newContext: UserContext) => {
    setContext(newContext);
    setCheckoutData(null);
  };

  return (
    <main className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 p-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-gray-800 mb-2">
            🏴 Vexilla Demo
          </h1>
          <p className="text-gray-600">
            Real-world demonstration of feature flags with deterministic rollout
          </p>
          <div className="mt-2 flex items-center gap-2">
            <div className={`w-3 h-3 rounded-full ${
              apiStatus === "connected" ? "bg-green-500" :
              apiStatus === "disconnected" ? "bg-red-500" : "bg-yellow-500"
            } animate-pulse`} />
            <span className="text-sm text-gray-600">
              API Status: <span className="font-semibold">
                {apiStatus === "connected" ? "Connected" :
                 apiStatus === "disconnected" ? "Disconnected" : "Checking..."}
              </span>
            </span>
          </div>
        </div>

        {apiStatus === "disconnected" && (
          <div className="mb-6 bg-red-50 border-2 border-red-300 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <span className="text-2xl">⚠️</span>
              <div>
                <div className="font-bold text-red-800 mb-1">Backend API Not Available</div>
                <p className="text-sm text-red-700 mb-2">
                  Make sure the backend is running on http://localhost:8080
                </p>
                <button
                  onClick={checkHealth}
                  className="px-4 py-2 bg-red-600 text-white text-sm rounded-md hover:bg-red-700"
                >
                  Retry Connection
                </button>
              </div>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column - User Simulator */}
          <div className="space-y-6">
            <UserSimulator
              initialContext={context}
              onContextChange={handleContextChange}
            />

            {/* Admin Actions (show if user is admin) */}
            {context.role === "admin" && (
              <AdminActions onInvalidate={loadFlags} />
            )}
          </div>

          {/* Middle Column - Flags Status */}
          <div className="space-y-6">
            <div className="bg-white rounded-lg shadow-md p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-2xl font-bold text-gray-800">🚩 Flags Status</h2>
                <button
                  onClick={loadFlags}
                  disabled={loading}
                  className="px-3 py-1 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 disabled:bg-gray-400"
                >
                  {loading ? "..." : "Refresh"}
                </button>
              </div>

              {loading && !flagsData ? (
                <div className="text-center py-8">
                  <div className="animate-spin rounded-full h-12 w-12 border-b-4 border-blue-600 mx-auto"></div>
                </div>
              ) : flagsData ? (
                <div className="space-y-3">
                  <FlagStatus
                    name="api.checkout.v2"
                    value={flagsData.flags["api.checkout.v2"]}
                  />
                  <FlagStatus
                    name="api.checkout.rollout"
                    value={flagsData.flags["api.checkout.rollout"]}
                    type="number"
                  />
                  <FlagStatus
                    name="api.rate_limit.enabled"
                    value={flagsData.flags["api.rate_limit.enabled"]}
                  />
                  <FlagStatus
                    name="api.kill_switch"
                    value={flagsData.flags["api.kill_switch"]}
                  />
                  <FlagStatus
                    name="frontend.new_ui"
                    value={flagsData.flags["frontend.new_ui"]}
                  />
                  <FlagStatus
                    name="frontend.beta_banner"
                    value={flagsData.flags["frontend.beta_banner"]}
                  />
                </div>
              ) : null}
            </div>

            {/* Rollout Indicator */}
            {flagsData && (
              <RolloutIndicator
                cpf={context.cpf}
                bucket={flagsData.context.bucket}
                rollout={flagsData.flags["api.checkout.rollout"]}
                enabled={flagsData.flags["api.checkout.v2"]}
              />
            )}
          </div>

          {/* Right Column - Checkout Demo */}
          <div className="space-y-6">
            <div>
              <button
                onClick={handleCheckout}
                disabled={checkoutLoading || apiStatus === "disconnected"}
                className="w-full px-6 py-4 bg-gradient-to-r from-blue-600 to-indigo-600 text-white font-bold text-lg rounded-lg hover:from-blue-700 hover:to-indigo-700 disabled:from-gray-400 disabled:to-gray-400 disabled:cursor-not-allowed shadow-lg transition-all transform hover:scale-105 active:scale-95"
              >
                {checkoutLoading ? "Processing..." : "🛒 Test Checkout"}
              </button>
            </div>

            <CheckoutDemo
              checkoutData={checkoutData}
              loading={checkoutLoading}
            />

            {/* Info Box */}
            {flagsData?.flags["frontend.beta_banner"] && (
              <div className="bg-gradient-to-r from-purple-50 to-pink-50 border-2 border-purple-300 rounded-lg p-4">
                <div className="flex items-start gap-3">
                  <span className="text-2xl">🎉</span>
                  <div>
                    <div className="font-bold text-purple-800 mb-1">Beta Features Available!</div>
                    <p className="text-sm text-purple-700">
                      You're seeing this banner because the <code className="bg-purple-200 px-1 rounded">frontend.beta_banner</code> flag is enabled.
                    </p>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-center text-sm text-gray-500">
          <p>
            Built with <strong>Vexilla</strong> - Feature flags for modern applications
          </p>
          <p className="mt-1">
            Backend: Go + Gin | Frontend: Next.js + TypeScript | Flags: Flagr
          </p>
        </div>
      </div>
    </main>
  );
}
