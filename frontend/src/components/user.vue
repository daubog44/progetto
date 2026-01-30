<template>
  <div class="vibely-page">
    <div class="content-card">
      
      <section class="col-left">
        <h1 class="brand-title">COMMUNITY</h1>
        <div class="toggle-bar">
          <button class="btn active">PUBBLICO</button>
          <button class="btn">PRIVATO</button>
        </div>
      </section>

      <section class="col-right">
        <div class="header-profile">
          <div class="user-box">
            <div class="circle-avatar">FOTO</div>
            <p class="user-name">*USERNAME*</p>
          </div>
          
          <div class="stats-box">
            <div class="s-item"><b>0</b><span>POST</span></div>
            <div class="s-item"><b>0</b><span>FOLLOWER</span></div>
            <div class="s-item"><b>0</b><span>SEGUITI</span></div>
          </div>
        </div>

        <div class="data-list">
          <div class="data-item">
            <div class="info"><b>0</b><p>LIBRI LETTI</p></div>
            <button class="action">VEDI I LIBRI LETTI</button>
          </div>
          <div class="data-item">
            <div class="info"><b>00h</b><p>FILM E SERIE TV VISTE</p></div>
            <button class="action">VEDI FILM/ SERIE TV VISTE</button>
          </div>
          <div class="data-item">
            <div class="info"><b>0</b><p> BRANI SALVATI</p></div>
            <button class="action">VEDI BRANI SALVATI</button>
          </div>
        </div>
      </section>

    </div>
  </div>
</template>
<style scoped>


.content-card {

  display: flex; /* Forza le due colonne affiancate */
  padding: 60px;
  gap: 50px; /* Spazio tra colonna sinistra e destra */
  box-shadow: 0 10px 40px rgba(0,0,0,0.1);
}

/* Gestione Colonne */
.col-left {
  flex: 0 0 250px; /* Larghezza fissa per la sinistra */
   background: white;
  border-radius: 25px;
  padding: 40px 20px;
  width: 200px;
  text-align: center;
  color: #000;
}

.col-right {
  flex: 1; /* La destra prende tutto lo spazio rimanente */
   background: white;
  border-radius: 25px;
  padding: 40px 20px;
  width: 200px;
  text-align: center;
  color: #000;
}

/* Dettagli Sinistra */
.brand-title {
  font-size: 28px;
  margin-bottom: 25px;
  font-weight: bold;
}

.toggle-bar {
  padding: 5px;
  border-radius: 50px;
  display: inline-flex;
}

.btn {
  border: none;
  background: none;
  padding: 10px 20px;
  border-radius: 50px;
  cursor: pointer;
  font-weight: bold;
}

.btn.active {
  background-color: #5d5d81;
  color: white;
}

/* Dettagli Destra (Profilo) */
.header-profile {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 80px;
}

.circle-avatar {
  width: 90px;
  height: 90px;
  background-color: #8a8aab;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  margin-bottom: 10px;
}

.user-name {
  font-size: 14px;
  text-align: center;
  font-weight: bold;
  color: black;
}

.stats-box {
  display: flex;
  gap: 40px;
  text-align: center;
}

.stats-box b { font-size: 20px; display: block; }
.stats-box span { font-size: 11px; color: #555; }

/* Liste Dati */
.data-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.info { text-align: center; flex: 1; }
.info b { font-size: 22px; display: block; }
.info p { font-size: 12px; font-weight: 800; margin: 0; }

.action {
  background-color: rgba(255,255,255,0.6);
  border: none;
  width: 200px;
  padding: 12px;
  border-radius: 50px;
  font-size: 11px;
  color: #666;
  cursor: pointer;
  transition: 0.2s;
}

.action:hover { background-color: #fff; }
</style>