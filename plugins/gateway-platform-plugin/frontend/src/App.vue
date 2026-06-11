<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const active = computed(() => route.path)
const handleSelect = (index: string) => router.push(index)
const currentTitle = computed(() => {
  if (route.path.startsWith('/routes')) return 'Routes 管理'
  if (route.path.startsWith('/keys')) return 'Keys 管理'
  if (route.path.startsWith('/references')) return '引用关系'
  return '本地网关配置中心'
})
</script>

<template>
  <el-container class="gp-shell">
    <el-aside class="gp-sidebar" width="260px">
      <div class="gp-brand">
        <div class="gp-brand-kicker">Gateway • Local Control</div>
        <div class="gp-brand-title">Gateway Platform</div>
        <div class="gp-brand-subtitle">
          统一管理本地网关路由、Key 注入和引用关系，保持配置、转发与调试都在同一个控制台里完成。
        </div>
      </div>
      <el-menu :default-active="active" @select="handleSelect">
        <el-menu-item index="/">Dashboard</el-menu-item>
        <el-menu-item index="/routes">Routes</el-menu-item>
        <el-menu-item index="/keys">Keys</el-menu-item>
        <el-menu-item index="/references">References</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container class="gp-main">
      <el-header class="gp-header">
        <div>
          <div class="gp-header-title">{{ currentTitle }}</div>
          <div class="gp-header-meta">Go + Gin + Vue 3 + Element Plus + SQLite</div>
        </div>
        <el-tag type="success" effect="dark">Rewrite Active</el-tag>
      </el-header>
      <el-main class="gp-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>
