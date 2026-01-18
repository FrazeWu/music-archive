<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useAuthStore } from '@/stores/auth';
import type { PendingSong } from '@/types';
import { useApi } from '@/composables/useApi';

const authStore = useAuthStore();
const pendingSongs = ref<PendingSong[]>([]);
const loading = ref(true);

const api = useApi();

const fetchPendingSongs = async () => {
  try {
    const response = await fetch(`${api.url}/admin/pending`, {
      headers: {
        'Authorization': `Bearer ${authStore.token}`
      }
    });
    if (response.ok) {
      pendingSongs.value = await response.json();
    }
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
};

const handleApprove = async (id: number) => {
  if (!confirm('确认通过此条目？')) return;
  try {
    const response = await fetch(`${api.url}/admin/approve/${id}`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    });
    if (response.ok) {
      pendingSongs.value = pendingSongs.value.filter(s => s.id !== id);
    }
  } catch (e) {
    alert('操作失败');
  }
};

const handleReject = async (id: number) => {
  if (!confirm('确认驳回此条目？')) return;
  try {
    const response = await fetch(`${api.url}/admin/reject/${id}`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    });
    if (response.ok) {
      pendingSongs.value = pendingSongs.value.filter(s => s.id !== id);
    }
  } catch (e) {
    alert('操作失败');
  }
};

const playAudio = (url: string) => {
  const audio = new Audio(url);
  audio.play();
};

onMounted(() => {
  fetchPendingSongs();
});
</script>

<template>
  <div class="max-w-6xl mx-auto px-8 py-12">
    <h1 class="text-3xl font-black tracking-tighter mb-8 flex items-center gap-4">
      档案审核队列
      <span class="text-sm bg-black text-white px-3 py-1 rounded-full font-medium">{{ pendingSongs.length }} 待处理</span>
    </h1>

    <div v-if="loading" class="text-center py-12 text-gray-400">加载中...</div>

    <div v-else-if="pendingSongs.length === 0" class="text-center py-20 bg-gray-50 border-2 border-dashed border-gray-200">
      <p class="text-gray-400 font-medium">暂无待审核条目</p>
    </div>

    <div v-else class="space-y-4">
      <div v-for="song in pendingSongs" :key="song.id" class="bg-white border-2 border-black p-6 flex gap-6 hover:shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] transition-shadow">
        <div class="flex-1 space-y-2">
          <div class="flex items-baseline gap-4">
            <h3 class="text-xl font-black">{{ song.title }}</h3>
            <span class="text-gray-500 font-mono text-sm">{{ song.year }}</span>
          </div>
          <div class="flex gap-8 text-sm text-gray-600">
            <p><span class="font-bold">专辑:</span> {{ song.album || '未归类' }}</p>
            <p><span class="font-bold">上传者:</span> {{ song.user?.username || '未知' }}</p>
          </div>
          
          <div class="mt-4 flex gap-4">
            <button @click="playAudio(song.audio_url)" class="text-xs font-black uppercase tracking-widest bg-gray-100 hover:bg-gray-200 px-4 py-2 flex items-center gap-2">
              ▶ 试听音频
            </button>
            <a :href="song.audio_url" target="_blank" class="text-xs font-black uppercase tracking-widest text-gray-400 hover:text-black px-4 py-2">
              下载原文件
            </a>
          </div>
        </div>

        <div class="flex flex-col gap-2 border-l-2 border-gray-100 pl-6 justify-center">
          <button @click="handleApprove(song.id)" class="bg-black text-white px-6 py-2 font-bold hover:bg-green-600 transition-colors w-24">
            通过
          </button>
          <button @click="handleReject(song.id)" class="bg-white text-black border-2 border-black px-6 py-2 font-bold hover:bg-red-600 hover:text-white hover:border-red-600 transition-colors w-24">
            驳回
          </button>
        </div>
      </div>
    </div>
  </div>
</template>