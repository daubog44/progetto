
// src/stores/auth.ts
import { defineStore } from "pinia";

const STORAGE_KEY = "auth";

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  expiresAt: number | null; // epoch ms
  userId: string | null;
}

interface SetTokensInput {
  access_token: string;
  refresh_token: string;
  expires_in: number; // seconds
  user_id: string;
}

export const useAuthStore = defineStore("auth", {
  state: (): AuthState => ({
    accessToken: null,
    refreshToken: null,
    expiresAt: null,
    userId: null
  }),

  getters: {
    isAuthenticated(state) {
      return !!state.accessToken && !!state.expiresAt && Date.now() < state.expiresAt;
    },
    remainingSeconds(state) {
      return state.expiresAt ? Math.max(0, Math.floor((state.expiresAt - Date.now()) / 1000)) : 0;
    }
  },

  actions: {
    setTokens(resp: SetTokensInput) {
      this.accessToken = resp.access_token;
      this.refreshToken = resp.refresh_token;
      this.expiresAt = Date.now() + resp.expires_in * 1000;
      this.userId = resp.user_id;

      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          accessToken: this.accessToken,
          refreshToken: this.refreshToken,
          expiresAt: this.expiresAt,
          userId: this.userId
        })
      );
    },

    hydrateFromStorage() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (!raw) return;
        const data = JSON.parse(raw);
        this.accessToken = data.accessToken ?? null;
        this.refreshToken = data.refreshToken ?? null;
        this.expiresAt = data.expiresAt ?? null;
        this.userId = data.userId ?? null;
      } catch {
        /* ignore */
      }
    },

    clear() {
      this.accessToken = null;
      this.refreshToken = null;
      this.expiresAt = null;
      this.userId = null;
      localStorage.removeItem(STORAGE_KEY);
    }
  }
});
