<script setup>
import { ref, computed } from 'vue';

import { MessageSquare, Bell, CheckCheck } from 'lucide-vue-next';


const notifications = ref([
  
]);

// 2. Stato del filtro
const currentFilter = ref('all');

// 3. Logica per filtrare
const filteredNotifications = computed(() => {
  if (currentFilter.value === 'all') return notifications.value;
  return notifications.value.filter(n => n.type === currentFilter.value);
});

// 4. Funzioni
const markAllRead = () => {
  notifications.value.forEach(n => n.read = true);
};

const markAsRead = (id) => {
  const n = notifications.value.find(notif => notif.id === id);
  if (n) n.read = true;
};
</script>

<template>
  <div page-wrapper>
  <div class="container">
    <div class="notifications-card">
      <header class="header">
        <div class="title-section">
          <span class="badge">{{ notifications.filter(n => !n.read).length }}</span>
        </div>
        <button @click="markAllRead" class="btn-text">
          <CheckCheck :size="16" /> Segna tutte come lette
        </button>
      </header>

      <div class="tabs">
        <button 
          @click="currentFilter = 'all'" 
          :class="{ active: currentFilter === 'all' }">
          Tutte
        </button>
        <button 
          @click="currentFilter = 'chat'" 
          :class="{ active: currentFilter === 'chat' }">
          Chat
        </button>
        <button 
          @click="currentFilter = 'news'" 
          :class="{ active: currentFilter === 'news' }">
          Novità
        </button>
      </div>

      <div class="list">
        <div 
          v-for="n in filteredNotifications" 
          :key="n.id" 
          class="item" 
          :class="{ unread: !n.read }"
          @click="markAsRead(n.id)"
        >
          <div class="icon-wrapper">
            <MessageSquare v-if="n.type === 'chat'" :size="20" />
            <Bell v-else :size="20" />
          </div>
          <div class="content">
            <div class="content-header">
              <span class="sender">{{ n.sender }}</span>
              <span class="time">{{ n.time }}</span>
            </div>
            <p class="text">{{ n.text }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
  </div>
</template>


<style scoped>
.page-wrapper {
    min-height: 100vh;
    margin: 0 auto;
    padding: 20px;
  background-color: #f8efff; 
  
}

.notifications-page {
  max-width: 600px;
  margin: 2rem auto;
  padding: 1rem;
  font-family: 'Inter', sans-serif;
  background-color: #f9fafb;
  border-radius: 12px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #e5e7eb;
}

header h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #111827;
}

header button {
  background: none;
  border: none;
  color: #2563eb;
  font-size: 0.875rem;
  cursor: pointer;
  font-weight: 500;
}

header button:hover {
  text-decoration: underline;
}

.list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  padding: 1rem;
  background: white;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.notification-item:hover {
  background-color: #f3f4f6;
  transform: translateY(-1px);
}

.notification-item.unread {
  border-left: 4px solid #2563eb;
  background-color: #eff6ff;
}

.icon {
  font-size: 1.5rem;
  margin-right: 1rem;
  min-width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f3f4f6;
  border-radius: 50%;
}

.content {
  flex: 1;
}

.content h3 {
  font-size: 1rem;
  margin: 0 0 0.25rem 0;
  color: #1f2937;
}

.content p {
  font-size: 0.875rem;
  margin: 0 0 0.5rem 0;
  color: #6b7280;
  line-height: 1.4;
}

.time {
  font-size: 0.75rem;
  color: #9ca3af;
}

.dot {
  width: 10px;
  height: 10px;
  background-color: #2563eb;
  border-radius: 50%;
  position: absolute;
  right: 1rem;
  top: 50%;
  transform: translateY(-50%);
}
</style>