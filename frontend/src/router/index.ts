
// src/router/index.ts
import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const routes = [

  
 // Redirect dinamico all'avvio
    {
      path: "/",
      redirect: () => {
        const auth = useAuthStore();
        if (!auth.accessToken) auth.hydrateFromStorage(); // se usi localStorage
        return auth.isAuthenticated ? "/home" : "/login";
      },
    },

  // Layout pubblico (per login e register)
  {
    path: "/",
    component: () => import("@/PublicLayout.vue"),
    children: [
      { path: "login", component: () => import("@/views/LoginView.vue") },
      { path: "register", component: () => import("@/views/RegistrationView.vue") },
      { path: "logout", component: () => import("@/views/LogoutView.vue") },
    ],
  },

  // Layout autenticato (per il resto dell'app)
  {
    path: "/",
    component: () => import("@/Content.vue"),  // <— percorso corretto
    meta: { requiresAuth: true },
    children: [
      { path: "home", component: () => import("@/views/HomeView.vue") },
      { path: "Onebook", component: () => import("@/views/OnebookView.vue") },
      { path: "books", component: () => import("@/views/booksView.vue") },
      { path: "filmsTV", component: () => import("@/views/filmTVView.vue") },

      // opzionale: default child per quando vai su "/"
      { path: "", redirect: "/home" },
    ],
  },

  // 404 → manda a /home (o /login se preferisci)
  { path: "/:pathMatch(.*)*", redirect: "/login" },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const auth = useAuthStore();

  if (!auth.accessToken) auth.hydrateFromStorage();

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { path: "/login", query: { redirect: to.fullPath } };
  }

  if ((to.path === "/login" || to.path === "/register") && auth.isAuthenticated) {
    return "/home";
  }
});

export default router;
