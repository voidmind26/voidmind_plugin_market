<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { apiErrorMessage, createRewrite, deleteRewrite, listKeys, listRewrites, updateRewrite } from '../api/client'

const props = defineProps<{
  routeId?: number
}>()

const loading = reactive({ value: false })
const items = reactive<any[]>([])
const keyOptions = reactive<any[]>([])
const editingId = ref<number | null>(null)
const advancedVisible = ref<Record<number, boolean>>({})

const canLoad = computed(() => !!props.routeId)

type RewritePreset = {
  key: string
  label: string
  rewriteType: 'header' | 'cookie' | 'query'
  targetName: string
  template: string
}

const presetGroups: Record<'header' | 'cookie' | 'query', RewritePreset[]> = {
  header: [
    { key: 'bearer', label: 'Bearer Token', rewriteType: 'header', targetName: 'Authorization', template: 'Bearer {{value}}' },
    { key: 'raw-header', label: 'Raw Header Token', rewriteType: 'header', targetName: 'Authorization', template: '{{value}}' },
  ],
  cookie: [
    { key: 'session-cookie', label: 'Session Cookie', rewriteType: 'cookie', targetName: 'ZYBIPSCAS', template: '{{value}}' },
    { key: 'custom-cookie', label: 'Custom Cookie', rewriteType: 'cookie', targetName: 'SESSION', template: '{{value}}' },
  ],
  query: [
    { key: 'token-query', label: 'Token Query', rewriteType: 'query', targetName: 'token', template: '{{value}}' },
    { key: 'custom-query', label: 'Custom Query', rewriteType: 'query', targetName: 'query_token', template: '{{value}}' },
  ],
}

const customPresetKey = 'custom'
const currentType = (item: any) => (item.rewrite_type ?? item.RewriteType ?? 'header') as 'header' | 'cookie' | 'query'
const currentPresets = (item: any) => presetGroups[currentType(item)]

const renderPresetLabel = (item: any) => {
  if (item.__presetKey === customPresetKey) return 'Custom Rule'
  const preset = currentPresets(item).find((option) => option.key === item.__presetKey)
  return preset?.label ?? 'Custom Rule'
}

const renderSummary = (item: any) => {
  const type = currentType(item)
  const target = item.target_name ?? item.TargetName
  const keyId = item.key_id ?? item.KeyID
  const keyItem = keyOptions.find((candidate) => (candidate.id ?? candidate.ID) === keyId)
  const keyName = keyItem?.name ?? keyItem?.Name ?? '未选择 Key'
  const typeLabel = type === 'header' ? 'Header 请求头' : type === 'cookie' ? 'Cookie 注入' : 'Query 参数'
  return `${typeLabel} · ${target} · ${keyName}`
}

const hydratePreset = (item: any) => {
  const type = currentType(item)
  const target = item.target_name ?? item.TargetName
  const template = item.template ?? item.Template
  const preset = presetGroups[type].find((candidate) => candidate.targetName === target && candidate.template === template)
  item.__presetKey = preset?.key ?? customPresetKey
}

const applyPreset = (item: any) => {
  const preset = currentPresets(item).find((candidate) => candidate.key === item.__presetKey)
  if (!preset) return
  item.rewrite_type = preset.rewriteType
  item.target_name = preset.targetName
  item.template = preset.template
}

const handleTypeChange = (item: any) => {
  item.__presetKey = currentPresets(item)[0]?.key ?? customPresetKey
  applyPreset(item)
}

const handleRewriteFieldChange = (item: any) => {
  hydratePreset(item)
}

const loadKeys = async () => {
  const { data } = await listKeys()
  keyOptions.splice(0, keyOptions.length, ...(data ?? []))
}

const loadRewrites = async () => {
  if (!props.routeId) {
    items.splice(0, items.length)
    return
  }
  loading.value = true
  try {
    const { data } = await listRewrites(props.routeId)
    const next = (data ?? []).map((item: any) => ({ ...item }))
    next.forEach(hydratePreset)
    items.splice(0, items.length, ...next)
  } finally {
    loading.value = false
  }
}

const addRewrite = async () => {
  if (!props.routeId) return
  try {
    const preset = presetGroups.header[0]
    const { data } = await createRewrite(props.routeId, {
      rewrite_type: preset.rewriteType,
      target_name: preset.targetName,
      key_id: keyOptions[0]?.id ?? keyOptions[0]?.ID ?? 0,
      template: preset.template,
      ordering: items.length + 1,
    })
    const next = { ...data, __presetKey: preset.key }
    items.push(next)
    editingId.value = next.id
    ElMessage.success('注入规则已新增')
  } catch (error) {
    ElMessage.error(`新增注入规则失败：${apiErrorMessage(error, '未知错误')}`)
  }
}

const startEdit = (item: any) => {
  editingId.value = item.id
}

