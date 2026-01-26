
<template>
  <Navbar />
  <Sidebar />

  <main>
    <RouterView />
  </main>
</template>
<style>
body {
  background-image: url('@/assets/sfondo.jpg');
  background-size: cover;
  height: 100vh;
}
</style>