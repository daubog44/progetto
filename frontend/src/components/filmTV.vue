<template>
    <div class="dashboard-filmTV">
             <div class="card-body">
                <aside class="side-menu">
                    <nav>
                        <ul>
                            <li><RouterLink to="/generi">STORICO</RouterLink></li>
                            <li><RouterLink to="/generi">GIALLO</RouterLink></li>
                            <li><RouterLink to="/generi">THRILLER</RouterLink></li>
                            <li><RouterLink to="/generi">AVVENTURA</RouterLink></li>
                            <li><RouterLink to="/generi">FANTASCIENZA</RouterLink></li>
                            <li><RouterLink to="/generi">FANTASY</RouterLink></li>
                            <li><RouterLink to="/generi">HORROR</RouterLink></li>
                            <li><RouterLink to="/generi">ROMANTICI</RouterLink></li>
                            <li><RouterLink to="/generi">ANIME</RouterLink></li>
                            <li><RouterLink to="/generi">DRAMA</RouterLink></li>
                            <li><RouterLink to="/generi">UMORISTICO</RouterLink></li>
                            <li><RouterLink to="/generi">AZIONE</RouterLink></li>
                            <li><RouterLink to="/generi">NUOVI</RouterLink></li>
                        </ul>
                    </nav>
                </aside>
                 
        </div>
    </div>
</template>
<style scoped>
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

</style>