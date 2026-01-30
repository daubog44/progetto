
<script lang="ts">
import { defineComponent } from "vue";
import { useRouter } from "vue-router";
import NavBar from "@/components/template/NavBar.vue";

export default defineComponent({
  name: "HomePage",        // opzionale ma utile
  components: { NavBar },  // ⬅️ REGISTRAZIONE NECESSARIA
  setup() {
    const router = useRouter();
    const onLogoutClick = () => router.push("/logout");
    return { onLogoutClick };
  },
});
</script>

<template>
      <div class="card-body">
        <aside class="side-menu">
          <nav>
            <ul>
              <!-- usa RouterLink per evitare il full reload -->
              <li><RouterLink to="/books">LIBRI</RouterLink></li>
              <li><RouterLink to="/filmsTV">FILM E SERIE TV</RouterLink></li>
              <li><RouterLink to="/music">MUSICA</RouterLink></li>
            </ul>
          </nav>
        </aside>

       

      
      </div>
</template>
<style scoped>

/* Menu laterale bianco arrotondato */
.side-menu {
  background: white;
  border-radius: 25px;
  padding: 40px 20px;
  width: 200px;
  text-align: center;
  color: #000;
}

.side-menu ul {
  list-style: none;
  padding: 0;
}

.side-menu li {
  margin-bottom: 50px;
  font-weight: bold;
  font-size: 0.9rem;
  letter-spacing: 1px;
}

/* Layout del corpo */
.card-body {
  display: flex;
  gap: 40px;
}


</style>