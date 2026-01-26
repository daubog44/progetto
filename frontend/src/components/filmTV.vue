<template>
    <div class="dashboard-filmTV">
        <div class="main-card">
            <header class="card-header">
                <h1 class="welcome-text">FILM E SERIE TV</h1>
                <Div class="search-bar">
                    <span class="iconic-search">🔍</span>
                <input type="text" placeholder="cerca..."/>
                </Div>
    <div class="top-icons"><button class="icon-btn">🔔</button>
            </div>
            </header>
             <div class="card-body">
                <aside class="side-menu">
                    <nav>
                        <ul>
                            <li>STORICO</li>
                            <li>GIALLO</li>
                            <li>THRILLER</li>
                            <li>AVVENTURA</li>
                            <li>FANTASCIENZA</li>
                            <li>FANTASY</li>
                            <li>HORROR</li>
                            <li>ROMANTICI</li>
                            <li>ANIME</li>
                            <li>DRAMA</li>
                            <li>UMORISTICO</li>
                            <li>AZIONE</li>
                            <li>NUOVI</li>
                        </ul>
                    </nav>
                </aside>
                 <aside class="right-section">
        <div class="sidebar-icons">
            <button class="icon-btn">👤</button>
            </div>
             <div class="sidebar-icons">
            <button class="icon-btn">🗂️</button>
        </div>
        <div class="community-box">
            <label> Community dove hai scritto recentemente:</label>
       </div>
       <main class="book-row">
        <h3>FILM E SERIE TV VISTE</h3>
        <div class="book-placeholder">
            <input type="button" placeholder="vedi altro"/>
        </div>
       </main>
  
    </aside>
        </div>
    </div>
    </div>
</template>
<style scoped>
.page-wrapper {
    min-height: 100vh;
    margin: 0 auto;
    padding: 20px;
  background-color: #f8efff; 
  
}


.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}
.search-bar{
    background-color: white;
    display: flex;
    align-items: center;
    padding: 8px 15px;
    border-radius: 15px;
    width: 100%;
    max-width: 300px;
}
.search-bar input{
    border: none;
    background: transparent;
    margin-left: 10px;
    width: 100%;
    outline: none;
    font-size: 0.9rem;
    color: #ccc;
}
.search-bar input::placeholder{
    color: #ccc;
}

.welcome-text {
  font-size: 2rem;
  font-weight: bold;
  color: #000;
}


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