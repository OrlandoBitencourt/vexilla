"use client";

import { CheckoutResponse } from "@/types";
import { useEffect, useState } from "react";

interface CheckoutDemoProps {
  checkoutData: CheckoutResponse | null;
  loading: boolean;
}

export default function CheckoutDemo({ checkoutData, loading }: CheckoutDemoProps) {
  const [animationKey, setAnimationKey] = useState(0);

  useEffect(() => {
    if (checkoutData) {
      setAnimationKey(prev => prev + 1);
    }
  }, [checkoutData]);

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow-md p-12 text-center">
        <div className="animate-spin rounded-full h-16 w-16 border-b-4 border-blue-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">Processing checkout...</p>
      </div>
    );
  }

  if (!checkoutData) {
    return (
      <div className="bg-gray-100 rounded-lg p-12 text-center">
        <div className="text-gray-400 text-6xl mb-4">🛒</div>
        <p className="text-gray-600">Click "Test Checkout" to see the demo</p>
      </div>
    );
  }

  const isV2 = checkoutData.version === "v2";
  const bgColor = isV2 ? "bg-gradient-to-br from-green-50 to-green-100" : "bg-gradient-to-br from-blue-50 to-blue-100";

  return (
    <div
      key={animationKey}
      className={`${bgColor} rounded-lg shadow-lg p-8 animate-fade-in`}
      style={{ animation: "fadeIn 0.5s ease-in" }}
    >
      <div className="flex items-center gap-3 mb-6">
        <div
          className="w-4 h-4 rounded-full"
          style={{ backgroundColor: checkoutData.details.color }}
        />
        <h2 className="text-3xl font-bold text-gray-800">
          {checkoutData.message}
        </h2>
        <span className={`ml-auto px-4 py-2 rounded-full text-sm font-bold ${
          isV2 ? "bg-green-600 text-white" : "bg-blue-600 text-white"
        }`}>
          {checkoutData.version.toUpperCase()}
        </span>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="bg-white bg-opacity-70 rounded-lg p-4">
          <div className="text-sm text-gray-600 mb-1">CPF</div>
          <div className="font-mono font-bold text-gray-800">{checkoutData.details.cpf}</div>
        </div>

        <div className="bg-white bg-opacity-70 rounded-lg p-4">
          <div className="text-sm text-gray-600 mb-1">Bucket</div>
          <div className="font-mono font-bold text-gray-800">
            {checkoutData.details.bucket.toString().padStart(2, '0')}
          </div>
        </div>

        <div className="bg-white bg-opacity-70 rounded-lg p-4">
          <div className="text-sm text-gray-600 mb-1">UI Style</div>
          <div className="font-mono font-bold text-gray-800 capitalize">{checkoutData.details.ui}</div>
        </div>

        <div className="bg-white bg-opacity-70 rounded-lg p-4">
          <div className="text-sm text-gray-600 mb-1">Color Theme</div>
          <div className="flex items-center gap-2">
            <div
              className="w-6 h-6 rounded border-2 border-gray-300"
              style={{ backgroundColor: checkoutData.details.color }}
            />
            <span className="font-mono text-sm text-gray-800">{checkoutData.details.color}</span>
          </div>
        </div>
      </div>

      {isV2 && checkoutData.features && (
        <div className="bg-white bg-opacity-70 rounded-lg p-4">
          <div className="text-sm font-semibold text-gray-700 mb-3">✨ Exclusive V2 Features:</div>
          <ul className="space-y-2">
            {checkoutData.features.map((feature, index) => (
              <li key={index} className="flex items-center gap-2 text-gray-700">
                <span className="text-green-600">✓</span>
                <span>{feature}</span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {!isV2 && (
        <div className="bg-white bg-opacity-70 rounded-lg p-4 text-center">
          <p className="text-gray-600">Using classic checkout flow</p>
        </div>
      )}

      <div className="mt-4 text-xs text-gray-500 text-center">
        {checkoutData.timestamp}
      </div>

      <style jsx>{`
        @keyframes fadeIn {
          from {
            opacity: 0;
            transform: translateY(10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </div>
  );
}
