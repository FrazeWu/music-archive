
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRoute, RouterLink } from 'vue-router';
import { usePlayerStore } from '@/stores/player.ts';
import { useAuthStore } from '@/stores/auth';
import { useApi } from '@/composables/useApi';

const route = useRoute();
const player = usePlayerStore();
const authStore = useAuthStore();

const singerName = decodeURIComponent(route.params.artist as string).replace(/_/g, ' ');
const albumName = decodeURIComponent(route.params.album as string).replace(/_/g, ' ');

const showCorrectionModal = ref(false);
const correctionForm = ref({
  artist: '',
  album: '',
  releaseDate: '',
  reason: ''
});
const isSubmitting = ref(false);

const api = useApi();

const albumSongs = computed(() => {
  return player.songs.filter(song => song.album === albumName && song.artist === singerName);
});

const albumInfo = computed(() => {
  if (albumSongs.value.length === 0) return null;
  const firstSong = albumSongs.value[0];
  return {
    title: firstSong.album,
    artist: firstSong.artist,
    year: firstSong.year,
    releaseDate: firstSong.release_date,
    cover: firstSong.cover_url,
    trackCount: albumSongs.value.length
  };
});

const openCorrectionModal = () => {
  if (!authStore.isAuthenticated) {
    alert('请先登录后再提交修正');
    return;
  }
  if (albumInfo.value) {
    correctionForm.value = {
      artist: albumInfo.value.artist,
      album: albumInfo.value.title,
      releaseDate: albumInfo.value.releaseDate,
      reason: ''
    };
  }
  showCorrectionModal.value = true;
};

const submitCorrection = async () => {
  if (!albumSongs.value.length) {
    alert('专辑信息未找到');
    return;
  }

  if (!correctionForm.value.reason.trim()) {
    alert('请填写修正原因');
    return;
  }

  isSubmitting.value = true;

  try {
    const firstSong = albumSongs.value[0];
    const response = await fetch(`${api.url}/corrections/batch`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        song_id: firstSong.id,
        artist: correctionForm.value.artist,
        album: correctionForm.value.album,
        release_date: correctionForm.value.releaseDate,
        reason: correctionForm.value.reason
      })
    });

    if (response.ok) {
      alert('修正提交成功，等待管理员审核');
      showCorrectionModal.value = false;
    } else {
      const error = await response.json();
      alert(`提交失败: ${error.error || '未知错误'}`);
    }
  } catch (error) {
    alert('提交失败，请重试');
    console.error(error);
  } finally {
    isSubmitting.value = false;
  }
};

player.fetchSongs();
</script>

