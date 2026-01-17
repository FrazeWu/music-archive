
<script setup lang="ts">
import AppTopbar from './components/AppTopbar.vue';
import AudioPlayer from './components/AudioPlayer.vue';
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <AppTopbar />
    <main class="flex-grow pb-32">
      <router-view></router-view>
    </main>
    <AudioPlayer />
  </div>
</template>
