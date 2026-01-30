<template>
  <div class="gallery-container">
    <h1 class="genre-title">*genere selezionato*</h1>

    <div class="card-grid">
      <div v-for="n in 12" :key="n" class="mini-card">
        <span class="dots">...</span>
      </div>
    </div>
  </div>
</template>

<script setup>
// Qui potrai aggiungere la logica per cambiare il titolo in base alla selezione
</script>
<style acoped>
.gallery-container {
  padding: 60px 10%;
  min-height: 100vh;
  font-family: 'Serif', Georgia, serif; /* Font elegante per il titolo */
}

.genre-title {
  font-size: 2rem;
  color: #000000;
  margin-bottom: 40px;
  font-style: italic;
}

/* LA GRIGLIA: Questa impedisce alle card di diventare giganti */
.card-grid {
  display: grid;
  /* Crea colonne automatiche: minimo 120px, massimo 1 parte dello spazio */
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 25px; /* Spazio tra le card */
  max-width: 1000px; /* Limita la larghezza totale della griglia */
}

.mini-card {
  background-color: #7b7de5; /* Il viola dell'immagine */
  aspect-ratio: 3 / 4; /* Mantiene la forma rettangolare alta */
  border-radius: 15px; /* Angoli smussati */
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s;
  cursor: pointer;
}

/* Effetto al passaggio del mouse */
.mini-card:hover {
  transform: translateY(-5px);
  filter: brightness(1.1);
}

.dots {
  color: rgba(255, 255, 255, 0.6);
  font-weight: bold;
}
</style>