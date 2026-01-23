
// src/api/auth.ts
import type { RegisterPayload, RegisterResponse, ApiError } from "./types";

/**
 * Base URL del gateway (porta 8888 come nel tuo ambiente dev).
 * In produzione leggi da variabile d'ambiente (es. import.meta.env.VITE_API_URL)
 */
const API_BASE = "http://localhost:8888";

/**
 * Registra un nuovo utente.
 * - Invia JSON con email/username/password
 * - Torna la risposta tipizzata con access_token, refresh_token, ecc.
 */
export async function registerUser(payload: RegisterPayload): Promise<RegisterResponse> {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // Se serve: Authorization: `Bearer ${token}`
    },
    body: JSON.stringify(payload),
    // Se usi cookie/sessions cross-site:
    // credentials: "include",
  });

  // Gestione errori: prova a leggere il JSON che spiega il problema
  if (!res.ok) {
    let detail: string | undefined;
    try {
      const maybeJson: ApiError | unknown = await res.json();
      detail = JSON.stringify(maybeJson);
    } catch {
      // se non è JSON, leggi testo
      try {
        detail = await res.text();
      } catch {
        detail = undefined;
      }
    }
    throw new Error(`Register failed: ${res.status} ${res.statusText}${detail ? " - " + detail : ""}`);
  }

  // Risposta OK tipizzata
  return (await res.json()) as RegisterResponse;
}
``
 