<template>
  <div class="max-w-5xl mx-auto px-8 py-20">
    <RouterLink to="/" class="inline-flex items-center gap-2 mb-8 font-bold hover:underline">
      ← 返回时间线
    </RouterLink>

    <div v-if="albumInfo" class="space-y-8">
      <div class="flex gap-8 items-start">
        <img 
          :src="albumInfo.cover || 'data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22300%22 height=%22300%22%3E%3Crect width=%22300%22 height=%22300%22 fill=%22%23000%22/%3E%3C/svg%3E'" 
          class="w-64 h-64 border-4 border-black grayscale object-cover flex-shrink-0 shadow-[15px_15px_0px_0px_rgba(0,0,0,1)]" 
          :alt="albumInfo.title" 
        />
        <div class="flex-1">
          <h1 class="text-5xl font-black tracking-tighter mb-2">{{ albumInfo.title }}</h1>
          <p class="text-xl font-bold text-gray-600 mb-1">{{ albumInfo.artist }}</p>
          <p class="text-lg text-gray-500 mb-6">{{ albumInfo.releaseDate }} · {{ albumInfo.trackCount }} {{ albumInfo.trackCount === 1 ? 'track' : 'tracks' }}</p>
          
          <div class="flex gap-4">
             <button
               @click="player.playSong(albumSongs[0])"
               class="bg-white text-black px-8 py-4 font-black text-sm uppercase tracking-widest hover:bg-black hover:text-white border-4 border-black transition-all shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] hover:shadow-none"
             >
               ▶ 播放专辑
             </button>
            <button 
              @click="openCorrectionModal"
              class="border-4 border-black px-8 py-4 font-black text-sm uppercase tracking-widest hover:bg-black hover:text-white transition-all shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] hover:shadow-none"
            >
              修正信息
            </button>
          </div>
        </div>
      </div>

      <div class="border-2 border-black bg-white">
        <div class="border-b-2 border-black bg-gray-50 px-6 py-3">
          <h2 class="font-black uppercase tracking-widest text-sm">歌曲列表</h2>
        </div>
        <div class="divide-y-2 divide-gray-100">
          <div 
            v-for="(song, index) in albumSongs" 
            :key="song.id" 
            class="flex items-center gap-6 px-6 py-4 hover:bg-gray-50 transition-colors"
          >
            <span class="text-xl font-black text-gray-400 w-12 text-right">{{ String(index + 1).padStart(2, '0') }}</span>
            <div class="flex-grow">
              <h3 class="font-bold text-lg">{{ song.title }}</h3>
            </div>
            <button 
              @click="player.playSong(song)"
              :class="['border-2 border-black px-5 py-2 font-black text-sm uppercase tracking-widest transition-all', 
                       (player.currentSong?.id === song.id && player.isPlaying) 
                         ? 'bg-black text-white' 
                         : 'hover:bg-black hover:text-white']"
            >
              {{ (player.currentSong?.id === song.id && player.isPlaying) ? '⏸ 暂停' : '▶ 播放' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showCorrectionModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4" @click.self="showCorrectionModal = false">
      <div class="bg-white border-4 border-black max-w-2xl w-full p-8 shadow-[20px_20px_0px_0px_rgba(0,0,0,1)]">
        <h2 class="text-3xl font-black mb-6 tracking-tighter">提交专辑信息修正</h2>
        
        <div class="space-y-6">
           <div>
             <label class="block text-sm font-black uppercase tracking-widest mb-2">艺术家</label>
             <input
               v-model="correctionForm.artist"
               type="text"
               class="w-full border-2 border-black p-3 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
               placeholder="艺术家名称"
             />
           </div>

           <div>
             <label class="block text-sm font-black uppercase tracking-widest mb-2">专辑名称</label>
             <input
               v-model="correctionForm.album"
               type="text"
               class="w-full border-2 border-black p-3 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
               placeholder="专辑名称"
             />
           </div>

           <div>
             <label class="block text-sm font-black uppercase tracking-widest mb-2">发行日期</label>
             <input
               v-model="correctionForm.releaseDate"
               type="date"
               class="w-full border-2 border-black p-3 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
             />
           </div>

           <div>
             <label class="block text-sm font-black uppercase tracking-widest mb-2">修正原因</label>
             <textarea
               v-model="correctionForm.reason"
               rows="3"
               class="w-full border-2 border-black p-3 focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all resize-none"
               placeholder="请详细说明修正原因"
             ></textarea>
           </div>

          <div class="flex gap-4 pt-4">
            <button 
              @click="submitCorrection"
              :disabled="isSubmitting"
              class="flex-1 bg-black text-white border-2 border-black px-6 py-3 font-black uppercase tracking-widest hover:bg-white hover:text-black transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ isSubmitting ? '提交中...' : '提交修正' }}
            </button>
            <button 
              @click="showCorrectionModal = false"
              :disabled="isSubmitting"
              class="flex-1 border-2 border-black px-6 py-3 font-black uppercase tracking-widest hover:bg-black hover:text-white transition-all disabled:opacity-50"
            >
              取消
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="text-center py-20">
      <p class="text-2xl font-black text-gray-400">专辑未找到</p>
    </div>
  </div>
</template>
