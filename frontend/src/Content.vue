<script lang="ts">
import { defineComponent } from "vue";
import { useRouter } from "vue-router";
import NavBar from "@/components/template/NavBar.vue";

export default defineComponent({
  name: "HomePage",        // opzionale ma utile
  components: { NavBar },  // ⬅️ REGISTRAZIONE NECESSARIA
  setup() {
    const router = useRouter();
    const onClickRedirect = (string: string) => {
      const route = `/${string}`;
      router.push(route); 
    };

    return { onClickRedirect };
  },
});

</script>

<template>
  <Navbar />
  <Sidebar />

  <main>
    
    <div class="page-wrapper">
      <div class="main-card">
        <NavBar />
        <RouterView />
      </div>
    </div>
  </main>
</template>
<style>
body {
  background-image: url('@/assets/sfondo.jpg');
  background-size: cover;
  height: 100vh;
}

.page-wrapper {
    height: 100vh;
    width: 150vh;
    margin: 0 auto;
    padding: 20px;
  background-color: #f8efff; 
  
}
.card-header {
  display: flex;
  justify-content: space-around;
  align-items: center;
  margin-bottom: 30px;
}
</style>