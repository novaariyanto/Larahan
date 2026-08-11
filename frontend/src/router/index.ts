import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: () => import('../views/DashboardView.vue') },
    { path: '/apache', name: 'apache', component: () => import('../views/ApacheView.vue') },
    { path: '/php', name: 'php', component: () => import('../views/PHPView.vue') },
    { path: '/mysql', name: 'mysql', component: () => import('../views/MySQLView.vue') },
    { path: '/phpmyadmin', name: 'phpmyadmin', component: () => import('../views/PhpMyAdminView.vue') },
    { path: '/settings', name: 'settings', component: () => import('../views/SettingsView.vue') },
  ],
})

export default router
