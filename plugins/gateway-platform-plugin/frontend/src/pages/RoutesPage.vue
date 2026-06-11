<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute } from 'vue-router'
import { deleteRoute, listRoutes } from '../api/client'
import RouteFormDrawer from '../components/RouteFormDrawer.vue'

const loading = ref(false)
const routes = ref<any[]>([])
const drawerVisible = ref(false)
const editingRoute = ref<any | null>(null)
const route = useRoute()

const load = async () => {
  loading.value = true
  try {
    const { data } = await listRoutes()
    routes.value = data ?? []
  } catch (error) {
    ElMessage.error('加载 Routes 路由失败')
  } finally {
    loading.value = false
  }
}

const handleCreate = (preset?: Record<string, any>) => {
  editingRoute.value = preset ? { ...preset } : null
  drawerVisible.value = true
}

const handleEdit = (row: any) => {
  editingRoute.value = { ...row }
  drawerVisible.value = true
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确认删除 Route 路由「${row.name ?? row.Name}」吗？`, '删除确认 Confirm Delete', { type: 'warning' })
    await deleteRoute(row.id ?? row.ID)
    ElMessage.success('Route 路由已删除')
    await load()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除 Route 路由失败')
    }
  }
}

const handleSaved = async (savedRoute?: any) => {
  await load()
  if (savedRoute) {
    editingRoute.value = savedRoute
  }
}

watch(
  () => route.query,
  () => {
    const routeName = route.query.route as string | undefined
    const keyName = route.query.key as string | undefined
    if (routeName) {
      const target = routes.value.find((item) => (item.name ?? item.Name) === routeName)
      if (target) {
        handleEdit(target)
        return
      }
    }
    if (keyName && !routeName) {
      handleCreate({ description: `Bind key: ${keyName}`, local_path: `/gateway-${keyName}` })
    }
  },
)

onMounted(async () => {
  await load()
  const routeName = route.query.route as string | undefined
  const keyName = route.query.key as string | undefined
  if (routeName) {
    const target = routes.value.find((item) => (item.name ?? item.Name) === routeName)
    if (target) handleEdit(target)
  } else if (keyName) {
    handleCreate({ description: `Bind key: ${keyName}`, local_path: `/gateway-${keyName}` })
  }
})
</script>

<template>
  <el-card class="gp-panel-card">
    <template #header>
      <div class="gp-action-bar">
        <div>
          <div style="font-weight: 700; color: white">Routes 路由</div>
          <div class="gp-muted">维护本地路由、上游地址、启用状态与转发入口。</div>
        </div>
        <el-button class="gp-primary-button" @click="handleCreate()">新建 Route 路由</el-button>
      </div>
    </template>
    <el-table class="gp-table" :data="routes" v-loading="loading" style="width:100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="Name 名称" />
      <el-table-column prop="local_path" label="Local Path 本地路径" />
      <el-table-column prop="upstream_url" label="Upstream URL 上游地址" min-width="260" />
      <el-table-column prop="enabled" label="Enabled 启用状态" width="110">
        <template #default="scope">
          <el-tag :type="(scope.row.enabled ?? scope.row.Enabled) ? 'success' : 'info'" effect="dark">
            {{ (scope.row.enabled ?? scope.row.Enabled) ? 'ON' : 'OFF' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作 Actions" width="240">
        <template #default="scope">
          <div class="gp-inline-actions">
            <el-button class="gp-soft-button" size="small" @click="handleEdit(scope.row)">编辑 Route</el-button>
            <el-button class="gp-soft-button" size="small" type="danger" @click="handleDelete(scope.row)">删除 Route</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <RouteFormDrawer v-model:visible="drawerVisible" :model-value="editingRoute" @saved="handleSaved" />
</template>
