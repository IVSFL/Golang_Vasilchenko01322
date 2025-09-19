<template>
  <div class="container">
    <h2>Вход / Регистрация</h2>

    <div class="input-box">
      <input v-model="login" placeholder="Логин" class="input" />
    </div>
    <div>
      <input v-model="password" placeholder="Пароль" class="input" type="password" />
    </div>

    <div class="btn-box">
      <button @click="doLogin" class="btn">Вход</button>
      <button @click="register" class="btn">Регистрация</button>
    </div>

    <div class="err-box">
      <ul>
        <li v-for="e in errors" :key="e">{{ e }}</li>
      </ul>
    </div>
  </div>
</template>

<script>
import api from "../api";

export default {
  name: "UserLogin",
  data() {
    return {
      login: "",
      password: "",
      errors: [],
    };
  },
  methods: {
    async doLogin() {
      this.errors = [];
      if (!this.login || !this.password) {
        this.errors.push("Укажите имя пользователя и пароль");
        return;
      }
      try {
        const res = await api.post("/login", {
          login: this.login,
          password: this.password,
        });
        const token = res.data.token;
        localStorage.setItem("token", token);
        this.$router.push("/cars");
      } catch (err) {
        this.errors = ["Неверное имя или пароль"];
        console.error(err);
      }
    },
    async register() {
      this.errors = [];
      if (!this.login || !this.password) {
        this.errors.push("Укажите имя пользователя и пароль");
        return;
      }
      try {
        await api.post("/register", {
          login: this.login,
          password: this.password,
        });
        await this.doLogin();
      } catch (err) {
        this.errors = ["Ошибка регистрации (пользователь существует)"];
        console.error(err);
      }
    },
  },
};
</script>

<style>
.container {
  max-width: 400px;
  margin: 40px auto;
}

.input-box {
  margin-bottom: 10px;
}

.input {
  width: 100%;
  padding: 8px;
}

.btn-box {
  display: flex;
  gap: 10px;
}

.btn {
  padding: 5px;
  font-size: 18px;
}
</style>
