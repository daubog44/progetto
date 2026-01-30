
<!-- src/views/LogoutView.vue -->
<script setup lang="ts">
import { onMounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import { useRouter } from "vue-router";

const auth = useAuthStore();
const router = useRouter();
const REDIRECT_DELAY_MS = 1500;


onMounted(async () => {
  // 1) idrata eventuali token salvati
  if (!auth.accessToken) auth.hydrateFromStorage();

  // 2) esegui il logout client-side


  auth.clear();

  // 3) delay minimo (ad es. 600ms)
  await new Promise((r) => setTimeout(r, REDIRECT_DELAY_MS));
  // 4) redirect alla login
  router.replace({ path: "/login" });
});
</script>

<template>
  <section style="min-height: 50vh; display: grid; place-items: center;">
    <div style="text-align: center;">
      <p>Logout in corso…</p>
    </div>
  </section>
</template>
<style>
 p{
    color: black;
 }
</style>