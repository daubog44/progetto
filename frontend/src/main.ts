import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { PiniaColada} from '@pinia/colada'
 import { PiniaColadaDelay } from '@pinia/colada-plugin-delay'
 import { PiniaColadaRetry } from '@pinia/colada-plugin-retry'
import Toast, { type PluginOptions } from "vue-toastification";
import VueScrollProgress from "vue-scroll-progress";
// Import the CSS or use your own!
import "vue-toastification/dist/index.css";
 

import App from './App.vue'
import router from './router'
import { PiniaColadaAutoRefetch } from '@pinia/colada-plugin-auto-refetch'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app
.use
(PiniaColada
, {
  // Optionally provide global options here for queries
  queryOptions
: {
    gcTime
: 300_000, // 5 minutes, the default
  },
  plugins: [
    PiniaColadaDelay({
      delay: 0, // disabled by default
    }), PiniaColadaRetry({
      // Pinia Colada Retry options
      retry: 0,
    }),
    PiniaColadaAutoRefetch({ autoRefetch: true })
  ],  
})

// VueScrollProgress: https://github.com/spemer/vue-scroll-progress?tab=readme-ov-file#set-progress-bar-style-and-customize-as-you-wantoptional

app.use(VueScrollProgress);

//Toast: https://vue-toastification.maronato.dev/

const options: PluginOptions = {

    // You can set your default options here
};

app.use(Toast, options);


/* others notable libraries installed: 
spinners: https://epic-spinners.vuestic.dev/
animejs: https://animejs.com/
form validation: https://github.com/logaretm/vee-validate
//schadcn ui + tailwindcss
*/

app.mount('#app')
