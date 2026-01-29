
<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useMutation } from "@pinia/colada";
import { loginUser } from "@/api/auth"; // <-- usa login
import type { LoginPayload, LoginResponse } from "@/api/types";
import { useAuthStore } from "@/stores/auth";

const router = useRouter();
const auth = useAuthStore();

// --- Form state ---
// Se il backend usa username al posto di email:
// 1) cambia sotto in: const username = ref("");
// 2) in submitForm usa { username: username.value, password: ... }
// 3) in types.ts imposta LoginPayload con username
const email = ref("");
const password = ref("");

// --- UI state ---
const isSubmitting = ref(false);
const isRedirecting = ref(false);
const errorMsg = ref<string | null>(null);

const REDIRECT_DELAY_MS = 1500;

// Mutation: <TData, TVars> = <LoginResponse, LoginPayload>
const { mutate: doLogin, data: response } = useMutation<LoginResponse, LoginPayload>({
  mutation: (vars: LoginPayload) => loginUser(vars),
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

  // Se usi username: doLogin({ username: username.value, password: password.value })
  doLogin({
    email: email.value,
    password: password.value
  });
}
</script>

<template>
  <div class="greetings">
    <h1 class="green">Accedi a Vibely</h1>
  </div>

  <form @submit.prevent="submitForm">
    <div class="label-size">
      <!-- Se usi username, cambia label e input in "Username" e v-model="username" -->
      <label class="info">Email:</label>
      <input v-model="email" type="email" required />
    </div>

    <div class="label-size">
      <label class="info">Password:</label>
      <input v-model="password" type="password" required />
    </div>

    <button type="submit" :disabled="isSubmitting || isRedirecting">
      {{ isSubmitting ? "Accesso in corso..." : "Accedi" }}
    </button>

    <p v-if="errorMsg" style="color: red">
      Errore: "Email o password errati."
    </p>
    <p v-if="response && !isRedirecting" style="color: green">
      Accesso effettuato! Reindirizzamento...
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
  color: #000000;
}
.info{ color: black;}
button:hover { background-color: rgb(124, 45, 107); }
.label-size { padding: 2%; color: #000000;}
.overlay { position: fixed; inset: 0; background: rgba(0,0,0,.6); color: #000000;
  display: grid; place-items: center; z-index: 999; }
.spinner { width: 50px; height: 50px; border: 6px solid rgba(0, 0, 0, 0.3);
  border-top-color: #000000; border-radius: 50%; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
