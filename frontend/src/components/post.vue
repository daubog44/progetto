
<template>
  <div class="post-card">
    <div class="post-header">
      <div class="user-info">
        <div class="avatar-placeholder"></div>
        <span class="username">{{ data.user }}</span>
      </div>
      <button class="more-btn">•••</button>
    </div>

    <div class="post-image-container">
      <img :src="data.image" class="post-image" @dblclick="toggleLike" />
    </div>

    <div class="post-content">
      <div class="actions">
        <button @click="toggleLike" :class="{ 'liked': isLiked }">
          {{ isLiked ? '💜' : '🤍' }}
        </button>
        <button>💬</button>
        <button>✈️</button>
      </div>
      
      <p class="likes-count">Piace a <strong>{{ data.likes }} persone</strong></p>
      
      <p class="caption">
        <strong>{{ data.user }}</strong> {{ data.caption }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
const props = defineProps(['data']);
const isLiked = ref(false);

const toggleLike = () => {
  isLiked.value = !isLiked.value;
  isLiked.value ? props.data.likes++ : props.data.likes--;
};
</script>

<style scoped>
.post-card {
  background: white;
  border-radius: 20px; 
  margin-bottom: 30px;
  border: 1px solid #eee;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}

.post-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.avatar-placeholder {
  width: 35px;
  height: 35px;
  background: linear-gradient(45deg, #f09433, #e6683c, #dc2743, #cc2366, #bc1888);
  border-radius: 50%;
}

.username {
  font-weight: 600;
  font-size: 0.9rem;
}

.post-image {
  width: 100%;
  display: block;
  max-height: 600px;
  object-fit: cover;
}

.post-content {
  padding: 15px;
}

.actions {
  display: flex;
  gap: 15px;
  margin-bottom: 10px;
  font-size: 1.2rem;
}

.actions button {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.likes-count, .caption {
  font-size: 0.9rem;
  margin-bottom: 5px;
}

.liked {
  transform: scale(1.1);
}
</style>