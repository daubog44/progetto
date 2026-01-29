
<!-- HomeHeader.vue (per esempio) -->
<script setup lang="ts">
import { useRouter } from "vue-router";

const router = useRouter();

function onLogoutClick() {
  router.push("/logout"); // va alla tua LogoutView che esegue clear + delay + redirect
}

document.title = "Vibely - Home";
</script>

<template>
  <div class="card-header">
    <h1 class="welcome-text">VIBELY</h1>

    <div class="search-bar">
      <span class="iconic-search">🔍</span>
      <input type="text" placeholder="cerca..." />
    </div>

    <div class="top-icons">
      <button class="icon-btn" type="button">🔔</button>
      <button class="icon-btn" type="button">👤</button>
      <button class="icon-btn" type="button">🗂️</button>

      <!-- bottone logout -->
      <button id="Logout" class="icon-btn" type="button" @click="onLogoutClick">➜]</button>
    </div>
  </div>
</template>

<style scoped>

#Logout {
  color: black;
}
.search-bar{
    background-color: white;
    display: flex;
    align-items: center;
    padding: 8px 15px;
    border-radius: 15px;
    width: 100%;
    max-width: 300px;
}
.search-bar input{
    border: none;
    background: transparent;
    width: 100%;
    outline: none;
    font-size: 0.9rem;
    color: #ccc;
}
.search-bar input::placeholder{
    color: #ccc;
}

.welcome-text {
  font-size: 2rem;
  font-weight: bold;
  color: #d121e8;
}

.sidebar-icons{
display: flex;
flex-direction: column;
align-items: flex-end;
transform: scale(1.1);
gap: 15px;
padding-right: 5px;
size: 30px;
}
</style>
