import { createRouter, createWebHashHistory } from "vue-router";
import HomeView from "@/views/HomeView.vue";
import UploadView from "@/views/UploadView.vue";
import LoginView from "@/views/LoginView.vue";
import AdminReviewView from "@/views/AdminReviewView.vue";
import AlbumDetailView from "@/views/AlbumDetailView.vue";
import EditAlbumView from "@/views/EditAlbumView.vue";
import AboutView from "@/views/AboutView.vue";
import { useAuthStore } from "@/stores/auth";

const routes = [
  { path: "/", component: HomeView },
  { path: "/about", component: AboutView },
  { path: "/artist=:artist/album=:album", component: AlbumDetailView },
  { path: "/artist=:artist/album=:album/edit", component: EditAlbumView },
  { path: "/upload", component: UploadView },
  { path: "/login", component: LoginView },
  { path: "/register", component: LoginView },
  {
    path: "/admin/review",
    component: AdminReviewView,
    beforeEnter: (to, from, next) => {
      const authStore = useAuthStore();
      if (authStore.user?.role === "admin") {
        next();
      } else {
        next("/");
      }
    },
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition;
    } else {
      return { top: 0 };
    }
  },
});

export default router;
