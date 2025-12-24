import { UserContext, FlagsSnapshot, CheckoutResponse, FlagMetrics } from "@/types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

class APIClient {
  private createHeaders(context: UserContext): HeadersInit {
    return {
      "Content-Type": "application/json",
      "X-User-ID": context.userId,
      "X-CPF": context.cpf,
      "X-User-Role": context.role,
      "X-Country": context.country,
    };
  }

  async getFlagsSnapshot(context: UserContext): Promise<FlagsSnapshot> {
    const response = await fetch(`${API_BASE_URL}/flags/snapshot`, {
      headers: this.createHeaders(context),
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch flags snapshot: ${response.statusText}`);
    }

    return response.json();
  }

  async checkout(context: UserContext): Promise<CheckoutResponse> {
    const response = await fetch(`${API_BASE_URL}/checkout`, {
      method: "POST",
      headers: this.createHeaders(context),
    });

    if (!response.ok) {
      throw new Error(`Checkout failed: ${response.statusText}`);
    }

    return response.json();
  }

  async invalidateAllFlags(): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/admin/flags/invalidate-all`, {
      method: "POST",
    });

    if (!response.ok) {
      throw new Error(`Failed to invalidate flags: ${response.statusText}`);
    }
  }

  async invalidateFlag(flagKey: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/admin/flags/${flagKey}`, {
      method: "POST",
    });

    if (!response.ok) {
      throw new Error(`Failed to invalidate flag ${flagKey}: ${response.statusText}`);
    }
  }

  async getMetrics(): Promise<FlagMetrics> {
    const response = await fetch(`${API_BASE_URL}/admin/flags/metrics`);

    if (!response.ok) {
      throw new Error(`Failed to fetch metrics: ${response.statusText}`);
    }

    return response.json();
  }

  async healthCheck(): Promise<{ status: string; service: string; time: string }> {
    const response = await fetch(`${API_BASE_URL}/health`);

    if (!response.ok) {
      throw new Error(`Health check failed: ${response.statusText}`);
    }

    return response.json();
  }
}

export const apiClient = new APIClient();
