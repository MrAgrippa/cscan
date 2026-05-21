<template>
  <div class="test-page">
    <div class="dark-card">
      <div class="card-header">
        <span class="card-title">Тестовая страница</span>
        <el-icon class="card-icon"><Tools /></el-icon>
      </div>

      <div class="test-section">
        <h3>Тест темы</h3>
        <p>Текущая тема: {{ themeStore.theme }}</p>
        <p>Цветовая тема: {{ themeStore.colorTheme }}</p>
        <p>Тёмный режим: {{ themeStore.isDark }}</p>

        <div class="button-group">
          <el-button @click="themeStore.setTheme('light')">Светлая</el-button>
          <el-button @click="themeStore.setTheme('dark')">Тёмная</el-button>
          <el-button @click="themeStore.setTheme('system')">Системная</el-button>
        </div>
      </div>

      <div class="test-section">
        <h3>Тест языка</h3>
        <p>Текущий язык: {{ localeStore.currentLocale }}</p>
        <p>Перевод: {{ $t('common.success') }}</p>

        <div class="button-group">
          <el-button @click="localeStore.changeLocale('ru-RU')">Русский</el-button>
          <el-button @click="localeStore.changeLocale('en-US')">English</el-button>
        </div>
      </div>

      <div class="test-section">
        <h3>Тест компонентов</h3>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-card>
              <template #header>Формы</template>
              <el-form label-width="100px">
                <el-form-item label="Поле ввода">
                  <el-input placeholder="Введите значение" />
                </el-form-item>
                <el-form-item label="Выбор">
                  <el-select placeholder="Выберите">
                    <el-option label="Вариант 1" value="1" />
                    <el-option label="Вариант 2" value="2" />
                  </el-select>
                </el-form-item>
                <el-form-item label="Переключатель">
                  <el-switch />
                </el-form-item>
              </el-form>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card>
              <template #header>Кнопки</template>
              <div class="button-demo">
                <el-button type="primary">Основная</el-button>
                <el-button type="success">Успех</el-button>
                <el-button type="warning">Внимание</el-button>
                <el-button type="danger">Опасно</el-button>
                <el-button type="info">Инфо</el-button>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useThemeStore } from '@/stores/theme'
import { useLocaleStore } from '@/stores/locale'
import { Tools } from '@element-plus/icons-vue'

const themeStore = useThemeStore()
const localeStore = useLocaleStore()
</script>

<style scoped>
.test-page {
  .test-section {
    margin-bottom: 32px;

    h3 {
      color: hsl(var(--foreground));
      margin-bottom: 16px;
      font-size: 16px;
    }

    p {
      color: hsl(var(--muted-foreground));
      margin-bottom: 8px;
    }
  }

  .button-group {
    margin-top: 16px;
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .button-demo {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
}
</style>
