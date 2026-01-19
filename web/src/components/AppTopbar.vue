<script setup lang="ts">
import { RouterLink, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { ref } from 'vue';

const authStore = useAuthStore();
const router = useRouter();
const isDropdownOpen = ref(false);

const handleLogout = () => {
  authStore.logout();
  router.push('/login');
  isDropdownOpen.value = false;
};
</script>

<template>
  <header class="sticky top-0 z-50 bg-white border-b-2 border-black h-16 flex items-center px-8 shadow-sm">
    <div class="max-w-6xl w-full mx-auto flex justify-between items-center">
      <RouterLink to="/" class="text-2xl font-black tracking-tighter hover:opacity-70 transition-opacity">
        KANYE ARCHIVE
      </RouterLink>
      <nav class="flex gap-8 font-medium items-center">
        <RouterLink to="/" class="hover:underline">时间线</RouterLink>
        <RouterLink v-if="authStore.isAuthenticated" to="/upload" class="hover:underline">上传</RouterLink>

        <RouterLink v-if="authStore.user?.role === 'admin'" to="/admin/review"
          class="text-red-600 hover:text-red-800 font-bold hover:underline">
          审核队列
        </RouterLink>

        <div v-if="authStore.isAuthenticated" class="relative" @mouseenter="isDropdownOpen = true"
          @mouseleave="isDropdownOpen = false">
          <button class="font-black hover:underline flex items-center gap-1 py-2">
            {{ authStore.user?.username }}
            <span class="text-xs transition-transform duration-200" :class="{ 'rotate-180': isDropdownOpen }">▼</span>
          </button>

          <div v-show="isDropdownOpen"
            class="absolute right-0 top-full w-32 animate-in fade-in slide-in-from-top-1 duration-200">
            <div class="bg-white border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
              <button @click="handleLogout"
                class="w-full text-left px-4 py-3 hover:bg-black hover:text-white font-bold transition-colors">
                登出
              </button>

            </div>

          </div>
        </div>

         <RouterLink v-else to="/login" class="hover:underline">登录</RouterLink>
         <RouterLink to="/about" class="hover:underline">关于</RouterLink>

      </nav>
    </div>
  </header>
</template>
