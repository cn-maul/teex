import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './style.css'

// Chart.js 统一注册（全局只注册一次）
import { Chart as ChartJS, ArcElement, CategoryScale, LinearScale, PointElement, LineElement, BarElement, RadialLinearScale, Tooltip, Legend, Filler } from 'chart.js'
ChartJS.register(ArcElement, CategoryScale, LinearScale, PointElement, LineElement, BarElement, RadialLinearScale, Tooltip, Legend, Filler)

createApp(App).use(router).mount('#app')
