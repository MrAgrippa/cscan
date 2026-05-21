<template>
  <div class="organization-page">
    <el-card class="page-header" shadow="never">
      <div class="header-content">
        <div>
          <h2 class="page-title">{{ $t('organization.title', 'Организации') }}</h2>
          <p class="page-desc">
            {{ $t('organization.desc', 'Управление организациями и их таргетами (IP, CIDR, домены). К таргетам автоматически привязываются ассеты при сканировании.') }}
          </p>
        </div>
        <div class="header-actions">
          <el-button type="primary" :icon="Plus" @click="openOrgDialog()">
            {{ $t('organization.newOrganization', 'Создать организацию') }}
          </el-button>
          <el-button :icon="Refresh" @click="loadOrgs">
            {{ $t('common.refresh', 'Обновить') }}
          </el-button>
        </div>
      </div>
    </el-card>

    <div class="page-body">
      <!-- Левая колонка: список организаций -->
      <el-card class="org-list-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>{{ $t('organization.list', 'Список организаций') }}</span>
            <el-input
              v-model="orgSearch"
              :placeholder="$t('common.search', 'Поиск')"
              :prefix-icon="Search"
              size="small"
              clearable
              style="width: 180px"
            />
          </div>
        </template>

        <div v-loading="orgLoading" class="org-list">
          <div
            v-for="org in filteredOrgs"
            :key="org.id"
            class="org-item"
            :class="{ active: selectedOrg?.id === org.id }"
            @click="selectOrg(org)"
          >
            <div class="org-item-main">
              <div class="org-name">
                <el-tag
                  :type="org.status === 'disable' ? 'info' : 'success'"
                  size="small"
                  effect="plain"
                  style="margin-right: 8px"
                >
                  {{ org.status === 'disable' ? $t('common.disabled', 'выкл') : $t('common.enabled', 'вкл') }}
                </el-tag>
                <span>{{ org.name }}</span>
              </div>
              <div class="org-desc" v-if="org.description">{{ org.description }}</div>
            </div>
            <el-dropdown trigger="click" @click.stop>
              <el-button text :icon="More" size="small" @click.stop />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="openOrgDialog(org)">
                    {{ $t('common.edit', 'Редактировать') }}
                  </el-dropdown-item>
                  <el-dropdown-item @click="toggleOrgStatus(org)">
                    {{ org.status === 'disable' ? $t('common.enable', 'Включить') : $t('common.disable', 'Выключить') }}
                  </el-dropdown-item>
                  <el-dropdown-item divided @click="confirmDeleteOrg(org)">
                    <span style="color: var(--el-color-danger)">
                      {{ $t('common.delete', 'Удалить') }}
                    </span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>

          <el-empty
            v-if="!orgLoading && filteredOrgs.length === 0"
            :description="$t('organization.empty', 'Нет организаций')"
          />
        </div>
      </el-card>

      <!-- Правая колонка: детали и таргеты -->
      <el-card class="org-detail-card" shadow="never">
        <template v-if="!selectedOrg">
          <el-empty :description="$t('organization.selectHint', 'Выберите организацию слева для управления таргетами')" />
        </template>
        <template v-else>
          <template #header>
            <div class="card-header">
              <div>
                <span class="detail-title">{{ selectedOrg.name }}</span>
                <el-tag size="small" style="margin-left: 8px">
                  {{ targetTotal }} {{ $t('organization.targets', 'таргетов') }}
                </el-tag>
              </div>
              <div>
                <el-button :icon="Refresh" size="small" @click="reloadAfterRetag" :loading="retagLoading">
                  {{ $t('organization.retag', 'Пересобрать привязку') }}
                </el-button>
              </div>
            </div>
          </template>

          <el-tabs v-model="activeTab">
            <el-tab-pane :label="$t('organization.tabTargets', 'Таргеты')" name="targets">
              <div class="target-toolbar">
                <el-input
                  v-model="targetSearch"
                  :placeholder="$t('organization.searchTarget', 'Поиск по значению')"
                  :prefix-icon="Search"
                  size="default"
                  clearable
                  style="width: 240px"
                  @input="debouncedLoadTargets"
                />
                <el-select v-model="targetTypeFilter" clearable :placeholder="$t('common.type', 'Тип')" size="default" style="width: 140px" @change="loadTargets">
                  <el-option label="IP" value="ip" />
                  <el-option label="CIDR" value="cidr" />
                  <el-option :label="$t('organization.typeDomain', 'Домен')" value="domain" />
                  <el-option :label="$t('organization.typeWildcard', 'Wildcard')" value="wildcard" />
                </el-select>
                <div style="flex: 1" />
                <el-button type="primary" :icon="Plus" @click="openTargetDialog()">
                  {{ $t('organization.addTarget', 'Добавить таргет') }}
                </el-button>
                <el-button :icon="Upload" @click="importDialogVisible = true">
                  {{ $t('organization.importTargets', 'Импорт') }}
                </el-button>
                <el-button v-if="selectedTargetIds.length > 0" type="danger" :icon="Delete" @click="confirmDeleteTargets()">
                  {{ $t('common.delete', 'Удалить') }} ({{ selectedTargetIds.length }})
                </el-button>
              </div>

              <el-table
                v-loading="targetLoading"
                :data="targets"
                row-key="id"
                @selection-change="onTargetSelect"
                stripe
              >
                <el-table-column type="selection" width="48" />
                <el-table-column :label="$t('common.type', 'Тип')" width="120">
                  <template #default="{ row }">
                    <el-tag size="small" :type="targetTypeColor(row.type)">{{ targetTypeLabel(row.type) }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="value" :label="$t('organization.targetValue', 'Значение')" min-width="220">
                  <template #default="{ row }">
                    <code class="target-value">{{ row.value }}</code>
                  </template>
                </el-table-column>
                <el-table-column prop="description" :label="$t('common.description', 'Описание')" min-width="200" />
                <el-table-column :label="$t('common.enabled', 'Активен')" width="110">
                  <template #default="{ row }">
                    <el-switch v-model="row.enabled" @change="toggleTargetEnabled(row)" />
                  </template>
                </el-table-column>
                <el-table-column :label="$t('common.createTime', 'Создан')" width="170">
                  <template #default="{ row }">
                    <span class="time-text">{{ formatTime(row.createTime) }}</span>
                  </template>
                </el-table-column>
                <el-table-column :label="$t('common.actions', 'Действия')" width="160" fixed="right">
                  <template #default="{ row }">
                    <el-button text type="primary" size="small" @click="openTargetDialog(row)">
                      {{ $t('common.edit', 'Изм.') }}
                    </el-button>
                    <el-button text type="danger" size="small" @click="confirmDeleteTargets([row.id])">
                      {{ $t('common.delete', 'Удалить') }}
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>

              <div class="pagination-wrap">
                <el-pagination
                  v-model:current-page="targetPage"
                  v-model:page-size="targetPageSize"
                  :page-sizes="[20, 50, 100]"
                  :total="targetTotal"
                  layout="total, sizes, prev, pager, next, jumper"
                  background
                  @size-change="loadTargets"
                  @current-change="loadTargets"
                />
              </div>
            </el-tab-pane>

            <el-tab-pane :label="$t('organization.tabInfo', 'Информация')" name="info">
              <el-descriptions :column="2" border>
                <el-descriptions-item :label="$t('common.name', 'Название')">
                  {{ selectedOrg.name }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('common.status', 'Статус')">
                  <el-tag :type="selectedOrg.status === 'disable' ? 'info' : 'success'" size="small">
                    {{ selectedOrg.status === 'disable' ? $t('common.disabled', 'выкл') : $t('common.enabled', 'вкл') }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item :label="$t('common.description', 'Описание')" :span="2">
                  {{ selectedOrg.description || '—' }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('common.createTime', 'Создана')">
                  {{ formatTime(selectedOrg.createTime) }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('common.updateTime', 'Обновлена')">
                  {{ formatTime(selectedOrg.updateTime) }}
                </el-descriptions-item>
              </el-descriptions>
            </el-tab-pane>
          </el-tabs>
        </template>
      </el-card>
    </div>

    <!-- Диалог создания/редактирования организации -->
    <el-dialog
      v-model="orgDialogVisible"
      :title="orgForm.id ? $t('organization.editOrganization', 'Редактировать организацию') : $t('organization.newOrganization', 'Создать организацию')"
      width="540px"
      destroy-on-close
    >
      <el-form ref="orgFormRef" :model="orgForm" :rules="orgRules" label-position="top">
        <el-form-item :label="$t('common.name', 'Название')" prop="name">
          <el-input v-model="orgForm.name" maxlength="128" show-word-limit />
        </el-form-item>
        <el-form-item :label="$t('common.description', 'Описание')">
          <el-input v-model="orgForm.description" type="textarea" :rows="3" maxlength="512" show-word-limit />
        </el-form-item>
        <el-form-item :label="$t('common.status', 'Статус')">
          <el-radio-group v-model="orgForm.status">
            <el-radio value="enable">{{ $t('common.enabled', 'включена') }}</el-radio>
            <el-radio value="disable">{{ $t('common.disabled', 'выключена') }}</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="orgDialogVisible = false">{{ $t('common.cancel', 'Отмена') }}</el-button>
        <el-button type="primary" @click="saveOrg" :loading="orgSaveLoading">
          {{ $t('common.save', 'Сохранить') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Диалог создания/редактирования таргета -->
    <el-dialog
      v-model="targetDialogVisible"
      :title="targetForm.id ? $t('organization.editTarget', 'Изменить таргет') : $t('organization.addTarget', 'Добавить таргет')"
      width="500px"
      destroy-on-close
    >
      <el-form ref="targetFormRef" :model="targetForm" :rules="targetRules" label-position="top">
        <el-form-item :label="$t('organization.targetValue', 'Значение')" prop="value">
          <el-input v-model="targetForm.value" placeholder="example.com / 10.0.0.1 / 192.168.0.0/24 / *.example.com" />
          <div class="form-hint">
            {{ $t('organization.targetValueHint', 'Тип определяется автоматически: IP, CIDR, домен, wildcard.') }}
          </div>
        </el-form-item>
        <el-form-item :label="$t('common.description', 'Описание')">
          <el-input v-model="targetForm.description" maxlength="256" show-word-limit />
        </el-form-item>
        <el-form-item v-if="targetForm.id" :label="$t('common.enabled', 'Активен')">
          <el-switch v-model="targetForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="targetDialogVisible = false">{{ $t('common.cancel', 'Отмена') }}</el-button>
        <el-button type="primary" @click="saveTarget" :loading="targetSaveLoading">
          {{ $t('common.save', 'Сохранить') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Диалог импорта таргетов -->
    <el-dialog
      v-model="importDialogVisible"
      :title="$t('organization.importTargets', 'Импорт таргетов')"
      width="600px"
      destroy-on-close
    >
      <p class="form-hint">
        {{ $t('organization.importHint', 'Вставьте список значений (IP, CIDR, домен, wildcard) — по одному в строке или через запятую.') }}
      </p>
      <el-input
        v-model="importText"
        type="textarea"
        :rows="12"
        placeholder="example.com&#10;192.168.0.0/24&#10;10.0.0.5&#10;*.subdomain.com"
      />
      <template #footer>
        <el-button @click="importDialogVisible = false">{{ $t('common.cancel', 'Отмена') }}</el-button>
        <el-button type="primary" @click="doImport" :loading="importLoading">
          {{ $t('common.import', 'Импортировать') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search, More, Delete, Upload } from '@element-plus/icons-vue'
import {
  getOrganizationList,
  saveOrganization,
  deleteOrganization,
  updateOrganizationStatus,
  getOrgTargetList,
  saveOrgTargets,
  importOrgTargets,
  updateOrgTarget,
  deleteOrgTargets,
  retagAssets
} from '@/api/organization'

const orgs = ref([])
const orgLoading = ref(false)
const orgSearch = ref('')
const selectedOrg = ref(null)

const targets = ref([])
const targetTotal = ref(0)
const targetLoading = ref(false)
const targetPage = ref(1)
const targetPageSize = ref(20)
const targetSearch = ref('')
const targetTypeFilter = ref('')
const selectedTargetIds = ref([])

const activeTab = ref('targets')
const retagLoading = ref(false)

// Диалоги
const orgDialogVisible = ref(false)
const orgFormRef = ref(null)
const orgForm = ref({ id: '', name: '', description: '', status: 'enable' })
const orgSaveLoading = ref(false)

const targetDialogVisible = ref(false)
const targetFormRef = ref(null)
const targetForm = ref({ id: '', value: '', description: '', enabled: true })
const targetSaveLoading = ref(false)

const importDialogVisible = ref(false)
const importText = ref('')
const importLoading = ref(false)

const orgRules = {
  name: [
    { required: true, message: 'Название обязательно', trigger: 'blur' },
    { min: 2, max: 128, message: 'От 2 до 128 символов', trigger: 'blur' }
  ]
}
const targetRules = {
  value: [{ required: true, message: 'Значение обязательно', trigger: 'blur' }]
}

const filteredOrgs = computed(() => {
  if (!orgSearch.value) return orgs.value
  const q = orgSearch.value.toLowerCase()
  return orgs.value.filter(
    (o) => o.name.toLowerCase().includes(q) || (o.description || '').toLowerCase().includes(q)
  )
})

async function loadOrgs() {
  orgLoading.value = true
  try {
    const res = await getOrganizationList({ page: 1, pageSize: 1000 })
    if (res?.code === 0) {
      orgs.value = res.list || []
      if (selectedOrg.value) {
        const same = orgs.value.find((o) => o.id === selectedOrg.value.id)
        selectedOrg.value = same || null
      }
    } else {
      ElMessage.error(res?.msg || 'Ошибка загрузки')
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Ошибка загрузки')
  } finally {
    orgLoading.value = false
  }
}

function selectOrg(org) {
  selectedOrg.value = org
  targetPage.value = 1
  targetSearch.value = ''
  targetTypeFilter.value = ''
  loadTargets()
}

function openOrgDialog(org = null) {
  if (org) {
    orgForm.value = { ...org }
  } else {
    orgForm.value = { id: '', name: '', description: '', status: 'enable' }
  }
  orgDialogVisible.value = true
}

async function saveOrg() {
  if (!orgFormRef.value) return
  const ok = await orgFormRef.value.validate().catch(() => false)
  if (!ok) return
  orgSaveLoading.value = true
  try {
    const res = await saveOrganization({ ...orgForm.value })
    if (res?.code === 0) {
      ElMessage.success(res.msg || 'Сохранено')
      orgDialogVisible.value = false
      await loadOrgs()
    } else {
      ElMessage.error(res?.msg || 'Ошибка сохранения')
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Ошибка сохранения')
  } finally {
    orgSaveLoading.value = false
  }
}

async function toggleOrgStatus(org) {
  const newStatus = org.status === 'disable' ? 'enable' : 'disable'
  try {
    const res = await updateOrganizationStatus({ id: org.id, status: newStatus })
    if (res?.code === 0) {
      ElMessage.success(res.msg || 'Готово')
      org.status = newStatus
    } else {
      ElMessage.error(res?.msg || 'Ошибка')
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Ошибка')
  }
}

async function confirmDeleteOrg(org) {
  try {
    await ElMessageBox.confirm(
      `Удалить организацию "${org.name}"? Все её таргеты также будут удалены, а ассеты потеряют привязку.`,
      'Подтверждение',
      { type: 'warning', confirmButtonText: 'Удалить', cancelButtonText: 'Отмена' }
    )
    const res = await deleteOrganization({ id: org.id })
    if (res?.code === 0) {
      ElMessage.success(res.msg || 'Удалено')
      if (selectedOrg.value?.id === org.id) selectedOrg.value = null
      await loadOrgs()
    } else {
      ElMessage.error(res?.msg || 'Ошибка удаления')
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e?.message || 'Ошибка')
  }
}

let targetSearchTimer = null
function debouncedLoadTargets() {
  if (targetSearchTimer) clearTimeout(targetSearchTimer)
  targetSearchTimer = setTimeout(() => {
    targetPage.value = 1
    loadTargets()
  }, 300)
}

async function loadTargets() {
  if (!selectedOrg.value) return
  targetLoading.value = true
  try {
    const res = await getOrgTargetList({
      orgId: selectedOrg.value.id,
      search: targetSearch.value || undefined,
      type: targetTypeFilter.value || undefined,
      page: targetPage.value,
      pageSize: targetPageSize.value
    })
    if (res?.code === 0) {
      targets.value = res.list || []
      targetTotal.value = res.total || 0
    } else {
      ElMessage.error(res?.msg || 'Ошибка')
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Ошибка')
  } finally {
    targetLoading.value = false
  }
}

function onTargetSelect(rows) {
  selectedTargetIds.value = rows.map((r) => r.id)
}

function openTargetDialog(target = null) {
  if (target) {
    targetForm.value = { ...target }
  } else {
    targetForm.value = { id: '', value: '', description: '', enabled: true }
  }
  targetDialogVisible.value = true
}

async function saveTarget() {
  if (!targetFormRef.value) return
  const ok = await targetFormRef.value.validate().catch(() => false)
  if (!ok) return
  targetSaveLoading.value = true
  try {
    let res
    if (targetForm.value.id) {
      res = await updateOrgTarget({
        id: targetForm.value.id,
        value: targetForm.value.value,
        description: targetForm.value.description,
        enabled: targetForm.value.enabled
      })
    } else {
      res = await saveOrgTargets({
        orgId: selectedOrg.value.id,
        items: [
          {
            value: targetForm.value.value.trim(),
            description: targetForm.value.description || ''
          }
        ]
      })
    }
    if (res?.code === 0) {
      ElMessage.success(res.msg || 'Сохранено')
      targetDialogVisible.value = false
      await loadTargets()
    } else {
      ElMessage.error(res?.msg || 'Ошибка')
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Ошибка')
  } finally {
    targetSaveLoading.value = false
  }
}

async function toggleTargetEnabled(row) {
  try {
    const res = await updateOrgTarget({ id: row.id, enabled: row.enabled })
    if (res?.code === 0) {
      ElMessage.success('Обновлено')
    } else {
      ElMessage.error(res?.msg || 'Ошибка')
      row.enabled = !row.enabled
    }
  } catch (e) {
    row.enabled = !row.enabled
    ElMessage.error(e?.message || 'Ошибка')
  }
}

async function confirmDeleteTargets(ids) {
  const list = ids || selectedTargetIds.value
  if (!list.length) return
  try {
    await ElMessageBox.confirm(`Удалить ${list.length} таргет(ов)?`, 'Подтверждение', {
      type: 'warning',
      confirmButtonText: 'Удалить',
      cancelButtonText: 'Отмена'
    })
    const res = await deleteOrgTargets({ ids: list })
    if (res?.code === 0) {
      ElMessage.success(res.msg || 'Удалено')
      selectedTargetIds.value = []
      await loadTargets()
    } else {
      ElMessage.error(res?.msg || 'Ошибка')
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e?.message || 'Ошибка')
  }
}

async function doImport() {
  if (!importText.value.trim()) {
    ElMessage.warning('Введите хотя бы один таргет')
    return
  }
  importLoading.value = true
  try {
    const res = await importOrgTargets({ orgId: selectedOrg.value.id, text: importText.value })
    if (res?.code === 0) {
      ElMessage.success(res.msg || `Импортировано: ${res.inserted || 0}`)
      importDialogVisible.value = false
      importText.value = ''
      await loadTargets()
    } else {
      ElMessage.error(res?.msg || 'Ошибка')
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Ошибка')
  } finally {
    importLoading.value = false
  }
}

async function reloadAfterRetag() {
  if (!selectedOrg.value) return
  retagLoading.value = true
  try {
    const res = await retagAssets({ orgId: selectedOrg.value.id })
    if (res?.code === 0) {
      ElMessage.success(res.msg || `Найдено: ${res.matched || 0}, обновлено: ${res.updated || 0}`)
    } else {
      ElMessage.error(res?.msg || 'Ошибка')
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Ошибка')
  } finally {
    retagLoading.value = false
  }
}

function targetTypeLabel(t) {
  return { ip: 'IP', cidr: 'CIDR', domain: 'Домен', wildcard: 'Wildcard' }[t] || t
}
function targetTypeColor(t) {
  return { ip: 'success', cidr: 'warning', domain: 'primary', wildcard: 'info' }[t] || 'info'
}
function formatTime(s) {
  if (!s) return '—'
  try {
    return new Date(s).toLocaleString('ru-RU')
  } catch {
    return s
  }
}

onMounted(() => {
  loadOrgs()
})
</script>

<style scoped>
.organization-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
  height: 100%;
  box-sizing: border-box;
}

.page-header :deep(.el-card__body) {
  padding: 18px 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.page-title {
  margin: 0 0 4px 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.page-desc {
  margin: 0;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  max-width: 720px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.page-body {
  flex: 1;
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: 16px;
  min-height: 0;
}

.org-list-card,
.org-detail-card {
  display: flex;
  flex-direction: column;
}

.org-list-card :deep(.el-card__body) {
  flex: 1;
  overflow: auto;
  padding: 0;
}

.org-detail-card :deep(.el-card__body) {
  padding: 16px 20px;
  flex: 1;
  overflow: auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.org-list {
  padding: 8px;
}

.org-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.15s;
  margin-bottom: 4px;
}

.org-item:hover {
  background-color: var(--el-fill-color-light);
}

.org-item.active {
  background-color: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary-light-7);
}

.org-item-main {
  flex: 1;
  min-width: 0;
}

.org-name {
  font-weight: 500;
  display: flex;
  align-items: center;
  margin-bottom: 4px;
}

.org-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.detail-title {
  font-size: 16px;
  font-weight: 600;
}

.target-toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.target-value {
  font-family: var(--el-font-family-mono, monospace);
  font-size: 13px;
  color: var(--el-color-primary);
}

.time-text {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.form-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>
