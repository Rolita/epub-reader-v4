import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import App from './App.vue'
import './style.css'
import { useThemeStore } from './stores/theme'

const app = createApp(App)
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

app.use(pinia)
app.mount('#app')

// 初始化主题
const themeStore = useThemeStore()
themeStore.initTheme()