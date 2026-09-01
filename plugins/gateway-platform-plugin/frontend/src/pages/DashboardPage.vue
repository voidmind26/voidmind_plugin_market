<script setup lang="ts">
import { useRouter } from 'vue-router'
import DataStorageStatus from '../components/DataStorageStatus.vue'

const router = useRouter()
const openRoutes = () => router.push('/routes')
const openCreateRoute = () => router.push({ path: '/routes', query: { create: '1' } })
const openKeys = () => router.push('/keys')
const openReferences = () => router.push('/references')

const steps = [
  {
    title: '1. 新建 Key 凭据',
    desc: '先录入一个可用于注入的 token / key，例如 ips-token。',
    code: 'ips-token',
    action: openKeys,
    button: '去新建 Key',
  },
  {
    title: '2. 新建 Route 路由',
    desc: '为目标服务创建一条本地转发入口。',
    code: '/gateway/ship',
    action: openCreateRoute,
    button: '去新建 Route',
  },
  {
    title: '3. 配置 Rewrite 注入规则',
    desc: '为当前 Route 选择注入方式，并关联刚刚创建的 Key。',
    code: 'Authorization: Bearer {{value}}',
    action: openRoutes,
    button: '去编辑 Route',
  },
  {
    title: '4. 检查 References 引用关系',
    desc: '确认没有缺失引用，也没有未使用的 Key。',
    code: 'GET /api/references',
    action: openReferences,
    button: '去看 References',
  },
  {
    title: '5. 使用本地网关地址',
    desc: '最终让调用方接入本地入口，而不是直接访问远端。',
    code: 'http://127.0.0.1:18787/gateway/ship',
    action: openRoutes,
    button: '返回 Routes',
  },
]
</script>

<template>
  <div>
    <div class="gp-grid">
      <div class="gp-stat">
        <div class="gp-stat-label">Service Port 服务端口</div>
        <div class="gp-stat-value">18787</div>
      </div>
      <div class="gp-stat">
        <div class="gp-stat-label">Frontend Entry 前端入口</div>
        <div class="gp-stat-value" style="font-size: 20px">/app</div>
      </div>
      <div class="gp-stat">
        <div class="gp-stat-label">Gateway Prefix 网关前缀</div>
        <div class="gp-stat-value" style="font-size: 20px">/gateway</div>
      </div>
    </div>

    <DataStorageStatus />

    <el-row :gutter="18">
      <el-col :span="10">
        <el-card class="gp-panel-card" style="height: 100%">
          <template #header>Dashboard 总览</template>
          <el-space direction="vertical" alignment="start" size="large">
            <el-tag type="success" effect="dark">Go 重写版本已接管配置中心入口</el-tag>
            <div class="gp-muted">
              现在可以在同一个控制台中完成本地网关的 Key 凭据管理、Route 路由管理、Rewrite 注入规则绑定和引用修复。
            </div>
            <div style="display:flex;gap:10px;flex-wrap:wrap;">
              <el-button class="gp-primary-button" @click="openCreateRoute">新建 Route 路由</el-button>
              <el-button class="gp-soft-button" @click="openKeys">新建 Key 凭据</el-button>
              <el-button class="gp-soft-button" @click="openReferences">查看 References 引用关系</el-button>
            </div>
          </el-space>
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card class="gp-panel-card">
          <template #header>配置示例 Configuration Guide</template>
          <div style="display:flex;flex-direction:column;gap:16px;">
            <div v-for="item in steps" :key="item.title" style="display:grid;grid-template-columns: 1fr auto;gap:14px;padding:16px 0;border-bottom:1px solid rgba(255,255,255,0.06);">
              <div>
                <div style="font-weight:700;color:white;">{{ item.title }}</div>
                <div class="gp-muted" style="margin-top:6px;line-height:1.7;">{{ item.desc }}</div>
                <el-tag effect="plain" style="margin-top:10px;background:rgba(124,155,255,0.08);border-color:rgba(124,155,255,0.2);color:#dbe4ff;">{{ item.code }}</el-tag>
              </div>
              <div style="display:flex;align-items:center;">
                <el-button class="gp-soft-button" @click="item.action()">{{ item.button }}</el-button>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
