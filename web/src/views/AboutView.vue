<script setup lang="ts">
// About page for Kanye West Song Archive
</script>

<template>
  <div class="max-w-4xl mx-auto px-8 py-20">
    <RouterLink to="/" class="inline-flex items-center gap-2 mb-8 font-bold hover:underline">
      ← 返回时间线
    </RouterLink>

    <div class="space-y-8">
      <h1 class="text-5xl font-black tracking-tighter border-l-8 border-black pl-6">
        关于这个收藏
      </h1>

      <div class="space-y-6 text-lg leading-relaxed">
        <p>
          这是一个关于 Kanye West 音乐生涯的互动档案项目。我们致力于收集、整理和展示 Ye 从早期作品到最新专辑的所有音乐内容。
        </p>

        <p>
          通过这个平台，你可以浏览他的音乐时间线、聆听经典作品、贡献新的发现，并帮助完善这个不断成长的音乐数据库。
        </p>

        <p>
          所有内容都经过审核，确保信息的准确性和可靠性。如果你发现了遗漏的作品或需要修正的信息，欢迎提交贡献。
        </p>

        <p>
          这个项目由热爱 Kanye West 音乐的开发者创建，希望能为粉丝提供一个全面而可靠的音乐资源。
        </p>
      </div>

      <div class="border-t-2 border-black pt-8">
        <h2 class="text-2xl font-black mb-4">贡献指南</h2>
        <ul class="space-y-2 text-gray-700">
          <li>• 注册账户后可以上传新的音乐内容</li>
          <li>• 所有上传内容将进入审核队列</li>
          <li>• 如果发现信息错误，可以提交修正请求</li>
          <li>• 保持内容的真实性和准确性</li>
        </ul>
      </div>
    </div>
  </div>
</template>