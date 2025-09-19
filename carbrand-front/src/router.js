import { createRouter, createWebHistory } from "vue-router";
import UserLogin from "./views/Login.vue";
import CarsList from "./views/Cars.vue";

const routes = [
  { path: "/", redirect: "/login" },
  { path: "/login", component: UserLogin },
  { path: "/cars", component: CarsList },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to, from, next) => {
  const publicPages = ["/login"];
  const authRequired = !publicPages.includes(to.path);
  const token = localStorage.getItem("token");

  if (authRequired && !token) {
    return next("/login");
  }
  next();
});

export default router;