"use client";

import { UserContext } from "@/types";
import { useState } from "react";

interface UserSimulatorProps {
  onContextChange: (context: UserContext) => void;
  initialContext: UserContext;
}

const MOCK_USERS = [
  { userId: "user-1", cpf: "12345678909", role: "user", country: "BR", name: "João Silva" },
  { userId: "user-2", cpf: "98765432100", role: "user", country: "BR", name: "Maria Santos" },
  { userId: "user-3", cpf: "11122233344", role: "admin", country: "BR", name: "Admin User" },
  { userId: "user-4", cpf: "55566677788", role: "user", country: "US", name: "John Doe" },
  { userId: "user-5", cpf: "99988877766", role: "user", country: "BR", name: "Ana Costa" },
];

export default function UserSimulator({ onContextChange, initialContext }: UserSimulatorProps) {
  const [context, setContext] = useState<UserContext>(initialContext);
  const [isCustom, setIsCustom] = useState(false);

  const handleSelectUser = (cpf: string) => {
    const user = MOCK_USERS.find(u => u.cpf === cpf);
    if (user) {
      const newContext = {
        userId: user.userId,
        cpf: user.cpf,
        role: user.role,
        country: user.country,
      };
      setContext(newContext);
      onContextChange(newContext);
      setIsCustom(false);
    }
  };

  const handleCustomChange = (field: keyof UserContext, value: string) => {
    const newContext = { ...context, [field]: value };
    setContext(newContext);
    onContextChange(newContext);
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <h2 className="text-2xl font-bold mb-4 text-gray-800">👤 User Simulator</h2>

      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Select Mock User
        </label>
        <select
          className="w-full p-2 border border-gray-300 rounded-md text-gray-800 bg-white"
          value={isCustom ? "custom" : context.cpf}
          onChange={(e) => {
            if (e.target.value === "custom") {
              setIsCustom(true);
            } else {
              handleSelectUser(e.target.value);
            }
          }}
        >
          {MOCK_USERS.map((user) => (
            <option key={user.cpf} value={user.cpf}>
              {user.name} - {user.cpf} ({user.role})
            </option>
          ))}
          <option value="custom">Custom User...</option>
        </select>
      </div>

      {isCustom && (
        <div className="space-y-3 mt-4 p-4 bg-gray-50 rounded-md">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              User ID
            </label>
            <input
              type="text"
              className="w-full p-2 border border-gray-300 rounded-md text-gray-800"
              value={context.userId}
              onChange={(e) => handleCustomChange("userId", e.target.value)}
              placeholder="user-1"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              CPF (11 digits)
            </label>
            <input
              type="text"
              className="w-full p-2 border border-gray-300 rounded-md text-gray-800"
              value={context.cpf}
              onChange={(e) => handleCustomChange("cpf", e.target.value)}
              placeholder="12345678909"
              maxLength={11}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Role
            </label>
            <select
              className="w-full p-2 border border-gray-300 rounded-md text-gray-800 bg-white"
              value={context.role}
              onChange={(e) => handleCustomChange("role", e.target.value)}
            >
              <option value="user">User</option>
              <option value="admin">Admin</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Country
            </label>
            <select
              className="w-full p-2 border border-gray-300 rounded-md text-gray-800 bg-white"
              value={context.country}
              onChange={(e) => handleCustomChange("country", e.target.value)}
            >
              <option value="BR">Brazil (BR)</option>
              <option value="US">United States (US)</option>
              <option value="UK">United Kingdom (UK)</option>
            </select>
          </div>
        </div>
      )}

      <div className="mt-4 p-4 bg-blue-50 rounded-md">
        <h3 className="text-sm font-semibold text-gray-700 mb-2">Current Context:</h3>
        <div className="text-xs space-y-1 font-mono text-gray-600">
          <div><span className="font-semibold">User ID:</span> {context.userId}</div>
          <div><span className="font-semibold">CPF:</span> {context.cpf}</div>
          <div><span className="font-semibold">Role:</span> {context.role}</div>
          <div><span className="font-semibold">Country:</span> {context.country}</div>
        </div>
      </div>
    </div>
  );
}
