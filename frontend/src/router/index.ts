
// src/router/index.ts
import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const routes = [
  { path: "/login", component: () => import("@/views/LoginView.vue") },
  { path: "/register", component: () => import("@/views/RegistrationView.vue") },
  { path: "/home", component: () => import("@/views/HomeView.vue"), meta: { requiresAuth: true } },
  { path: "/", redirect: "/home" }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach((to) => {
  const auth = useAuthStore();
  // se ricarichi direttamente /home, assicuriamoci di aver idratato
  if (!auth.accessToken) auth.hydrateFromStorage();

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { path: "/login", query: { redirect: to.fullPath } };
  }
  if ((to.path === "/login" || to.path === "/register") && auth.isAuthenticated) {
    return "/home";
  }
});

export default router;
