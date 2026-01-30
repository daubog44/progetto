<template>
    <div class="dashboard-music">
             <div class="card-body">
                <aside class="side-menu">
                    <nav>
                        <ul>
                            <li><RouterLink to="/generi">POP</RouterLink></li>
                            <li><RouterLink to="/generi">JAZZ</RouterLink></li>
                            <li><RouterLink to="/generi">HIP HOP</RouterLink></li>
                            <li><RouterLink to="/generi">ELETTRONICA</RouterLink></li>
                            <li><RouterLink to="/generi">FUNK</RouterLink></li>
                            <li><RouterLink to="/generi">COUNTRY</RouterLink></li>
                            <li><RouterLink to="/generi">ROCK</RouterLink></li>
                            <li><RouterLink to="/generi">CLASSICA</RouterLink></li>
                            <li><RouterLink to="/generi">RnB</RouterLink></li>
                            <li><RouterLink to="/generi">SOUL</RouterLink></li>
                            <li><RouterLink to="/generi">BEATS</RouterLink></li>
                            <li><RouterLink to="/generi">ACOUSTIC</RouterLink></li>
                            <li><RouterLink to="/generi">EDM</RouterLink></li>
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