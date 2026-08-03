import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import { createApp } from 'vue'
import { createVuetify } from 'vuetify'

import App from './App.vue'

const vuetify = createVuetify({
  theme: {
    defaultTheme: 'dark',
  },
})

createApp(App).use(vuetify).mount('#app')
