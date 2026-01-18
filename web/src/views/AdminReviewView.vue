<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useApi } from '@/composables/useApi';

interface Song {
  id: number;
  title: string;
  year: number;
  audio_url: string;
  album?: string;
  track_number?: number;
}

interface PendingRequest {
  id: string; // Batch ID or Group Key
  batch_id: string;
  album: string;
  artist: string;
  user: {
    username: string;
  };
  created_at: string;
  count: number;
  songs: Song[];
  song_ids: number[];
}

const authStore = useAuthStore();
const pendingRequests = ref<PendingRequest[]>([]);
const loading = ref(true);

const api = useApi();

const fetchPendingRequests = async () => {
  try {
    const response = await fetch(`${api.url}/admin/pending`, {
      headers: {
        'Authorization': `Bearer ${authStore.token}`
      }
    });
    if (response.ok) {
      pendingRequests.value = await response.json();
    }
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
};

const handleApproveBatch = async (req: PendingRequest) => {
  if (!confirm(`确认通过 "${req.album}" 的 ${req.count} 首歌曲？`)) return;
  try {
    const response = await fetch(`${api.url}/admin/approve-batch`, {
      method: 'POST',
      headers: { 
        'Authorization': `Bearer ${authStore.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ ids: req.song_ids })
    });
    if (response.ok) {
      pendingRequests.value = pendingRequests.value.filter(r => r.id !== req.id);
    }
  } catch (e) {
    alert('操作失败');
  }
};

const handleRejectBatch = async (req: PendingRequest) => {
  if (!confirm('确认驳回此批次？')) return;
  try {
    const response = await fetch(`${api.url}/admin/reject-batch`, {
      method: 'POST',
      headers: { 
        'Authorization': `Bearer ${authStore.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ ids: req.song_ids })
    });
    if (response.ok) {
      pendingRequests.value = pendingRequests.value.filter(r => r.id !== req.id);
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
  fetchPendingRequests();
});
</script>

<template>
  <div class="max-w-6xl mx-auto px-8 py-12">
    <h1 class="text-3xl font-black tracking-tighter mb-8 flex items-center gap-4">
      档案审核队列
      <span class="text-sm bg-black text-white px-3 py-1 rounded-full font-medium">{{ pendingRequests.length }} 待处理请求</span>
    </h1>

    <div v-if="loading" class="text-center py-12 text-gray-400">加载中...</div>

    <div v-else-if="pendingRequests.length === 0" class="text-center py-20 bg-gray-50 border-2 border-dashed border-gray-200">
      <p class="text-gray-400 font-medium">暂无待审核请求</p>
    </div>

    <div v-else class="space-y-8">
      <div v-for="req in pendingRequests" :key="req.id" class="bg-white border-2 border-black p-8 hover:shadow-[15px_15px_0px_0px_rgba(0,0,0,1)] transition-shadow">
        
        <!-- Header -->
        <div class="flex justify-between items-start mb-6 border-b-2 border-gray-100 pb-4">
          <div>
            <h2 class="text-2xl font-black mb-1">{{ req.album || 'Unknown Album' }}</h2>
            <p class="text-gray-500 font-bold">{{ req.artist }}</p>
            <div class="text-sm text-gray-400 mt-2 flex gap-4">
              <span>上传者: {{ req.user?.username }}</span>
              <span>{{ new Date(req.created_at).toLocaleString() }}</span>
              <span class="bg-gray-100 text-black px-2 rounded font-bold">{{ req.count }} tracks</span>
            </div>
          </div>
          
          <div class="flex gap-4">
             <button @click="handleApproveBatch(req)" class="bg-black text-white px-6 py-3 font-bold hover:bg-green-600 transition-colors">
               全部通过
             </button>
             <button @click="handleRejectBatch(req)" class="bg-white text-black border-2 border-black px-6 py-3 font-bold hover:bg-red-600 hover:text-white hover:border-red-600 transition-colors">
               全部驳回
             </button>
          </div>
        </div>

        <!-- Song List -->
        <div class="space-y-2">
           <div v-for="song in req.songs" :key="song.id" class="flex items-center gap-4 p-3 bg-gray-50 hover:bg-gray-100">
             <span class="text-gray-400 font-mono w-8 text-right">{{ song.track_number || '#' }}</span>
             <div class="flex-1 font-bold">{{ song.title }}</div>
             <button @click="playAudio(song.audio_url)" class="text-xs font-black uppercase tracking-widest bg-white border border-black px-3 py-1 hover:bg-black hover:text-white">
               ▶ 试听
             </button>
             <a :href="song.audio_url" target="_blank" class="text-gray-400 hover:text-black text-sm">⬇</a>
           </div>
        </div>

      </div>
    </div>
  </div>
</template>