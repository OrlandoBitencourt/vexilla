"use client";

interface FlagStatusProps {
  name: string;
  value: boolean | number;
  type?: "boolean" | "number";
}

export default function FlagStatus({ name, value, type = "boolean" }: FlagStatusProps) {
  const isEnabled = type === "boolean" ? value : true;
  const displayValue = type === "number" ? `${value}%` : value ? "ENABLED" : "DISABLED";

  return (
    <div className="flex items-center justify-between p-3 bg-gray-50 rounded-md">
      <div className="flex items-center gap-3">
        <span className={`text-2xl ${isEnabled ? "text-green-500" : "text-red-500"}`}>
          {type === "boolean" ? (value ? "✔" : "✖") : "🎯"}
        </span>
        <div>
          <div className="font-mono text-sm font-semibold text-gray-800">{name}</div>
          <div className="text-xs text-gray-600">
            {type === "boolean" ? "Boolean Flag" : "Rollout Percentage"}
          </div>
        </div>
      </div>
      <div className={`px-3 py-1 rounded-full text-sm font-semibold ${
        isEnabled
          ? "bg-green-100 text-green-800"
          : "bg-red-100 text-red-800"
      }`}>
        {displayValue}
      </div>
    </div>
  );
}