const cancelEdit = () => {
  editingId.value = null
}

const toggleAdvanced = (item: any) => {
  advancedVisible.value[item.id] = !advancedVisible.value[item.id]
}

const saveRewrite = async (row: any) => {
  if (!props.routeId || !row?.id) return
  try {
    const { data } = await updateRewrite(props.routeId, row.id, {
      rewrite_type: row.rewrite_type,
      target_name: row.target_name,
      key_id: row.key_id,
      template: row.template,
      ordering: row.ordering,
    })
    Object.assign(row, data)
    hydratePreset(row)
    editingId.value = null
    ElMessage.success('注入规则已更新')
  } catch (error) {
    ElMessage.error(`更新注入规则失败：${apiErrorMessage(error, '未知错误')}`)
  }
}

const removeRewrite = async (row: any) => {
  if (!props.routeId || !row?.id) return
  try {
    await deleteRewrite(props.routeId, row.id)
    const index = items.findIndex((item) => item.id === row.id)
    if (index >= 0) items.splice(index, 1)
    ElMessage.success('注入规则已删除')
  } catch (error) {
    ElMessage.error(`删除注入规则失败：${apiErrorMessage(error, '未知错误')}`)
  }
}

watch(
  () => props.routeId,
  async () => {
    await loadKeys()
    await loadRewrites()
  },
  { immediate: true },
)
</script>

<template>
  <el-card class="gp-panel-card" shadow="never" style="margin-top: 18px">
    <template #header>
      <div class="gp-action-bar">
        <div>
          <div style="font-weight: 700; color: white">Rewrites 注入规则</div>
          <div class="gp-muted">通过常见模式快速配置注入规则；复杂字段可在高级设置中展开。</div>
        </div>
        <el-button class="gp-soft-button" :disabled="!canLoad" @click="addRewrite">新增注入规则</el-button>
      </div>
    </template>

    <el-empty v-if="!canLoad" description="请先保存 Route 路由，再编辑 Rewrite 注入规则" />
    <div v-else class="rewrite-list">
      <el-empty v-if="items.length === 0" description="当前还没有任何注入规则，可先新增一条常见模式规则。" />
      <div v-for="item in items" :key="item.id" class="rewrite-card">
        <div class="rewrite-card-summary">
          <div class="rewrite-card-title">{{ renderPresetLabel(item) }}</div>
          <div class="rewrite-card-meta">{{ renderSummary(item) }}</div>
        </div>
        <div class="rewrite-card-actions">
          <el-button class="gp-soft-button" size="small" @click="startEdit(item)">编辑规则</el-button>
          <el-button class="gp-soft-button" size="small" @click="toggleAdvanced(item)">高级设置</el-button>
          <el-button class="gp-soft-button" size="small" type="danger" @click="removeRewrite(item)">删除规则</el-button>
        </div>
        <div v-if="editingId === item.id" class="rewrite-card-editor">
          <el-form label-position="top" class="rewrite-card-form">
            <el-row :gutter="12">
              <el-col :span="8">
                <el-form-item label="Type 注入位置">
                  <el-select v-model="item.rewrite_type" style="width:100%" @change="handleTypeChange(item)">
                    <el-option label="Header 请求头" value="header" />
                    <el-option label="Query 查询参数" value="query" />
                    <el-option label="Cookie 注入" value="cookie" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="Preset 常见模式">
                  <el-select v-model="item.__presetKey" style="width:100%" @change="applyPreset(item)">
                    <el-option v-if="item.__presetKey === customPresetKey" label="Custom Rule" :value="customPresetKey" />
                    <el-option v-for="preset in currentPresets(item)" :key="preset.key" :label="preset.label" :value="preset.key" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="Key 凭据">
                  <el-select v-model="item.key_id" style="width:100%">
                    <el-option v-for="option in keyOptions" :key="option.id" :label="option.name" :value="option.id" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="12">
              <el-col :span="12">
                <el-form-item label="Target name 目标字段">
                  <el-input v-model="item.target_name" @input="handleRewriteFieldChange(item)" />
                  <div class="gp-muted">例如 Authorization、ZYBIPSCAS、token。</div>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="Template 模板">
                  <el-input v-model="item.template" @input="handleRewriteFieldChange(item)" />
                  <div class="gp-muted">例如 Bearer {{value}} 或 {{value}}。</div>
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item v-if="advancedVisible[item.id]" label="Ordering 排序">
              <el-input-number v-model="item.ordering" :min="1" style="width:160px" />
            </el-form-item>
            <div class="rewrite-card-footer-actions">
              <el-button class="gp-soft-button" @click="cancelEdit">取消</el-button>
              <el-button class="gp-primary-button" @click="saveRewrite(item)">保存规则</el-button>
            </div>
          </el-form>
        </div>
      </div>
    </div>
  </el-card>
</template>
