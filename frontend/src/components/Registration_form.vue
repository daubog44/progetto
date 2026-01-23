
<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useMutation } from "@pinia/colada";
import { registerUser } from "@/api/auth";
import type { RegisterPayload, RegisterResponse } from "@/api/types";

// Router
const router = useRouter();

// Form state
const username = ref("");
const password = ref("");
const email = ref("");

// UI state
const isSubmitting = ref(false);
const isRedirecting = ref(false);
const errorMsg = ref<string | null>(null);

// Delay in ms before redirect
const REDIRECT_DELAY_MS = 1500;

// Mutation (TIPI CORRETTI + arrow function per forzare l'input/output)
const {
  mutate: createUser,
  data: response,
  error, // opzionale: utile per debug
} = useMutation<RegisterResponse, RegisterPayload>({
  // 👇 Forza esplicitamente l'input (RegisterPayload) ed evita inferenze sbagliate
  mutation: (vars: RegisterPayload) => registerUser(vars),
  onSuccess: async () => {
    isRedirecting.value = true;
    await new Promise((resolve) => setTimeout(resolve, REDIRECT_DELAY_MS));
    router.push("/home");
  },
  onError: (err) => {
    errorMsg.value = err instanceof Error ? err.message : String(err);
  },
  onSettled: () => {
    isSubmitting.value = false;
  },
});

// Submit
function submitForm() {
  errorMsg.value = null;
  isSubmitting.value = true;

  createUser({
    email: email.value,
    password: password.value,
    username: username.value,
  });
}
</script>

<template>
  <div class="greetings">
    <h1 class="green">Registrati a Vibely</h1>
  </div>

  <form @submit.prevent="submitForm">
    <div class="label-size">
      <label>Username:</label>
      <input v-model="username" required />
    </div>

    <div class="label-size">
      <label>Email:</label>
      <input v-model="email" type="email" required />
    </div>

    <div class="label-size">
      <label>Password:</label>
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

  <!-- Overlay -->
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
button:hover {
  background-color: rgb(124, 45, 107);
}
.label-size {
  padding: 2%;
}

/* overlay */
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  display: grid;
  place-items: center;
  z-index: 999;
}
.spinner {
  width: 50px;
  height: 50px;
  border: 6px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
