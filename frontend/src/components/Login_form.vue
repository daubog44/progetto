<script setup lang="ts">

const openReadmeInEditor = () => fetch('/__open-in-editor?file=README.md')

</script>

<style>

    body {
    background-image: url('@/assets/sfondo.jpg');
    background-size: cover;
    height: 100vh;
    }

    button {
        margin-left: 120px;
        cursor: pointer;
        padding: 2% 5%;
        color: #000;
    }

    button:hover {
        background-color: rgb(0, 0, 0);

    }

    .label-size {
        padding: 2%;
        color: #000;
    }

    h1{
        color: black;
    }
</style>

<template>

    <div class="greetings">
        <h1 class="green">Accedi a Vibely</h1>
    </div>
    <div>
    
        <form action="" method="post">
            <div class="label-size">
                <label for="username">Username:</label>
                <input type="text" id="username" name="username" required>
            </div>
            <div class="label-size">
                <label for="password">Password:</label>
                <input type="password" id="password" name="password" required>
            </div>
            <div>
                <button type="submit">Accedi</button>
            </div>
        </form>
    </div>
</template>
