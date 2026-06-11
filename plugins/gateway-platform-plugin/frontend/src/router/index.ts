import { createRouter, createWebHistory } from 'vue-router'
import DashboardPage from '../pages/DashboardPage.vue'
import RoutesPage from '../pages/RoutesPage.vue'
import KeysPage from '../pages/KeysPage.vue'
import ReferencesPage from '../pages/ReferencesPage.vue'

export default createRouter({
  history: createWebHistory('/app/'),
  routes: [
    { path: '/', component: DashboardPage },
    { path: '/routes', component: RoutesPage },
    { path: '/keys', component: KeysPage },
    { path: '/references', component: ReferencesPage },
  ],
})
