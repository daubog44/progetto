0<script setup lang="ts">
defineProps<{
  msg: string
}>()
</script>

<template>
  <div class="page-background">

  <div class="greetings">
    <h1 class="purple">{{ msg }}</h1>

    </div>
  </div>
</template>

<style scoped>
  body {
    background-image: url('@/assets/sfondo.jpg');
    background-size: cover;
    height: 100vh;
  }
  
h1 {
  font-weight: 500;
  color: black;
  font-size: 2.6rem;
  position: relative;
  top: -10px;
}

h3 {
  font-size: 1.2rem;
}

.greetings h1,
.greetings h3 {
  text-align: center;
}

@media (min-width: 1024px) {
  .greetings h1,
  .greetings h3 {
    text-align: left;
  }
}
</style>
