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

## 🛠 Stack Tecnologico
Vibely utilizza le migliori tecnologie per ogni componente:
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