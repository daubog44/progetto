
// src/api/http.ts
import { useAuthStore } from "@/stores/auth";

export async function authedFetch(input: RequestInfo, init: RequestInit = {}) {
  const auth = useAuthStore();
  const headers = new Headers(init.headers || {});
  if (auth.isAuthenticated && auth.accessToken) {
    headers.set("Authorization", `Bearer ${auth.accessToken}`);
  }
  return fetch(input, { ...init, headers });
}
