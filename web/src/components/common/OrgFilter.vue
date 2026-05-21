<template>
  <el-select
    v-model="model"
    :placeholder="$t('common.organization', 'Организация')"
    clearable
    filterable
    :size="size"
    :style="{ width }"
    @change="onChange"
    @clear="onClear"
  >
    <el-option
      v-for="org in options"
      :key="org.id"
      :label="org.name"
      :value="org.id"
    >
      <span>{{ org.name }}</span>
      <span v-if="org.status === 'inactive'" class="org-inactive-tag">
        {{ $t('common.inactive', 'неактивна') }}
      </span>
    </el-option>
  </el-select>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue'
import { getOrganizationList } from '@/api/organization'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  width: {
    type: String,
    default: '220px'
  },
  size: {
    type: String,
    default: 'default'
  },
  // Показывать только активные организации
  activeOnly: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const options = ref([])
const loading = ref(false)

const model = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

async function loadOptions() {
  loading.value = true
  try {
    const res = await getOrganizationList({ page: 1, pageSize: 1000 })
    if (res?.code === 0 && Array.isArray(res.list)) {
      let list = res.list
      if (props.activeOnly) {
        list = list.filter((o) => o.status !== 'inactive')
      }
      options.value = list
    }
  } catch (e) {
    console.error('[OrgFilter] failed to load organizations', e)
  } finally {
    loading.value = false
  }
}

function onChange(val) {
  emit('change', val)
}

function onClear() {
  emit('update:modelValue', '')
  emit('change', '')
}

// Перезагружаем список когда меняется activeOnly
watch(() => props.activeOnly, () => loadOptions())

onMounted(() => {
  loadOptions()
})

// Публичные методы
defineExpose({
  reload: loadOptions
})
</script>

<style scoped>
.org-inactive-tag {
  margin-left: 8px;
  font-size: 12px;
  color: var(--el-color-info);
}
</style>
