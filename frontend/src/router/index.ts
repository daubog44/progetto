import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import LoginView from '../views/LoginView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'login',
      component: LoginView,
    },
    {
      path: '/registration',
      name: 'registration',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/RegistrationView.vue'),
    },
    {
      path: '/Home',
      name: 'home',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/HomeView.vue'),
    },
  ],
})


// Guard globale: redirect /docs → VitePress dev server
router.beforeEach((to: RouteLocationNormalized) => {
  if (to.path.startsWith('/docs')) {
    // Redirect esterno (apre nuova tab o sostituisce)
    window.location.href = import.meta.env.VITE_URL_DOCS + to.path.slice(5)  // mantieni subpath
    return false  // blocca navigazione Vue
  }
})


export default router
