<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deleteKey, listKeys } from '../api/client'
import KeyFormDrawer from '../components/KeyFormDrawer.vue'

const loading = ref(false)
const keys = ref<any[]>([])
const drawerVisible = ref(false)
const editingKey = ref<any | null>(null)

const load = async () => {
  loading.value = true
  try {
    const { data } = await listKeys()
    keys.value = data ?? []
  } catch (error) {
    ElMessage.error('加载 Keys 凭据失败')
  } finally {
    loading.value = false
  }
}

const handleCreate = () => {
  editingKey.value = null
  drawerVisible.value = true
}

const handleEdit = (row: any) => {
  editingKey.value = { ...row }
  drawerVisible.value = true
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确认删除 Key 凭据「${row.name ?? row.Name}」吗？`, '删除确认 Confirm Delete', { type: 'warning' })
    await deleteKey(row.id ?? row.ID)
    ElMessage.success('Key 凭据已删除')
    await load()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除 Key 凭据失败，可能仍被 Route 引用')
    }
  }
}

onMounted(load)
</script>

<template>
  <el-card class="gp-panel-card">
    <template #header>
      <div class="gp-action-bar">
        <div>
          <div style="font-weight: 700; color: white">Keys 凭据</div>
          <div class="gp-muted">维护 token / key / credential，并为路由注入提供来源。</div>
        </div>
        <el-button class="gp-primary-button" @click="handleCreate">新建 Key 凭据</el-button>
      </div>
    </template>
    <el-table class="gp-table" :data="keys" v-loading="loading" style="width:100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="Name 名称" />
      <el-table-column prop="description" label="Description 说明" />
      <el-table-column prop="source" label="Source 来源" width="120" />
      <el-table-column label="操作 Actions" width="240">
        <template #default="scope">
          <div class="gp-inline-actions">
            <el-button class="gp-soft-button" size="small" @click="handleEdit(scope.row)">编辑 Key</el-button>
            <el-button class="gp-soft-button" size="small" type="danger" @click="handleDelete(scope.row)">删除 Key</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <KeyFormDrawer v-model:visible="drawerVisible" :model-value="editingKey" @saved="load" />
</template>
