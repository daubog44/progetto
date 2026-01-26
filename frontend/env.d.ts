/// <reference types="vite/client" />

declare module 'vue-scroll-progress';

declare module '*.vue' {
  import { DefineComponent } from 'vue';
  const component: DefineComponent<{}, {}, any>;
  export default component;
}
