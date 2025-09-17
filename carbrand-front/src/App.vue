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
    <div v-if="errorMessage.length">
      <ul>
        <li v-for="err in errorMessage" :key="err">{{ err }}</li>
      </ul>
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
//import axios from "axios";
import api from "./api.js"

export default {
  name: "App",
  data() {
    return {
      carBrands: [],
      newCar: { name: "", country: "", year: "", capitalization: "" },
      editing: null,
      errorMessage: [],
      count: 0
    };
  },
  created() {
    this.fetchCarBrands();
  },
  methods: {
    validateCar(car) {
      const errors = [];
      // Пустые поля
      if(!car.name) errors.push("Поле 'Название' не должно быть пустым");
      if(!car.country) errors.push("Поле 'Страна' не должно быть пустым");
      if(!car.year) errors.push("Поле 'Год' не должно быть пустым");
      if(!car.capitalization) errors.push("Поле 'Капитализация' не должно быть пустым")

      // Проверка чисел
      const currentDate = new Date();
      const currentYear = currentDate.getFullYear();
      if(car.year && isNaN(Number(car.year)) || Number(car.year) > currentYear) errors.push("Поле 'Год' может состоять только из цифр, и не быть больше текущего года");
      if(car.capitalization && isNaN(Number(car.capitalization))) errors.push("Поле 'Капитализация' может состоять только из цифр");

      // Уникальное название
      if(!this.editing && this.carBrands.some(c => c.name === car.name)) {
        errors.push("Такое название уже существует")
      }

      // Проверка букв и пробелов
      const regex = /^[a-zA-Z\s]+$/;
      if(car.name && !regex.test(car.name)) errors.push("Поле 'Название' может состоять только из букв и пробелов");
      if(car.country && !regex.test(car.country)) errors.push("Поле 'Страна' может состоять только из букв и пробелов")

      this.errorMessage = errors;

      return errors.length === 0;
    },

    async fetchCarBrands() {
      const res = await api.get("/carBrands");
      this.carBrands = res.data;
    },
    async saveCarBrand() {
      const payload = {
        ...this.newCar,
        year: Number(this.newCar.year),
        capitalization: Number(this.newCar.capitalization),
      };
      
      // Валидация
      if(!this.validateCar(payload)) return;

      try {
        if(this.editing) {
          // Редакутирование
          await api.put(`/carBrands/${this.editing.id}`, payload);
          this.editing = null;
        } else {
          // Добавление
          await api.post("/carBrands", payload);
        }

        // Очистка форм
        this.errorMessage = [];
        this.newCar = { name: "", country: "", year: "", capitalization: "" };
        this.fetchCarBrands();
      } catch(err) {
        this.errorMessage = ["Ошибка при сохранении"];
        console.log(err);
      }
    },
    editCarBrand(car) {
      this.newCar = { ...car,}; 
      this.editing = car;
    },
    async deleteCarBrand(id) {
      await api.delete(`/carBrands/${id}`);
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
