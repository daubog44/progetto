
<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useMutation } from "@pinia/colada";
import { registerUser } from "@/api/auth";
import type { RegisterPayload, RegisterResponse } from "@/api/types";
import { useAuthStore } from "@/stores/auth";

const router = useRouter();
const auth = useAuthStore();

const username = ref("");
const password = ref("");
const email = ref("");

const isSubmitting = ref(false);
const isRedirecting = ref(false);
const errorMsg = ref<string | null>(null);

const REDIRECT_DELAY_MS = 1500;

const { mutate: createUser, data: response } = useMutation<RegisterResponse, RegisterPayload>({
  mutation: (vars: RegisterPayload) => registerUser(vars),
  onSuccess: async (data) => {
    // 1) salva token (persistente)
    auth.setTokens({
      access_token: data.access_token,
      refresh_token: data.refresh_token,
      expires_in: data.expires_in,
      user_id: data.user_id
    });

    // 2) overlay e redirect
    isRedirecting.value = true;
    await new Promise((r) => setTimeout(r, REDIRECT_DELAY_MS));
    router.push("/home");
  },
  onError: (err) => {
    errorMsg.value = err instanceof Error ? err.message : String(err);
  },
  onSettled: () => {
    isSubmitting.value = false;
  }
});

function submitForm() {
  errorMsg.value = null;
  isSubmitting.value = true;

  createUser({
    email: email.value,
    password: password.value,
    username: username.value
  });
}
</script>

<template>
  <div class="greetings">
    <h1 class="green">Registrati a Vibely</h1>
  </div>

  <form @submit.prevent="submitForm">
    <div class="label-size">
      <label class="info">Username:</label>
      <input v-model="username" required />
    </div>

    <div class="label-size">
      <label class="info">Email:</label>
      <input v-model="email" type="email" required />
    </div>

    <div class="label-size">
      <label class="info">Password:</label>
      <input v-model="password" type="password" required />
    </div>

    <button type="submit" :disabled="isSubmitting || isRedirecting">
      {{ isSubmitting ? "Registrazione..." : "Registrati" }}
    </button>

    <p v-if="errorMsg" style="color: red">Errore: {{ errorMsg }}</p>
    <p v-if="response && !isRedirecting" style="color: green">
      Registrazione completata! Reindirizzamento...
    </p>
  </form>

  <div v-if="isRedirecting" class="overlay">
    <div class="spinner"></div>
    <p>Accesso in corso…</p>
  </div>
</template>

<style scoped>
button {
  margin-left: 120px;
  cursor: pointer;
  padding: 2% 5%;
  color: black;
}
button:hover { background-color: rgb(124, 45, 107); }
.info {color: black;}
.label-size { padding: 2%; color: #000000; }
.overlay { position: fixed; inset: 0; background: rgba(0,0,0,.6); color: #000000;
  display: grid; place-items: center; z-index: 999; }
.spinner { width: 50px; height: 50px; border: 6px solid rgba(0, 0, 0, 0.3);
  border-top-color: #000000; border-radius: 50%; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
