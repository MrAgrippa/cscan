import request from '@/api/request'

// ==================== Организации ====================

// Список организаций
export function getOrganizationList(data) {
  return request({
    url: '/organization/list',
    method: 'post',
    data
  })
}

// Создать или обновить организацию
export function saveOrganization(data) {
  return request({
    url: '/organization/save',
    method: 'post',
    data
  })
}

// Удалить организацию
export function deleteOrganization(data) {
  return request({
    url: '/organization/delete',
    method: 'post',
    data
  })
}

// Изменить статус организации
export function updateOrganizationStatus(data) {
  return request({
    url: '/organization/updateStatus',
    method: 'post',
    data
  })
}

// ==================== Таргеты организации ====================

// Получить список таргетов организации
export function getOrgTargetList(data) {
  return request({
    url: '/organization/target/list',
    method: 'post',
    data
  })
}

// Сохранить таргеты (массив)
export function saveOrgTargets(data) {
  return request({
    url: '/organization/target/save',
    method: 'post',
    data
  })
}

// Импортировать таргеты текстом
export function importOrgTargets(data) {
  return request({
    url: '/organization/target/import',
    method: 'post',
    data
  })
}

// Обновить один таргет
export function updateOrgTarget(data) {
  return request({
    url: '/organization/target/update',
    method: 'post',
    data
  })
}

// Удалить таргет(ы)
export function deleteOrgTargets(data) {
  return request({
    url: '/organization/target/delete',
    method: 'post',
    data
  })
}

// ==================== Привязка ассетов к организации ====================

// Назначить ассеты организации (или отвязать, передав пустой orgId)
export function assignAssetsToOrg(data) {
  return request({
    url: '/organization/asset/assign',
    method: 'post',
    data
  })
}

// Запустить ре-индексацию (re-tag) ассетов по таргетам
export function retagAssets(data) {
  return request({
    url: '/organization/retag',
    method: 'post',
    data
  })
}
