
// src/api/auth.ts
import type {
  RegisterPayload,
  RegisterResponse,
  ApiError,
  LoginPayload,
  LoginResponse,
} from "./types";

const API_BASE = "http://localhost:8888";

export async function registerUser(payload: RegisterPayload): Promise<RegisterResponse> {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    let detail: string | undefined;
    try {
      const maybeJson: ApiError | unknown = await res.json();
      detail = JSON.stringify(maybeJson);
    } catch {
      try { detail = await res.text(); } catch { detail = undefined; }
    }
    throw new Error(
      `Register failed: ${res.status} ${res.statusText}${detail ? " - " + detail : ""}`
    );
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
    throw new Error(
      `Login failed: ${res.status} ${res.statusText}${detail ? " - " + detail : ""}`
    );
  }
  return (await res.json()) as LoginResponse;
}

/**
 * Effettua il logout. Prova a chiamare l'endpoint server (se presente),
 * altrimenti procede comunque (logout client-side).
 *
 * @param accessToken - opzionale; se presente, viene inviato come Authorization: Bearer
 * @returns {Promise<void>} - non ritorna dati; se il server risponde con 401/404 viene ignorato.
 */
export async function logoutUser(accessToken?: string): Promise<void> {
  // Se il tuo backend ha un endpoint esplicito per logout:
  // - spesso è POST /auth/logout
  // - alcune API usano anche DELETE /auth/logout o /sessions/me
  // Adegua il path/metodo se necessario.
  const logoutUrl = `${API_BASE}/auth/logout`;

  try {
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    if (accessToken) {
      headers["Authorization"] = `Bearer ${accessToken}`;
    }

    const res = await fetch(logoutUrl, {
      method: "POST",
      headers,
      // Se usi cookie httpOnly per refresh token, abilita le credenziali:
      // credentials: "include",
    });

    // Consideriamo "ok" anche 204 No Content
    if (!res.ok && res.status !== 401 && res.status !== 404) {
      // 401/404 possono capitare se il token è già invalidato o l'endpoint non esiste.
      // In questi casi non blocchiamo il logout client-side.
      let detail: string | undefined;
      try { detail = JSON.stringify(await res.json()); } catch { try { detail = await res.text(); } catch { detail = undefined; } }
      throw new Error(
        `Logout failed: ${res.status} ${res.statusText}${detail ? " - " + detail : ""}`
      );
    }
  } catch (e) {
    // Non rilanciamo: il logout client-side deve comunque andare avanti.
    // Puoi loggare l'errore se ti serve per debug.
    // console.warn("Logout endpoint failed or unreachable:", e);
  }

  // Nessun return: lascia che lo store pulisca lo stato/token
}
