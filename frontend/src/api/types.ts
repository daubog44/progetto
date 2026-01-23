
// src/api/types.ts

/**
 * Payload per la registrazione utente /auth/register
 * - Tutti i campi sono required secondo la tua documentazione.
 */
export interface RegisterPayload {
  email: string;
  password: string;
  username: string;
}

/**
 * Risposta dell'endpoint /auth/register
 */
export interface RegisterResponse {
  $schema: string;       // URL del JSON schema restituito dal backend
  access_token: string;  // JWT o token d'accesso
  refresh_token: string; // token di refresh
  expires_in: number;    // secondi (int64 in Go -> number in TS)
  user_id: string;       // id dell'utente creato
}

/**
 * Struttura d'errore opzionale che potresti ricevere dal backend
 * (utile per il debug/UX lato client).
 */
export interface ApiError {
  error?: string;
  message?: string;
  // alcuni backend usano questa forma:
  errors?: Array<{ field?: string; message?: string }>;
}
