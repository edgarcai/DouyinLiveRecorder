import {createApp} from 'vue'
import {createI18n} from 'vue-i18n'
import App from './App.vue'
import './style.css';
import en from './locales/en.json'
import zh from './locales/zh.json'

const i18n = createI18n({
  legacy: false,
  locale: 'zh', // default locale
  fallbackLocale: 'en',
  messages: {
    en,
    zh
  }
})

createApp(App).use(i18n).mount('#app')
