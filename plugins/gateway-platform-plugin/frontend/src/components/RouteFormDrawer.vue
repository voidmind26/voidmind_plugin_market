<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { apiErrorMessage, createRoute, updateRoute } from '../api/client'
import DrawerShell from './DrawerShell.vue'
import RewriteTableEditor from './RewriteTableEditor.vue'

const props = defineProps<{
  visible: boolean
  modelValue: any | null
}>()
const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved', route?: any): void
}>()

const form = reactive({
  name: '',
  enabled: true,
  local_path: '',
  upstream_url: '',
  timeout_ms: 30000,
  description: '',
})
const saving = ref(false)

const isEdit = computed(() => !!(props.modelValue?.id ?? props.modelValue?.ID))
const routeId = computed(() => props.modelValue?.id ?? props.modelValue?.ID)

watch(
  () => props.modelValue,
  (value) => {
    form.name = value?.name ?? value?.Name ?? ''
    form.enabled = value?.enabled ?? value?.Enabled ?? true
    form.local_path = value?.local_path ?? value?.LocalPath ?? ''
    form.upstream_url = value?.upstream_url ?? value?.UpstreamURL ?? ''
    form.timeout_ms = value?.timeout_ms ?? value?.TimeoutMS ?? 30000
    form.description = value?.description ?? value?.Description ?? ''
  },
  { immediate: true },
)

const close = () => emit('update:visible', false)

const submit = async () => {
  if (saving.value) return

  const wasEdit = isEdit.value
  saving.value = true
  try {
    let route
    if (wasEdit) {
      const { data } = await updateRoute(routeId.value, form)
      route = data
      ElMessage.success('Route 路由已更新')
    } else {
      const { data } = await createRoute(form)
      route = data
      ElMessage.success('Route 路由已创建')
    }
    emit('saved', route)
    if (wasEdit) close()
  } catch (error) {
    ElMessage.error(`保存 Route 路由失败：${apiErrorMessage(error, '未知错误')}`)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <DrawerShell
    :model-value="visible"
    :title="isEdit ? '编辑 Route 路由' : '新建 Route 路由'"
    subtitle="维护本地路径、上游地址、启用状态与基础运行参数。"
    width="760px"
    @update:model-value="emit('update:visible', $event)"
  >
    <el-form label-position="top">
      <el-row :gutter="14">
        <el-col :span="12">
          <el-form-item label="Name 名称">
            <el-input v-model="form.name" />
            <div class="gp-muted">建议用业务语义命名，例如 ship、docs、apifox。</div>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="Local Path 本地路径">
            <el-input v-model="form.local_path" />
            <div class="gp-muted">本地网关入口路径，例如 /gateway/ship。</div>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="Upstream URL 上游地址">
        <el-input v-model="form.upstream_url" />
        <div class="gp-muted">填写真实目标服务地址，通常是完整的 http(s) URL。</div>
      </el-form-item>
      <el-row :gutter="14">
        <el-col :span="12">
          <el-form-item label="Timeout (ms) 超时">
            <el-input-number v-model="form.timeout_ms" :min="1" style="width:100%" />
            <div class="gp-muted">建议保留默认值；只有上游确实较慢时再调大。</div>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="Enabled 启用状态">
            <el-switch v-model="form.enabled" />
            <div class="gp-muted">关闭后该路由将拒绝转发请求。</div>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="Description 说明">
        <el-input v-model="form.description" type="textarea" :rows="4" />
        <div class="gp-muted">写清楚这条路由服务于哪个系统或场景。</div>
      </el-form-item>
    </el-form>

    <el-alert type="info" :closable="false" style="margin-bottom: 16px">
      <template #title>
        先选注入位置，再选常见模式；只有自定义场景才需要展开高级设置。
      </template>
    </el-alert>

    <RewriteTableEditor :route-id="routeId" />

    <template #footer>
      <div style="display:flex;justify-content:space-between;align-items:center;width:100%">
        <div class="gp-muted">保存 Route 后会刷新列表；Rewrite 可在当前弹出页继续维护。</div>
        <div style="display:flex;gap:10px">
          <el-button class="gp-soft-button" @click="close">取消</el-button>
          <el-button class="gp-primary-button" :loading="saving" @click="submit">保存 Route</el-button>
        </div>
      </div>
    </template>
  </DrawerShell>
</template>
