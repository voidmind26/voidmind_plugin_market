<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { apiErrorMessage, getHealth } from '../api/client'

interface HealthStatus {
  data_dir?: string
  database_writable?: boolean
}

const health = ref<HealthStatus | null>(null)
const loading = ref(true)
const loadError = ref('')

const dataDirectory = computed(() => {
  if (health.value?.data_dir) return health.value.data_dir
  return loading.value ? '正在读取...' : '未能读取数据目录'
})

const statusText = computed(() => {
  if (loading.value) return 'CHECKING'
  if (loadError.value) return 'UNAVAILABLE'
  return health.value?.database_writable ? 'READ / WRITE' : 'NOT WRITABLE'
})

const statusType = computed(() => {
  if (loading.value) return 'info'
  return health.value?.database_writable ? 'success' : 'danger'
})

onMounted(async () => {
  try {
    const { data } = await getHealth()
    health.value = data
  } catch (error) {
    loadError.value = apiErrorMessage(error, '健康检查请求失败')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="gp-stat gp-storage-status">
    <div class="gp-storage-details">
      <div class="gp-stat-label">Data Directory 数据目录</div>
      <div class="gp-stat-value gp-storage-path">
        {{ dataDirectory }}
      </div>
      <div v-if="loadError" class="gp-muted gp-storage-error">
        {{ loadError }}
      </div>
    </div>
    <el-tag :type="statusType" effect="dark">
      {{ statusText }}
    </el-tag>
  </div>
</template>

<style scoped>
.gp-storage-status {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 22px;
}

.gp-storage-details {
  min-width: 0;
}

.gp-storage-path {
  font-size: 15px;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.gp-storage-error {
  margin-top: 8px;
  font-size: 12px;
}
</style>
