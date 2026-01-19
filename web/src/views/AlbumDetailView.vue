<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, RouterLink } from 'vue-router';
import { usePlayerStore } from '@/stores/player.ts';
import { useAuthStore } from '@/stores/auth';

const route = useRoute();
const player = usePlayerStore();
const authStore = useAuthStore();

const singerName = decodeURIComponent(route.params.artist as string).replace(/_/g, ' ');
const albumName = decodeURIComponent(route.params.album as string).replace(/_/g, ' ');

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
    trackCount: albumSongs.value.length,
    status: firstSong.status
  };
});

// Since the correction modal is moved to EditAlbumView, this component
// no longer needs the correction-related functions or refs.

// Initial song fetching, ensuring playerStore has song data for computed properties
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
          :src="albumInfo.cover || 'data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22300%22 height=%22300%22 fill=%22%23000%22/%3E%3C/svg%3E'"
          class="w-64 h-64 border-4 border-black object-cover flex-shrink-0 shadow-[15px_15px_0px_0px_rgba(0,0,0,1)] hover:shadow-none"
          :alt="albumInfo.title" />
        <div class="flex-1">
          <h1 class="text-5xl font-black tracking-tighter mb-2">
            {{ albumInfo.title }}
            <span v-if="albumInfo.status === 'pending'"
              class="inline-block bg-yellow-400 text-yellow-900 text-base font-bold px-3 py-1 rounded-full ml-4 align-middle">
              待审核
            </span>
          </h1>
          <p class="text-xl font-bold text-gray-600 mb-1">{{ albumInfo.artist }}</p>
          <p class="text-lg text-gray-500 mb-6">{{ albumInfo.trackCount === 1 ? 'track' : 'tracks' }}</p>

          <div class="flex gap-4">
            <button @click="player.playSong(albumSongs[0])"
              class="bg-white text-black px-8 py-4 font-black text-sm uppercase tracking-widest hover:bg-black hover:text-white border-4 border-black transition-all shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] hover:shadow-none">
              ▶ 播放专辑
            </button>
            <RouterLink v-if="authStore.isAuthenticated"
              :to="`/artist=${singerName.replace(/ /g, '_')}/album=${albumName.replace(/ /g, '_')}/edit`"
              class="border-4 border-black px-8 py-4 font-black text-sm uppercase tracking-widest hover:bg-black hover:text-white transition-all shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] hover:shadow-none inline-block text-center">
              编辑专辑
            </RouterLink>
          </div>
        </div>
      </div>

      <div class="border-2 border-black bg-white">
        <div class="border-b-2 border-black bg-gray-50 px-6 py-3">
          <h2 class="font-black uppercase tracking-widest text-sm">歌曲列表</h2>
        </div>
        <div class="divide-y-2 divide-gray-100">
          <div v-for="(song, index) in albumSongs" :key="song.id"
            class="flex items-center gap-6 px-6 py-4 hover:bg-gray-50 transition-colors">
            <span class="text-xl font-black text-gray-400 w-12 text-right">{{ String(index + 1).padStart(2, '0')
              }}</span>
            <div class="flex-grow">
              <h3 class="font-bold text-lg">{{ song.title }}</h3>
            </div>
            <button @click="player.playSong(song)" :class="['border-2 border-black px-5 py-2 font-black text-sm uppercase tracking-widest transition-all',
              (player.currentSong?.id === song.id && player.isPlaying)
                ? 'bg-black text-white'
                : 'hover:bg-black hover:text-white']">
              {{ (player.currentSong?.id === song.id && player.isPlaying) ? '⏸ 暂停' : '▶ 播放' }}
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
