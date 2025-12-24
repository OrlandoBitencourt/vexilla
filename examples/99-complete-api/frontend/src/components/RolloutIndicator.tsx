"use client";

interface RolloutIndicatorProps {
  cpf: string;
  bucket: number | null;
  rollout: number;
  enabled: boolean;
}

export default function RolloutIndicator({ cpf, bucket, rollout, enabled }: RolloutIndicatorProps) {
  const isInRollout = bucket !== null && bucket < rollout;

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <h3 className="text-xl font-bold mb-4 text-gray-800">🎲 Deterministic Rollout</h3>

      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="p-4 bg-blue-50 rounded-md">
            <div className="text-sm text-gray-600 mb-1">CPF</div>
            <div className="font-mono text-lg font-bold text-blue-800">{cpf}</div>
          </div>

          <div className="p-4 bg-purple-50 rounded-md">
            <div className="text-sm text-gray-600 mb-1">Calculated Bucket</div>
            <div className="font-mono text-lg font-bold text-purple-800">
              {bucket !== null ? bucket.toString().padStart(2, '0') : '--'}
            </div>
          </div>
        </div>

        <div className="p-4 bg-indigo-50 rounded-md">
          <div className="text-sm text-gray-600 mb-1">Rollout Percentage</div>
          <div className="font-mono text-2xl font-bold text-indigo-800">{rollout}%</div>
          <div className="mt-2 w-full bg-gray-200 rounded-full h-4">
            <div
              className="bg-indigo-600 h-4 rounded-full transition-all duration-300"
              style={{ width: `${rollout}%` }}
            />
          </div>
        </div>

        <div className={`p-4 rounded-md ${
          isInRollout
            ? "bg-green-50 border-2 border-green-300"
            : "bg-red-50 border-2 border-red-300"
        }`}>
          <div className="flex items-center gap-3">
            <span className="text-3xl">
              {isInRollout ? "✅" : "❌"}
            </span>
            <div>
              <div className="font-bold text-lg text-gray-800">
                {isInRollout ? "User in Rollout" : "User NOT in Rollout"}
              </div>
              <div className="text-sm text-gray-600">
                {bucket !== null && (
                  <>
                    Bucket {bucket} is {isInRollout ? "<" : ">="} Rollout {rollout}
                  </>
                )}
              </div>
            </div>
          </div>
        </div>

        <div className="text-xs text-gray-500 bg-gray-50 p-3 rounded-md">
          <div className="font-semibold mb-1">ℹ️ How it works:</div>
          <ol className="list-decimal list-inside space-y-1">
            <li>CPF is hashed using SHA-256</li>
            <li>Hash is converted to a bucket number (0-99)</li>
            <li>If bucket &lt; rollout%, user gets new version</li>
            <li>Same CPF = same bucket = deterministic result</li>
          </ol>
        </div>
      </div>
    </div>
  );
}
