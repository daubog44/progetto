<script setup lang="ts">

const openReadmeInEditor = () => fetch('/__open-in-editor?file=README.md')

</script>

<style>
    button {
        margin-left: 120px;
        cursor: pointer;
        padding: 2% 5%;
    }

    button:hover {
        background-color: lightgrey;

    }

    .label-size {
        padding: 2%;
    }
</style>

<template>

    <div class="greetings">
        <h1 class="green">Login to your account to continue to Vibely</h1>
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
                <button type="submit">Login</button>
            </div>
        </form>
    </div>
</template>