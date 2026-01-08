import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/about',
      name: 'about',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/AboutView.vue'),
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
