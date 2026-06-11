<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
  title: string
  subtitle?: string
  width?: string
}>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()
</script>

<template>
  <el-drawer
    :model-value="props.modelValue"
    :size="props.width ?? '560px'"
    modal-class="gp-drawer-modal"
    class="gp-drawer-panel"
    header-class="gp-drawer-header-wrap"
    body-class="gp-drawer-body-wrap"
    footer-class="gp-drawer-footer-wrap"
    @close="emit('update:modelValue', false)"
  >
    <template #header>
      <div class="gp-drawer-header">
        <div class="gp-drawer-title">{{ props.title }}</div>
        <div v-if="props.subtitle" class="gp-drawer-subtitle">{{ props.subtitle }}</div>
      </div>
    </template>

    <div class="gp-drawer-body">
      <slot />
    </div>

    <template #footer>
      <div class="gp-drawer-footer">
        <slot name="footer" />
      </div>
    </template>
  </el-drawer>
</template>
