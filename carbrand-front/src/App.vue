<template>
  <div id="app" class="container">
    <h1>Автомобили</h1>

    <!-- форма добавления -->
    <div class="form">
      <input v-model="newCar.name" placeholder="Название" />
      <input v-model="newCar.country" placeholder="Страна" />
      <input v-model.number="newCar.year" placeholder="Год основания" />
      <input v-model.number="newCar.capitalization" placeholder="Капитализация" />
      <button @click="saveCarBrand">Добавить</button>
    </div>

    <!-- таблица -->
    <table border="1" cellpadding="5">
      <thead>
        <tr>
          <th>ID</th>
          <th>Название</th>
          <th>Страна</th>
          <th>Год основания</th>
          <th>Капитализация</th>
          <th>Кнопки</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="car in carBrands" :key="car.id">
          <td>{{ car.id }}</td>
          <td>{{ car.name }}</td>
          <td>{{ car.country }}</td>
          <td>{{ car.year }}</td>
          <td>{{ car.capitalization }}</td>
          <td>
            <button @click="editCarBrand(car)">Редактировать</button>
            <button @click="deleteCarBrand(car.id)">Удалить</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "App",
  data() {
    return {
      carBrands: [],
      newCar: { name: "", country: "", year: "", capitalization: "" },
      editing: null,
    };
  },
  created() {
    this.fetchCarBrands();
  },
  methods: {
    async fetchCarBrands() {
      const res = await axios.get("http://localhost:8080/carBrands");
      this.carBrands = res.data;
    },
    async saveCarBrand() {
      if (this.editing) {
        await axios.put(`http://localhost:8080/carBrands/${this.editing.id}`, this.newCar);
        this.editing = null;
      } else {
        await axios.post("http://localhost:8080/carBrands", this.newCar);
      }
      this.newCar = { name: "", country: "", year: "", capitalization: "" };
      this.fetchCarBrands();
    },
    editCarBrand(car) {
      this.newCar = { ...car, year: car.year_founded }; // поле в API называется year
      this.editing = car;
    },
    async deleteCarBrand(id) {
      await axios.delete(`http://localhost:8080/carBrands/${id}`);
      this.fetchCarBrands();
    },
  },
};
</script>

<style>
.container {
  padding: 20px;
  font-family: Arial, sans-serif;
}
.form {
  margin-bottom: 20px;
}
input {
  margin: 5px;
}
button {
  margin: 5px;
}
</style>
