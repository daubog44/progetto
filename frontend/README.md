# Frontend - Progetto Microservizi

Questo è il client frontend per il progetto a microservizi, sviluppato con **Vue 3** e **Vite**.
Offre un'interfaccia moderna e reattiva per interagire con l'ecosistema di servizi backend (Auth, Social, Notifications, ecc.).

## 🛠️ Stack Tecnologico

- **Framework**: Vue 3 (Composition API)
- **Build Tool**: Vite
- **Store**: Pinia (con Colada per data fetching avanzato)
- **Routing**: Vue Router
- **Styling**: TailwindCSS 4 + Shadcn UI
- **Language**: TypeScript

## 📚 Librerie Notevoli

Il progetto fa uso di diverse librerie per garantire un'esperienza utente ricca e performante:

- **UI Components**: [Shadcn Vue](https://www.shadcn-vue.com/) + [TailwindCSS](https://tailwindcss.com/)
- **Gestione Form**: [Vee-Validate](https://vee-validate.logaretm.com/v4/)
- **Animazioni**: [Anime.js](https://animejs.com/) & `tw-animate-css`
- **Loading Spinners**: [Epic Spinners](https://epic-spinners.vuestic.dev/)
- **Notifiche**: Vue Toastification

## 🚀 Setup & Sviluppo

### Prerequisiti

- Node.js 20+
- npm

### Installazione Dipendenze

```sh
npm install
```

### Avvio Server di Sviluppo

```sh
npm run dev
```

### Build per Produzione

```sh
npm run build
```

## 📖 Documentazione

La documentazione completa del progetto (inclusi dettagli architetturali, database e data flows) è disponibile nella cartella `docs` e servita tramite **VitePress**.

Per avviare la documentazione in locale:

```sh
npm run docs:dev
```

## IDE Setup Raccomandato

[VS Code](https://code.visualstudio.com/) + [Vue (Official)](https://marketplace.visualstudio.com/items?itemName=Vue.volar).
Disabilitare Vetur se installato.
