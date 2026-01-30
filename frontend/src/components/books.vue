<template>
    <div class="dashboard-books">
  
             <div class="card-body">
                <aside class="side-menu">
                    <nav>
                      <ul>
                            <li><RouterLink to="/generi">BIOGRAFIA E AUTOBIOGRAFIA</RouterLink></li>
                            <li><RouterLink to="/generi">ROMANZO STORICO</RouterLink></li>
                            <li><RouterLink to="/generi">GIALLO</RouterLink></li>
                            <li><RouterLink to="/generi">THRILLER</RouterLink></li>
                            <li><RouterLink to="/generi">AVVENTURA E AZIONE</RouterLink></li>
                            <li><RouterLink to="/generi">FANTASCIENZA</RouterLink></li>
                            <li><RouterLink to="/generi">FANTASY</RouterLink></li>
                            <li><RouterLink to="/generi">HORROR</RouterLink></li>
                            <li><RouterLink to="/generi">YOUNG ADULT</RouterLink></li>
                            <li><RouterLink to="/generi">ROMANCE</RouterLink></li>
                            <li><RouterLink to="/generi">DARK ROMANCE</RouterLink></li>
                            <li><RouterLink to="/generi">UMORISTICO</RouterLink></li>
                            <li><RouterLink to="/generi">FILOSOFICO</RouterLink></li>
                            <li><RouterLink to="/generi">EPICI</RouterLink></li>
                        </ul>
                    </nav>
                </aside>
                 <aside class="right-section">
       
  
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
.sidebar-icons{
display: flex;
flex-direction: column;
align-items: flex-end;
transform: scale(1.1);
gap: 15px;
padding-right: 5px;
size: 30px;
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