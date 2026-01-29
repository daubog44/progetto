
<script setup lang="ts">
import { useRouter } from "vue-router";

const router = useRouter();

function onClickRedirect(string: string) {
  const route = `/${string}`;
  router.push(route); 
}

</script>
<template>
  <div class="card-header">
    <button class="welcome-text" @click="onClickRedirect('home')">VIBELY</button>

    <div class="search-bar">
      <span class="iconic-search">🔍</span>
      <input type="text" placeholder="cerca..." />
    </div>

    <div class="top-icons">
      <button class="icon-btn" type="button" @click="onClickRedirect('notifiche')">🔔</button>
      <button class="icon-btn" type="button">👤</button>
      <button class="icon-btn" type="button">🗂️</button>

      <!-- bottone logout -->
      <button id="Logout" class="icon-btn" type="button" @click="onClickRedirect('logout')">➜]</button>
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
  cursor: pointer;
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

.icon-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  margin-left: 15px;
}
</style>
