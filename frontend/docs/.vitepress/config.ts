import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

// https://vitepress.dev/reference/site-config
export default withMermaid(defineConfig({
  srcDir: "src",
  
  title: "Progetto Microservizi",
  description: "Documentazione dettagliata del progetto a microservizi.",
  
  themeConfig: {
    search: {
      provider: 'local'
    },
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Progetto', link: '/project/introduction' }
    ],

    sidebar: {
      '/project/': [
        {
          text: 'Panoramica',
          items: [
            { text: 'Introduzione', link: '/project/introduction' },
            { text: 'Principi di Design', link: '/project/design-principles' },
            { text: 'Struttura e Monorepo', link: '/project/structure' },
            { text: 'Architettura', link: '/project/architecture' },
            { text: 'Data Flow & Operazioni', link: '/project/data-flow' },
            { text: 'Schemi Database', link: '/project/database-schemas' },
            { text: 'Database (Concetti)', link: '/project/database' },
            { text: 'Error Handling', link: '/project/error-handling' },
            { text: 'Osservabilità', link: '/project/observability' }
          ]
        }
      ]
    },

    socialLinks: [
      // { icon: 'github', link: 'https://github.com/...' } 
    ]
  },
  mermaid: {
    // mermaidConfig: {
    //   securityLevel: 'loose',
    // }
  }
}))
