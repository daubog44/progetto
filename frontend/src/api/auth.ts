
// src/api/auth.ts
import type { RegisterPayload, RegisterResponse, ApiError, LoginPayload, LoginResponse } from "./types";

const API_BASE = "http://localhost:8888";

export async function registerUser(payload: RegisterPayload): Promise<RegisterResponse> {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });

  if (!res.ok) {
    let detail: string | undefined;
    try {
      const maybeJson: ApiError | unknown = await res.json();
      detail = JSON.stringify(maybeJson);
    } catch {
      try { detail = await res.text(); } catch { detail = undefined; }
    }
    throw new Error(`Register failed: ${res.status} ${res.statusText}${detail ? " - " + detail : ""}`);
  }
  return (await res.json()) as RegisterResponse;
}

export async function loginUser(payload: LoginPayload): Promise<LoginResponse> {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    let detail: string | undefined;
    try { detail = JSON.stringify(await res.json()); } catch { try { detail = await res.text(); } catch { detail = undefined; } }
    throw new Error(`Login failed: ${res.status} ${res.statusText}${detail ? " - " + detail : ""}`);
  }
  return (await res.json()) as LoginResponse;
}
