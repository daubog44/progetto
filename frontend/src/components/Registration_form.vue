
<script setup lang="ts">
import { ref } from "vue";
import { useMutation } from "@pinia/colada";
import { registerUser } from "@/api/auth";
// Stato del form
const username = ref("");
const password = ref("");
const email = ref("");

// Mutation per registrare l’utente
const {
  mutate: createUser,
  isLoading,
  data: response,
  error
} = useMutation({
  mutation: registerUser,               // funzione API
  onSuccess: (data) => {
    console.log("Registrazione completata:", data);
  }
});

// Submit del form
function submitForm() {
  createUser({
    email: email.value,
    password: password.value,
    username: username.value

  });
}
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
        <h1 class="green">Registration to Vibely</h1>
    </div>

    <div>
        <form @submit.prevent="submitForm">
            <div class="label-size">
                <label for="username">Username:</label>
                <input v-model="username" type="text" id="username" name="username" required />
            </div>

             <div class="label-size">
                <label for="email">email:</label>
                <input v-model="email" type="email" id="email" name="email" required />
            </div>

            <div class="label-size">
                <label for="password">Password:</label>
                <input v-model="password" type="password" id="password" name="password" required />
            </div>

            <div>
                <button type="submit" :disabled="isLoading">
                    {{ isLoading ? "Registering..." : "Register" }}
                </button>
            </div>

            <!-- Messaggi feedback -->
            <p v-if="error" style="color:red">Errore: {{ error.message }}</p>
            <p v-if="response" style="color:green">Registrazione completata!</p>
        </form>
    </div>

</template>
