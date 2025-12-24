export interface UserContext {
  userId: string;
  cpf: string;
  role: string;
  country: string;
}

export interface FlagsSnapshot {
  flags: {
    "api.checkout.v2": boolean;
    "api.rate_limit.enabled": boolean;
    "api.kill_switch": boolean;
    "api.checkout.rollout": number;
    "frontend.new_ui": boolean;
    "frontend.beta_banner": boolean;
  };
  context: {
    user_id: string;
    cpf: string;
    role: string;
    country: string;
    bucket: number | null;
  };
  timestamp: string;
}

export interface CheckoutResponse {
  version: string;
  message: string;
  details: {
    cpf: string;
    bucket: number;
    ui: string;
    color: string;
  };
  features?: string[];
  timestamp: string;
}

export interface FlagMetrics {
  cache_hits: number;
  cache_miss: number;
  flags_loaded: number;
  circuit_breaker: string;
  last_refresh: string;
  uptime_seconds: number;
}
