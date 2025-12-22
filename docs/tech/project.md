# 🌐 Vibely: Piattaforma Culturale

Vibely è un ecosistema digitale progettato per connettere persone attraverso la passione per **Libri, Film, Serie TV e Musica**. Non è un semplice social network, ma un "hub culurale" dove l'utente può catalogare i propri progressi, scoprire nuovi talenti e discutere in modo sicuro.

## 🎯 Visione
In un mondo di informazioni frammentate, Vibely aggrega contenuti multimediali in un'unica esperienza fluida, supportata da un'architettura a microservizi pensata per la scalabilità e le prestazioni estreme.

## ✨ Caratteristiche Chiave

### 📚 Catalogazione Multimediale
- **Archivio Personale**: Traccia cosa hai letto, visto o ascoltato.
- **Whitelist/Blacklist**: Filtra i contenuti in base ai tuoi gusti culturali.

### 💬 Community & Social
- **Chat Spoiler-Safe**: Sistema di protezione avanzato che nasconde i messaggi se non hai ancora completato l'opera.
- **Relazioni di Valore**: Segui utenti con gusti simili ai tuoi grazie all'analisi del grafo di interesse (Neo4j).

### 🚀 Spazio Emergenti
- **Discovery**: Una sezione dedicata esclusivamente ad autori, registi e musicisti indipendenti per dare visibilità ai nuovi talenti.

## 🛠 Stack Tecnologico & Architettura
Vibely utilizza le migliori tecnologie per ogni componente. Per approfondimenti sui pilastri tecnici, consulta il documento su [Sicurezza, Performance e Resilienza](security-performance-resilience.md).

- **Backend**: Go (Golang) con gRPC e Kafka.
- **Databases**:
  - **Postgres**: Identity e Master of Record.
  - **MongoDB**: Metadati opere e Post flessibili.
  - **Neo4j**: Grafi di interesse e Relazioni.
  - **Cassandra**: Feed, Timeline e Messaggistica massiva.
  - **Redis**: Real-time Pub/Sub e Caching.

## 📈 Roadmap
1. **MVP**: Lancio dei cataloghi base e community di genere.
2. **Phase 2**: Implementazione chat spoiler-safe e algoritmo di raccomandazione.
3. **Phase 3**: Integrazione diretta con API di streaming e store digitali.

# 🎯 Vision & Goals

Vibely è un ecosistema digitale d'avanguardia progettato per connettere le persone attraverso la passione per **Libri, Film, Serie TV e Musica**. Non è un semplice social network, ma un "hub culturale" dove l'utente può catalogare i propri progressi, scoprire nuovi talenti e discutere in modo sicuro.

## 🚀 Analisi di Sistema

### Obiettivi Funzionali
- **Connessione Culturale**: Permettere agli utenti di connettersi tramite post, recensioni e feed personalizzati basati su interessi comuni.
- **Gestione Profilo & Social**: Creazione profili, sistema di follow e messaggistica in tempo reale.
- **User Experience Fluida**: Pubblicazione di contenuti ottimizzata per percepire operazioni istantanee.
- **Ricerca Avanzata**: Strumenti di discovery rapidi per opere, utenti e post.
- **Sicurezza & Privacy**: Gestione dell'eliminazione degli account e moderazione (AI + umana) per contenuti offensivi o spoiler.

### Obiettivi Non Funzionali
- **Alta Disponibilità**: Funzionamento continuo anche durante aggiornamenti o guasti parziali.
- **Scalabilità Elastica**: Capacità di adattarsi al traffico in tempo reale.
- **Integrità & Performance**: Protezione dei dati e latenze minime tramite caching aggressivo.
- **Osservabilità**: Monitoraggio costante per interventi proattivi.

---

## 👥 Stakeholder

- **Utenti Finali**: Appassionati che cercano condivisione e socializzazione.
- **Team Tecnico**: Sviluppatori, Ops e Security specialist che mantengono la piattaforma.
- **Content Managers**: Moderatori (AI e umani) che garantiscono un ambiente sano.
- **Product Owners**: Chi definisce la direzione strategica e le priorità.
- **External Providers**: Fornitori di infrastruttura cloud e servizi di terze parti.
