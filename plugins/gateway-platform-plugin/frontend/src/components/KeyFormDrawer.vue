<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { createKey, updateKey } from '../api/client'
import DrawerShell from './DrawerShell.vue'

const props = defineProps<{
  visible: boolean
  modelValue: any | null
}>()
const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved'): void
}>()

const form = reactive({
  name: '',
  value: '',
  description: '',
  source: 'manual',
})

const isEdit = computed(() => !!props.modelValue?.id)

watch(
  () => props.modelValue,
  (value) => {
    form.name = value?.name ?? ''
    form.value = value?.value ?? ''
    form.description = value?.description ?? ''
    form.source = value?.source ?? 'manual'
  },
  { immediate: true },
)

const close = () => emit('update:visible', false)

const submit = async () => {
  try {
    if (isEdit.value) {
      await updateKey(props.modelValue.id, form)
      ElMessage.success('Key 凭据已更新')
    } else {
      await createKey(form)
      ElMessage.success('Key 凭据已创建')
    }
    emit('saved')
    close()
  } catch (error) {
    ElMessage.error('保存 Key 凭据失败')
  }
}
</script>

<template>
  <DrawerShell
    :model-value="visible"
    :title="isEdit ? '编辑 Key 凭据' : '新建 Key 凭据'"
    subtitle="维护 token / key / credential 的值与来源信息。"
    width="520px"
    @update:model-value="emit('update:visible', $event)"
  >
    <el-form label-position="top">
      <el-form-item label="Name 名称">
        <el-input v-model="form.name" />
        <div class="gp-muted">建议使用易识别的业务名，例如 ips-token、apifox-token。</div>
      </el-form-item>
      <el-form-item label="Value 凭据值">
        <el-input v-model="form.value" type="textarea" :rows="4" />
        <div class="gp-muted">填写真实 token / key / cookie 值；保存后会作为注入来源。</div>
      </el-form-item>
      <el-form-item label="Description 说明">
        <el-input v-model="form.description" />
        <div class="gp-muted">写清楚这个凭据给谁使用。</div>
      </el-form-item>
      <el-form-item label="Source 来源">
        <el-input v-model="form.source" />
        <div class="gp-muted">可填写 manual、local、imported 等来源标记。</div>
      </el-form-item>
    </el-form>

    <template #footer>
      <div style="display:flex;justify-content:space-between;align-items:center;width:100%">
        <div class="gp-muted">敏感值修改后会立即写入本地配置源。</div>
        <div style="display:flex;gap:10px">
          <el-button class="gp-soft-button" @click="close">取消</el-button>
          <el-button class="gp-primary-button" @click="submit">保存 Key</el-button>
        </div>
      </div>
    </template>
  </DrawerShell>
</template>
