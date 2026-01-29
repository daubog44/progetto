<template>
    <div class="dashboard-books">
  
             <div class="card-body">
                <aside class="side-menu">
                    <nav>
                      <ul>
                            <li>BIOGRAFIA E AUTOBIOGRAFIA</li>
                            <li>ROMANZO STORICO</li>
                            <li>GIALLO</li>
                            <li>THRILLER</li>
                            <li>AVVENTURA E AZIONE</li>
                            <li>FANTASCIENZA</li>
                            <li>FANTASY</li>
                            <li>HORROR</li>
                            <li>YOUNG ADULT</li>
                            <li>ROMANCE</li>
                            <li>DARK ROMANCE</li>
                            <li>UMORISTICO</li>
                            <li>FILOSOFICO</li>
                            <li>EPICI</li>
                        </ul>
                    </nav>
                </aside>
                 <aside class="right-section">
        <div class="community-box">
            <label> Community dove hai scritto recentemente:</label>
       </div>
       <main class="book-row">
        <h3>Le tue letture</h3>
        <div class="book-placeholder">
            <input type="button" placeholder="vedi altro"/>
        </div>
       </main>
  
    </aside>
        </div>
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
.community-box {
  background: white;
  padding: 15px;
  border-radius: 15px;
  margin-top: 50px;
  font-size: 0.7rem;
  color: #000;
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
  margin-bottom: 5px;
  font-weight: auto;
  font-size: 0.9rem;
  letter-spacing: 1px;
}

/* Layout del corpo */
.card-body {
  display: flex;
  gap: 40px;
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
.book-row{
   background: white;
  border-radius: 25px;
  padding: 40px 20px;
  width: 200px;
  text-align: center;
  color: #000;
}

</style>