<template>
  <div>
    <nav class="nav">
      <router-link to="/cars" class="nav-car">Автомобили</router-link>
      <router-link to="/login">Вход</router-link>
      <button v-if="token" @click="logout" class="logout">Выйти</button>
    </nav>
    <router-view />
  </div>
</template>

<script>
export default {
  computed: {
    token() {
      return localStorage.getItem("token");
    },
  },
  methods: {
    logout() {
      localStorage.removeItem("token");
      this.$router.push("/login");
    },
  },
};
</script>

<style>
.nav{
  padding: 10px;
  background-color: #f2f2f2;
}
.nav-car{
  margin-right: 10px;
}
.logout{
  float: right;
}
</style>