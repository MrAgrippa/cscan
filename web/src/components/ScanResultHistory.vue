<template>
  <div class="scan-result-history">
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>{{ t('asset.history.loading') }}</span>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-container">
      <el-icon><WarningFilled /></el-icon>
      <span>{{ t('asset.history.loadError') }}</span>
      <el-button type="primary" size="small" @click="loadHistory">
        <el-icon><Refresh /></el-icon>
        {{ t('asset.history.retry') }}
      </el-button>
    </div>

    <!-- 空状态 -->
    <div v-else-if="!versions || versions.length === 0" class="empty-container">
      <el-empty :description="t('asset.history.noHistory')" />
    </div>

    <!-- 时间线视图 -->
    <div v-else class="timeline-container">
      <!-- 比较按钮 -->
      <div v-if="selectedVersions.length === 2" class="compare-actions">
        <el-button type="primary" @click="compareSelectedVersions">
          <el-icon><Connection /></el-icon>
          {{ t('asset.history.compareSelected') }}
        </el-button>
        <el-button @click="clearSelection">
          {{ t('asset.history.clearSelection') }}
        </el-button>
      </div>

      <el-timeline>
        <el-timeline-item
          v-for="(version, index) in versions"
          :key="version.versionId"
          :timestamp="formatTimestamp(version.scanTimestamp)"
          placement="top"
          :type="index === 0 ? 'primary' : 'default'"
        >
          <el-card class="version-card" :class="{ 'latest-version': index === 0, 'selected': isVersionSelected(version.versionId) }">
            <!-- 版本头部 -->
            <div class="version-header">
              <div class="version-info">
                <!-- 选择框 -->
                <el-checkbox
                  v-model="selectedVersions"
                  :value="version.versionId"
                  :disabled="selectedVersions.length >= 2 && !isVersionSelected(version.versionId)"
                  @change="toggleVersionSelection(version.versionId)"
                />
                <el-tag v-if="index === 0" type="success" size="small">
                  {{ t('asset.history.latest') }}
                </el-tag>
                <span class="version-time">{{ formatDateTime(version.scanTimestamp) }}</span>
              </div>
              <el-button
                type="primary"
                size="small"
                @click="viewVersionDetails(version)"
              >
                <el-icon><View /></el-icon>
                {{ t('asset.history.viewDetails') }}
              </el-button>
            </div>

            <!-- 扫描统计 -->
            <div class="version-stats">
              <div class="stat-item">
                <el-icon class="stat-icon"><FolderOpened /></el-icon>
                <div class="stat-content">
                  <span class="stat-label">{{ t('asset.history.dirScans') }}</span>
                  <span class="stat-value">{{ version.dirScanCount }}</span>
                </div>
              </div>
              <div class="stat-item">
                <el-icon class="stat-icon"><Warning /></el-icon>
                <div class="stat-content">
                  <span class="stat-label">{{ t('asset.history.vulnScans') }}</span>
                  <span class="stat-value">{{ version.vulnScanCount }}</span>
                </div>
              </div>
            </div>

            <!-- 变更摘要 -->
            <div v-if="version.changesSummary" class="version-changes">
              <div class="changes-label">
                <el-icon><DocumentCopy /></el-icon>
                {{ t('asset.history.changesSummary') }}:
              </div>
              <div class="changes-text">{{ version.changesSummary }}</div>
            </div>
          </el-card>
        </el-timeline-item>
      </el-timeline>
    </div>

    <!-- 版本详情对话框 -->
    <el-dialog
      v-model="showDetailsDialog"
      :title="t('asset.history.versionDetails')"
      width="80%"
      :close-on-click-modal="false"
    >
      <div v-if="selectedVersion" class="version-details">
        <!-- 版本信息 -->
        <div class="detail-section">
          <h4 class="section-title">{{ t('asset.history.versionInfo') }}</h4>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">{{ t('asset.history.scanTime') }}:</span>
              <span class="info-value">{{ formatDateTime(selectedVersion.scanTimestamp) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('asset.history.versionId') }}:</span>
              <span class="info-value">{{ selectedVersion.versionId }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('asset.history.dirScans') }}:</span>
              <span class="info-value">{{ selectedVersion.dirScanCount }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('asset.history.vulnScans') }}:</span>
              <span class="info-value">{{ selectedVersion.vulnScanCount }}</span>
            </div>
          </div>
        </div>

        <!-- 变更摘要 -->
        <div v-if="selectedVersion.changesSummary" class="detail-section">
          <h4 class="section-title">{{ t('asset.history.changesSummary') }}</h4>
          <div class="changes-detail">{{ selectedVersion.changesSummary }}</div>
        </div>
      </div>

      <template #footer>
        <el-button @click="showDetailsDialog = false">{{ t('common.close') }}</el-button>
      </template>
    </el-dialog>

    <!-- 版本比较对话框 -->
    <el-dialog
      v-model="showComparisonDialog"
      :title="t('asset.history.versionComparison')"
      width="90%"
      :close-on-click-modal="false"
    >
      <div v-if="comparisonData" class="comparison-container">
        <!-- 比较摘要 -->
        <div class="comparison-summary">
          <el-alert
            :title="t('asset.history.comparisonSummary')"
            type="info"
            :closable="false"
          >
            <div class="summary-content">
              <div class="summary-item">
                <el-icon class="summary-icon"><FolderOpened /></el-icon>
                <span>{{ t('asset.history.dirScansChange') }}: </span>
                <span v-if="comparisonData.dirScansAdded > 0" class="change-added">
                  +{{ comparisonData.dirScansAdded }}
                </span>
                <span v-if="comparisonData.dirScansRemoved > 0" class="change-removed">
                  -{{ comparisonData.dirScansRemoved }}
                </span>
                <span v-if="comparisonData.dirScansAdded === 0 && comparisonData.dirScansRemoved === 0" class="change-none">
                  {{ t('asset.history.noChange') }}
                </span>
              </div>
              <div class="summary-item">
                <el-icon class="summary-icon"><Warning /></el-icon>
                <span>{{ t('asset.history.vulnsChange') }}: </span>
                <span v-if="comparisonData.vulnsAdded > 0" class="change-added">
                  +{{ comparisonData.vulnsAdded }}
                </span>
                <span v-if="comparisonData.vulnsRemoved > 0" class="change-removed">
                  -{{ comparisonData.vulnsRemoved }}
                </span>
                <span v-if="comparisonData.vulnsAdded === 0 && comparisonData.vulnsRemoved === 0" class="change-none">
                  {{ t('asset.history.noChange') }}
                </span>
              </div>
            </div>
          </el-alert>
        </div>

        <!-- 并排比较 -->
        <div class="side-by-side-comparison">
          <!-- 版本1 -->
          <div class="version-column">
            <div class="column-header">
              <h4>{{ t('asset.history.version') }} 1</h4>
              <el-tag type="info" size="small">
                {{ formatDateTime(comparisonData.version1.scanTimestamp) }}
              </el-tag>
            </div>
            <div class="version-stats-grid">
              <div class="stat-card">
                <div class="stat-label">{{ t('asset.history.dirScans') }}</div>
                <div class="stat-value">{{ comparisonData.version1.dirScanCount }}</div>
              </div>
              <div class="stat-card">
                <div class="stat-label">{{ t('asset.history.vulnScans') }}</div>
                <div class="stat-value">{{ comparisonData.version1.vulnScanCount }}</div>
              </div>
            </div>
            <div v-if="comparisonData.version1.changesSummary" class="version-summary">
              <div class="summary-label">{{ t('asset.history.changesSummary') }}:</div>
              <div class="summary-text">{{ comparisonData.version1.changesSummary }}</div>
            </div>
          </div>

          <!-- 分隔符 -->
          <div class="comparison-divider">
            <el-icon class="divider-icon"><Right /></el-icon>
          </div>

          <!-- 版本2 -->
          <div class="version-column">
            <div class="column-header">
              <h4>{{ t('asset.history.version') }} 2</h4>
              <el-tag type="info" size="small">
                {{ formatDateTime(comparisonData.version2.scanTimestamp) }}
              </el-tag>
            </div>
            <div class="version-stats-grid">
              <div class="stat-card">
                <div class="stat-label">{{ t('asset.history.dirScans') }}</div>
                <div class="stat-value">{{ comparisonData.version2.dirScanCount }}</div>
              </div>
              <div class="stat-card">
                <div class="stat-label">{{ t('asset.history.vulnScans') }}</div>
                <div class="stat-value">{{ comparisonData.version2.vulnScanCount }}</div>
              </div>
            </div>
            <div v-if="comparisonData.version2.changesSummary" class="version-summary">
              <div class="summary-label">{{ t('asset.history.changesSummary') }}:</div>
              <div class="summary-text">{{ comparisonData.version2.changesSummary }}</div>
            </div>
          </div>
        </div>

        <!-- 详细比较信息 -->
        <div v-if="comparisonData.comparisonDetail" class="comparison-detail">
          <h4 class="detail-title">{{ t('asset.history.detailedComparison') }}</h4>
          <pre class="detail-content">{{ comparisonData.comparisonDetail }}</pre>
        </div>
      </div>

      <template #footer>
        <el-button @click="showComparisonDialog = false">{{ t('common.close') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import {
  Loading,
  WarningFilled,
  Refresh,
  View,
  FolderOpened,
  Warning,
  DocumentCopy,
  Connection,
  Right
} from '@element-plus/icons-vue'
import { getAssetHistory, compareVersions } from '@/api/asset'

const { t } = useI18n()

// Props
const props = defineProps({
  assetId: {
    type: String,
    required: true
  },
  workspaceId: {
    type: String,
    default: ''
  },
  authority: {
    type: String,
    default: ''
  },
  host: {
    type: String,
    default: ''
  },
  port: {
    type: Number,
    default: 0
  }
})

// State
const loading = ref(false)
const error = ref(false)
const versions = ref([])
const showDetailsDialog = ref(false)
const selectedVersion = ref(null)
const selectedVersions = ref([])
const showComparisonDialog = ref(false)
const comparisonData = ref(null)

// 加载历史记录
const loadHistory = async () => {
  if (!props.assetId) {
    return
  }

  loading.value = true
  error.value = false

  try {
    const res = await getAssetHistory({
      assetId: props.assetId,
      workspaceId: props.workspaceId,
      authority: props.authority,
      host: props.host,
      port: props.port,
      limit: 50
    })

    if (res.code === 0) {
      versions.value = res.versions || []
    } else {
      error.value = true
      ElMessage.error(res.msg || t('asset.history.loadError'))
    }
  } catch (err) {
    error.value = true
    console.error('Failed to load scan history:', err)
    ElMessage.error(t('asset.history.loadError'))
  } finally {
    loading.value = false
  }
}

// 查看版本详情
const viewVersionDetails = (version) => {
  selectedVersion.value = version
  showDetailsDialog.value = true
}

// 切换版本选择
const toggleVersionSelection = (versionId) => {
  const index = selectedVersions.value.indexOf(versionId)
  if (index > -1) {
    selectedVersions.value.splice(index, 1)
  } else if (selectedVersions.value.length < 2) {
    selectedVersions.value.push(versionId)
  }
}

// 检查版本是否被选中
const isVersionSelected = (versionId) => {
  return selectedVersions.value.includes(versionId)
}

// 清除选择
const clearSelection = () => {
  selectedVersions.value = []
}

// 比较选中的版本
const compareSelectedVersions = async () => {
  if (selectedVersions.value.length !== 2) {
    ElMessage.warning(t('asset.history.selectTwoVersions'))
    return
  }

  try {
    const res = await compareVersions({
      versionId1: selectedVersions.value[0],
      versionId2: selectedVersions.value[1]
    })

    if (res.code === 0) {
      comparisonData.value = res
      showComparisonDialog.value = true
    } else {
      ElMessage.error(res.msg || t('asset.history.compareError'))
    }
  } catch (err) {
    console.error('Failed to compare versions:', err)
    ElMessage.error(t('asset.history.compareError'))
  }
}

// 格式化时间戳（用于时间线）
const formatTimestamp = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleString('ru-RU', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 格式化日期时间（完整格式）
const formatDateTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleString('ru-RU', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 监听 assetId 变化
watch(() => props.assetId, () => {
  if (props.assetId) {
    loadHistory()
  }
}, { immediate: true })

// 组件挂载时加载数据
onMounted(() => {
  if (props.assetId) {
    loadHistory()
  }
})
</script>

<style lang="scss" scoped>
.scan-result-history {
  padding: 16px 0;

  .loading-container,
  .error-container,
  .empty-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 48px 20px;
    gap: 16px;
    color: hsl(var(--muted-foreground));

    .el-icon {
      font-size: 48px;
    }

    .el-button {
      margin-top: 8px;
    }
  }

  .timeline-container {
    padding: 0 16px;

    .compare-actions {
      display: flex;
      gap: 12px;
      margin-bottom: 24px;
      padding: 16px;
      background: hsl(var(--muted) / 0.3);
      border-radius: 8px;
      border: 1px solid hsl(var(--border));
    }

    :deep(.el-timeline) {
      padding-left: 0;
    }

    :deep(.el-timeline-item__timestamp) {
      font-size: 13px;
      color: hsl(var(--muted-foreground));
      margin-bottom: 8px;
    }

    .version-card {
      margin-bottom: 8px;
      border: 1px solid hsl(var(--border));
      transition: all 0.2s;

      &.latest-version {
        border-color: hsl(var(--primary));
        background: hsl(var(--primary) / 0.05);
      }

      &.selected {
        border-color: hsl(var(--primary));
        background: hsl(var(--primary) / 0.1);
        box-shadow: 0 2px 8px hsl(var(--primary) / 0.2);
      }

      &:hover {
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      }

      :deep(.el-card__body) {
        padding: 16px;
      }

      .version-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;

        .version-info {
          display: flex;
          align-items: center;
          gap: 12px;

          .version-time {
            font-size: 14px;
            font-weight: 500;
            color: hsl(var(--foreground));
          }
        }
      }

      .version-stats {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 16px;
        margin-bottom: 16px;

        .stat-item {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 12px;
          background: hsl(var(--muted) / 0.3);
          border-radius: 6px;

          .stat-icon {
            font-size: 24px;
            color: hsl(var(--primary));
          }

          .stat-content {
            display: flex;
            flex-direction: column;
            gap: 4px;

            .stat-label {
              font-size: 12px;
              color: hsl(var(--muted-foreground));
            }

            .stat-value {
              font-size: 18px;
              font-weight: 600;
              color: hsl(var(--foreground));
            }
          }
        }
      }

      .version-changes {
        padding: 12px;
        background: hsl(var(--muted) / 0.2);
        border-radius: 6px;
        border-left: 3px solid hsl(var(--primary));

        .changes-label {
          display: flex;
          align-items: center;
          gap: 6px;
          font-size: 13px;
          font-weight: 500;
          color: hsl(var(--foreground));
          margin-bottom: 6px;

          .el-icon {
            font-size: 16px;
          }
        }

        .changes-text {
          font-size: 13px;
          color: hsl(var(--muted-foreground));
          line-height: 1.6;
        }
      }
    }
  }

  .version-details {
    .detail-section {
      margin-bottom: 24px;

      &:last-child {
        margin-bottom: 0;
      }

      .section-title {
        font-size: 16px;
        font-weight: 600;
        color: hsl(var(--foreground));
        margin: 0 0 16px 0;
        padding-bottom: 8px;
        border-bottom: 1px solid hsl(var(--border));
      }

      .info-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 16px;

        .info-item {
          display: flex;
          gap: 12px;

          .info-label {
            font-weight: 500;
            color: hsl(var(--muted-foreground));
            min-width: 100px;
          }

          .info-value {
            color: hsl(var(--foreground));
            word-break: break-all;
          }
        }
      }

      .changes-detail {
        padding: 16px;
        background: hsl(var(--muted) / 0.3);
        border-radius: 6px;
        font-size: 14px;
        color: hsl(var(--foreground));
        line-height: 1.6;
        white-space: pre-wrap;
      }
    }
  }

  // 版本比较样式
  .comparison-container {
    .comparison-summary {
      margin-bottom: 24px;

      .summary-content {
        display: flex;
        flex-direction: column;
        gap: 12px;
        margin-top: 12px;

        .summary-item {
          display: flex;
          align-items: center;
          gap: 8px;
          font-size: 14px;

          .summary-icon {
            font-size: 18px;
            color: hsl(var(--primary));
          }

          .change-added {
            color: #67c23a;
            font-weight: 600;
          }

          .change-removed {
            color: #f56c6c;
            font-weight: 600;
          }

          .change-none {
            color: hsl(var(--muted-foreground));
          }
        }
      }
    }

    .side-by-side-comparison {
      display: grid;
      grid-template-columns: 1fr auto 1fr;
      gap: 24px;
      margin-bottom: 24px;

      .version-column {
        padding: 20px;
        background: hsl(var(--muted) / 0.2);
        border-radius: 8px;
        border: 1px solid hsl(var(--border));

        .column-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
          padding-bottom: 12px;
          border-bottom: 2px solid hsl(var(--border));

          h4 {
            margin: 0;
            font-size: 16px;
            font-weight: 600;
            color: hsl(var(--foreground));
          }
        }

        .version-stats-grid {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: 12px;
          margin-bottom: 16px;

          .stat-card {
            padding: 12px;
            background: hsl(var(--background));
            border-radius: 6px;
            border: 1px solid hsl(var(--border));

            .stat-label {
              font-size: 12px;
              color: hsl(var(--muted-foreground));
              margin-bottom: 6px;
            }

            .stat-value {
              font-size: 20px;
              font-weight: 600;
              color: hsl(var(--foreground));
            }
          }
        }

        .version-summary {
          padding: 12px;
          background: hsl(var(--background));
          border-radius: 6px;
          border-left: 3px solid hsl(var(--primary));

          .summary-label {
            font-size: 13px;
            font-weight: 500;
            color: hsl(var(--foreground));
            margin-bottom: 6px;
          }

          .summary-text {
            font-size: 13px;
            color: hsl(var(--muted-foreground));
            line-height: 1.6;
          }
        }
      }

      .comparison-divider {
        display: flex;
        align-items: center;
        justify-content: center;

        .divider-icon {
          font-size: 32px;
          color: hsl(var(--primary));
        }
      }
    }

    .comparison-detail {
      padding: 20px;
      background: hsl(var(--muted) / 0.2);
      border-radius: 8px;
      border: 1px solid hsl(var(--border));

      .detail-title {
        margin: 0 0 16px 0;
        font-size: 16px;
        font-weight: 600;
        color: hsl(var(--foreground));
        padding-bottom: 12px;
        border-bottom: 2px solid hsl(var(--border));
      }

      .detail-content {
        margin: 0;
        padding: 16px;
        background: hsl(var(--background));
        border-radius: 6px;
        font-size: 13px;
        color: hsl(var(--foreground));
        line-height: 1.6;
        white-space: pre-wrap;
        font-family: 'Courier New', monospace;
      }
    }
  }
}
</style>
