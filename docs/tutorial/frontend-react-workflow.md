# React Workflow: Gestione Eventi Asincroni e WebSocket

In questa guida implementeremo un sistema robusto in React per gestire il ciclo di vita dei WebSocket e reagire agli eventi asincroni del backend (Kafka).

---

## 🏗️ 1. Il Custom Hook: `useVibelyEvents`

Invece di gestire i socket nei singoli componenti, creiamo un hook centralizzato che gestisce **connessione, sottoscrizione e cleanup**.

```tsx
import { useEffect, useState, useRef } from 'react';

export function useVibelyEvents(userId: string | null) {
  const [isReady, setIsReady] = useState(false);
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    // 1. Quando inizializzare: Solo se l'utente è loggato
    if (!userId) return;

    const socketUrl = `ws://api.vibely.com/ws/events?userId=${userId}`;
    socketRef.current = new WebSocket(socketUrl);

    socketRef.current.onopen = () => {
      console.log('Connected to Vibely Events');
      setIsReady(true);
    };

    socketRef.current.onclose = () => {
      console.log('Disconnected');
      setIsReady(false);
    };

    // 2. Quando farli morire: Cleanup automatico all'unmount o logout
    return () => {
      if (socketRef.current) {
        socketRef.current.close();
      }
    };
  }, [userId]); // Re-inizializza se userId cambia

  return { isReady, socket: socketRef.current };
}
```

---

## 🛠️ 2. Integrazione nel Componente (Esempio Registrazione)

Ecco come gestire il flusso: **Chiamata gRPC -> Attesa Asincrona -> Notifica**.

```tsx
import React, { useState, useEffect } from 'react';
import { useVibelyEvents } from './hooks/useVibelyEvents';

export const RegistrationWorkflow = () => {
  const [status, setStatus] = useState<'idle' | 'sync_done' | 'completed' | 'error'>('idle');
  const [userId, setUserId] = useState<string | null>(null);
  
  // Inizializziamo l'ascolto quando abbiamo l'ID utente (dopo la fase sincrona)
  const { isReady, socket } = useVibelyEvents(userId);

  useEffect(() => {
    if (!socket || !isReady) return;

    // Ascoltiamo i messaggi dal Notification Service
    const handleMessage = (event: MessageEvent) => {
      const data = JSON.parse(event.data);
      
      if (data.type === 'PROFILE_READY') {
        setStatus('completed');
        // Redirect alla dashboard dopo un breve delay
      }
      
      if (data.type === 'SETUP_ERROR') {
        setStatus('error');
      }
    };

    socket.addEventListener('message', handleMessage);
    
    // Cleanup specifico del listener
    return () => socket.removeEventListener('message', handleMessage);
  }, [socket, isReady]);

  const onRegister = async (data: any) => {
    try {
      // FASE 1: Sincrona (gRPC)
      const res = await api.auth.createUser(data);
      setUserId(res.id); // Questo triggera l'apertura del WebSocket nel hook
      setStatus('sync_done');
    } catch (e) {
      setStatus('error');
    }
  };

  return (
    <div>
      {status === 'idle' && <RegisterForm onSubmit={onRegister} />}
      
      {status === 'sync_done' && (
        <div className="loader">
          <p>Account creato! Stiamo preparando il tuo profilo culturale...</p>
          <p>Stato Connessione: {isReady ? 'In attesa di eventi...' : 'Connessione in corso...'}</p>
        </div>
      )}

      {status === 'completed' && <p>✅ Profilo pronto! Benvenuto su Vibely.</p>}
      
      {status === 'error' && <p>❌ Ops, qualcosa è andato storto. Riprova tra poco.</p>}
    </div>
  );
};
```

---

## 📌 Regole d'Oro per il Frontend React

### 1. Quando Inizializzare?
- **Login/Registration**: Il socket deve nascere quando l'identità dell'utente è nota. Se usi un sistema come Auth0 o Clerk, inizializzalo nel `Layout` principale o nel `ProtectedRouter`.
- **Context API**: Se vuoi che il socket sia accessibile ovunque (es. per notifiche in tempo reale in ogni pagina), avvolgi l'app in un `EventProvider`.

### 2. Quando farli morire?
- **Logout**: Chiudi esplicitamente la connessione per evitare "ghost connections" sul server.
- **Unmount critico**: Se il socket serve solo per una specifica pagina (es. una chat), usa il cleanup di `useEffect` per chiuderlo quando l'utente cambia pagina.

### 3. Gestione Fallimenti (Resilienza)
- **Reconection Logic**: Se il socket cade per problemi di rete, usa una libreria come `reconnecting-websocket` o implementa un intervallo di retry esponenziale.
- **State Management**: Usa uno store (Zustand, Redux o simple Context) per memorizzare gli eventi asincroni se l'utente cambia pagina mentre l'evento è in transito.
