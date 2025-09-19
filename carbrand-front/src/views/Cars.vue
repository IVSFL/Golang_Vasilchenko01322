<template>
  <div class="container" style="padding:20px; max-width:900px; margin: 20px auto;">
    <h1>Автомобили</h1>

    <!-- форма добавления -->
    <div class="form" style="margin-bottom:20px;">
      <input v-model="newCar.name" placeholder="Название" />
      <input v-model="newCar.country" placeholder="Страна" />
      <input v-model="newCar.year" placeholder="Год основания" />
      <input v-model="newCar.capitalization" placeholder="Капитализация" />
      <button @click="saveCarBrand">{{ editing ? "Сохранить" : "Добавить" }}</button>

      <div v-if="errorMessages.length" style="color: red; margin-top: 10px;">
        <ul>
          <li v-for="err in errorMessages" :key="err">{{ err }}</li>
        </ul>
      </div>
    </div>

    <!-- таблица -->
    <table border="1" cellpadding="5" style="width:100%; border-collapse:collapse;">
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
import api from "../api";

export default {
  name: "CarList",
  data() {
    return {
      carBrands: [],
      newCar: { name: "", country: "", year: "", capitalization: "" },
      editing: null,
      errorMessages: [],
    };
  },
  created() {
    this.fetchCarBrands();
  },
  methods: {
    async fetchCarBrands() {
      try {
        const res = await api.get("/carBrands/");
        this.carBrands = res.data;
      } catch (err) {
        console.error(err);
      }
    },

    // validateCar принимает значения формы (строки), собирает все ошибки в массив
    validateCar(form) {
      const errors = [];
      const currentYear = new Date().getFullYear();

      const name = (form.name || "").trim();
      const country = (form.country || "").trim();
      const yearStr = String(form.year || "").trim();
      const capStr = String(form.capitalization || "").trim();

      // пустые
      if (!name) errors.push("Поле 'Название' не должно быть пустым");
      if (!country) errors.push("Поле 'Страна' не должно быть пустым");
      if (!yearStr) errors.push("Поле 'Год' не должно быть пустым");
      if (!capStr) errors.push("Поле 'Капитализация' не должно быть пустым");

      // буквы и пробелы (лат + кир)
      const lettersRegex = /^[A-Za-zА-Яа-яЁё\s]+$/;
      if (name && !lettersRegex.test(name)) errors.push("Поле 'Название' может состоять только из букв и пробелов");
      if (country && !lettersRegex.test(country)) errors.push("Поле 'Страна' может состоять только из букв и пробелов");

      // числа
      if (yearStr && (isNaN(Number(yearStr)) || Number(yearStr) > currentYear)) {
        errors.push("Поле 'Год' должно быть числом и не больше текущего года");
      }
      if (capStr && isNaN(Number(capStr))) {
        errors.push("Поле 'Капитализация' должно быть числом");
      }

      // уникальное имя (если создаём новую)
      const nameLower = name.toLowerCase();
      if (!this.editing) {
        if (this.carBrands.some(c => c.name.trim().toLowerCase() === nameLower)) {
          errors.push("Такое название уже существует");
        }
      } else {
        // при редактировании разрешено оставить текущее название, но нельзя задать название другого существующего бренда
        if (this.carBrands.some(c => c.name.trim().toLowerCase() === nameLower && c.id !== this.editing.id)) {
          errors.push("Такое название уже существует");
        }
      }

      this.errorMessages = errors;
      return errors.length === 0;
    },

    async saveCarBrand() {
      // валидируем форму (строки) — не преобразовываем пока
      if (!this.validateCar(this.newCar)) return;

      const payload = {
        name: this.newCar.name.trim(),
        country: this.newCar.country.trim(),
        year: Number(String(this.newCar.year).trim()),
        capitalization: Number(String(this.newCar.capitalization).trim()),
      };

      try {
        if (this.editing) {
          await api.put(`/carBrands/${this.editing.id}`, payload);
          this.editing = null;
        } else {
          await api.post("/carBrands/", payload);
        }
        this.newCar = { name: "", country: "", year: "", capitalization: "" };
        this.errorMessages = [];
        this.fetchCarBrands();
      } catch (err) {
        console.error(err);
        // если сервер вернул проблему unique — покажем
        if (err.response && err.response.data) {
          const msg = typeof err.response.data === "string" ? err.response.data : (err.response.data.error || JSON.stringify(err.response.data));
          this.errorMessages = [msg];
        } else {
          this.errorMessages = ["Ошибка при сохранении на сервере"];
        }
      }
    },

    editCarBrand(car) {
      // при заполнении формы преобразуем числа в строки, чтобы валидация была однородной
      this.newCar = {
        name: car.name,
        country: car.country,
        year: String(car.year),
        capitalization: String(car.capitalization),
      };
      this.editing = car;
      this.errorMessages = [];
    },

    async deleteCarBrand(id) {
      try {
        await api.delete(`/carBrands/${id}`);
        this.fetchCarBrands();
      } catch (err) {
        console.error(err);
      }
    },
  },
};
</script>

<style>
input { margin: 6px; padding: 6px; }
button { margin: 6px; padding: 6px 10px; }
</style>
