<script setup lang="ts">

</script>

<style>
    button {
        margin-left: 120px;
        cursor: pointer;
        padding: 2% 5%;
        color: black;
    }

    button:hover {
        background-color: rgb(124, 45, 107);

    }

    .label-size {
        padding: 2%;
    }
</style>

<template>

    <div class="greetings">
        <h1 class="green">Registrati a Vibely</h1>
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
            <div >
                <button type="submit">Registrati</button>
            </div>
        </form>
    </div>
</template>