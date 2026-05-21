import { createI18n } from 'vue-i18n'

// Импорт языковых файлов
import ruRU from './locales/ru-RU.json'
import enUS from './locales/en-US.json'

export const SUPPORT_LOCALES = ['ru-RU', 'en-US']
export const DEFAULT_LOCALE = 'ru-RU'

// Определение языка по браузеру
function getDefaultLocale() {
  const locale = navigator.language
  if (locale && locale.toLowerCase().startsWith('en')) {
    return 'en-US'
  }
  return 'ru-RU'
}

// Создание экземпляра i18n
export const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('locale') || getDefaultLocale(),
  fallbackLocale: 'en-US',
  globalInjection: true,
  messages: {
    'ru-RU': ruRU,
    'en-US': enUS,
  },
})

// Переключение языка
export function setLocale(locale) {
  i18n.global.locale.value = locale
  localStorage.setItem('locale', locale)
  document.documentElement.lang = locale
}

// Получение текущего языка
export function getLocale() {
  return i18n.global.locale.value
}

// Установка плагина
export function setupI18n(app) {
  app.use(i18n)
}

export default i18n
