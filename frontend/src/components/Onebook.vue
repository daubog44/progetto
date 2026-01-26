<template>
  <div class="book-detail-page">
    <div class="main-content">
      <div class="left-panel">
        <div class="book-cover">LIBRO</div>
        
        <div class="status-options">
          <label>
            <input type="radio" value="lista" v-model="status" /> LISTA LETTURA
          </label>
          <label>
            <input type="radio" value="completato" v-model="status" /> COMPLETATO
          </label>
        </div>
      </div>

      <div class="right-panel">
        <div class="tabs">
          <button @click="activeTab = 'trama'" :class="{ active: activeTab === 'trama' }">trama</button>
          <button @click="activeTab = 'recensioni'" :class="{ active: activeTab === 'recensioni' }">recensioni</button>
        </div>

        <div class="info-box">
          <p v-if="activeTab === 'trama'">Il testo della trama va qui...</p>
          <p v-else>Le recensioni degli utenti vanno qui...</p>
        </div>

        <div class="chat-actions">
          <button class="chat-btn">Entra nella chat</button>
          <button class="chat-btn">Entra nella chat privata</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
const currentTab = ref('trama');
const status = ref('lista');
</script>
<style scoped>

/* Card centrale */
.main-card {
  background: white;
  width: 100%;
  max-width: 800px;
  border-radius: 30px;
  padding: 40px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.05);
}

/* Top Bar */
.top-bar {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 30px;
}

.search-input {
  width: 100%;
  background: #f3f3f3;
  border: none;
  border-radius: 20px;
  padding: 10px 20px;
}

/* Griglia contenuto */
.content-grid {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 40px;
}

/* Copertina Libro */
.book-cover {
  background-color: #7b52ff;
  aspect-ratio: 2/3;
  border-radius: 15px;
  display: flex;
  justify-content: center;
  align-items: center;
  color: white;
  font-weight: bold;
  margin-bottom: 15px;
}

/* Tabs */
.tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.tabs button {
  border: 1px solid #ddd;
  background: white;
  border-radius: 20px;
  padding: 5px 20px;
  cursor: pointer;
  transition: 0.3s;
}

.tabs button.active {
  background: #7b52ff;
  color: white;
  border-color: #7b52ff;
}

/* Box info bianco */
.info-box {
  background: #fdfdfd;
  border: 1px solid #eee;
  border-radius: 20px;
  height: 150px;
  padding: 20px;
}

/* Footer Chat */
.chat-actions {
  display: flex;
  justify-content: space-around;
  margin-top: 40px;
}

.chat-btn {
  background: none;
  border: none;
  font-weight: 800;
  font-size: 1.1rem;
  cursor: pointer;
  color: #333;
}

.chat-btn:hover {
  color: #7b52ff;
}
</style>