<template>
  <div class="main-container">
    <div class="card-wrapper">
      
      <aside class="sidebar">
        <h2>COMMUNITY</h2>
        <div class="toggle-container">
          <button class="btn-toggle active">PUBBLICO</button>
          <button class="btn-toggle">PRIVATO</button>
        </div>
      </aside>

      <main class="profile-section">
        <div class="profile-header">
          <div class="avatar-container">
            <div class="avatar-circle">FOTO</div>
            <span class="username">*USERNAME*</span>
          </div>
          
          <div class="stats-container">
            <div class="stat-item">
              <span class="stat-value">0</span>
              <span class="stat-label">POST</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">0</span>
              <span class="stat-label">FOLLOWER</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">0</span>
              <span class="stat-label">SEGUITI</span>
            </div>
          </div>
        </div>
      </main>

        <main class="stats-section">
          <div class="stat-item">
            <h2>0</h2>
            <label>LIBRI LETTI</label>
          </div>
          <div class="stat-item">
            <h2>00h</h2>
            <label>FILM E SERIE TV VISTE</label>
          </div>
          <div class="stat-item">
            <h2>0</h2>
            <label>PLAYLIST SALVATE</label>
          </div>
        </main>

    </div>
  </div>
</template>
<style scoped>

/* Pannello Sinistro */
.left-panel {
  flex: 1;
}

.title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #000000;
}

.toggle-group {
  display: inline-flex;
  background: rgba(0, 0, 0, 0.4);
  padding: 5px;
  border-radius: 30px;
}

.btn-toggle {
  border: none;
  padding: 8px 20px;
  border-radius: 25px;
  cursor: pointer;
  background: transparent;
  font-weight: bold;
}

.btn-toggle.active {
  background-color: #595959;
  color: white;
}

/* Pannello Destro */
.right-panel {
  flex: 2;
  display: flex;
  flex-direction: column;
}

.profile-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 60px;
}

.avatar {
  width: 80px;
  height: 80px;
  background-color: #9292e7;
  border-radius: 50%;
  display: flex;
  justify-content: center;
  align-items: center;
  color: white;
  margin-bottom: 5px;
}

.username {
  font-size: 12px;
  font-weight: bold;
  display: block;
  text-align: center;
}

.stats {
  display: flex;
  gap: 30px;
  text-align: center;
}

.num {
  display: block;
  font-size: 18px;
  font-weight: bold;
}

.stat span:not(.num) {
  font-size: 10px;
  color: #000000;
}
/* Statistiche */
.stats-section {
  flex-grow: 1;
}

.stat-item {
  margin-bottom: 30px;
}

.stat-item h2 {
  font-size: 2.5rem;
  margin: 0;
  color: #000;
}

.stat-item label {
  font-size: 0.7rem;
  font-weight: bold;
  color: #000000;
}


.pill-button {
  background-color: rgba(255,255,255,0.5);
  border: none;
  padding: 10px;
  width: 180px;
  border-radius: 20px;
  color: #666;
  font-size: 10px;
  cursor: pointer;
}
</style>