<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { listReferences } from '../api/client'

const loading = ref(false)
const missing = ref<any[]>([])
const unused = ref<any[]>([])
const router = useRouter()

const load = async () => {
  loading.value = true
  try {
    const { data } = await listReferences()
    missing.value = data.missing_references ?? []
    unused.value = data.unused_keys ?? []
  } catch (error) {
    ElMessage.error('加载 References 引用关系失败')
  } finally {
    loading.value = false
  }
}

const goFixRoute = (row: any) => {
  router.push({ path: '/routes', query: { route: row.route, key: row.key } })
}

const goBindRoute = (row: any) => {
  router.push({ path: '/routes', query: { key: row.key } })
}

onMounted(load)
</script>

<template>
  <el-row :gutter="18" v-loading="loading">
    <el-col :span="12">
      <el-card class="gp-panel-card">
        <template #header>缺失引用 Missing References</template>
        <el-empty v-if="missing.length === 0" description="当前没有缺失引用，配置关系看起来是完整的。" />
        <el-table v-else class="gp-table" :data="missing">
          <el-table-column prop="route" label="Route 路由" />
          <el-table-column prop="key" label="Key 凭据" />
          <el-table-column prop="type" label="Type 注入位置" />
          <el-table-column prop="target_name" label="Target 目标字段" />
          <el-table-column label="修复 Actions" width="160">
            <template #default="scope">
              <el-button class="gp-soft-button" size="small" @click="goFixRoute(scope.row)">去 Route 修复</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-col>
    <el-col :span="12">
      <el-card class="gp-panel-card">
        <template #header>未使用 Keys</template>
        <el-empty v-if="unused.length === 0" description="当前没有未使用 Key，可继续维护已有绑定。" />
        <el-table v-else class="gp-table" :data="unused">
          <el-table-column prop="key" label="Key 凭据" />
          <el-table-column prop="description" label="Description 说明" />
          <el-table-column label="修复 Actions" width="160">
            <template #default="scope">
              <el-button class="gp-soft-button" size="small" @click="goBindRoute(scope.row)">去绑定 Route</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-col>
  </el-row>
</template>
