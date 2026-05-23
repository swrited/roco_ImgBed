import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guest: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { guest: true },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('@/views/ForgotPasswordView.vue'),
      meta: { guest: true },
    },
    {
      path: '/reset-password/:token',
      name: 'reset-password',
      component: () => import('@/views/ResetPasswordView.vue'),
      meta: { guest: true },
    },
    // Authenticated user routes
    {
      path: '/',
      component: () => import('@/layouts/AppLayout.vue'),
      children: [
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/DashboardView.vue'),
          meta: { requiresAuth: true },
        },
        {
          path: 'upload',
          name: 'upload',
          component: () => import('@/views/UploadView.vue'),
          meta: { requiresAuth: true },
        },
        {
          path: 'ai-generate',
          name: 'ai-generate',
          component: () => import('@/views/AIGenerateView.vue'),
          meta: { requiresAuth: true },
        },
        {
          path: 'images',
          name: 'images',
          component: () => import('@/views/ImagesView.vue'),
          meta: { requiresAuth: true },
        },
        {
          path: 'albums',
          name: 'albums',
          component: () => import('@/views/AlbumsView.vue'),
          meta: { requiresAuth: true },
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/views/SettingsView.vue'),
          meta: { requiresAuth: true },
        },
        {
          path: 'gallery',
          name: 'gallery',
          component: () => import('@/views/GalleryView.vue'),
        },
        {
          path: 'api-doc',
          name: 'api-doc',
          component: () => import('@/views/ApiDocView.vue'),
        },
        {
          path: 'api-test',
          name: 'api-test',
          component: () => import('@/views/ApiTestView.vue'),
        },
        {
          path: 'api-keys',
          name: 'api-keys',
          component: () => import('@/views/ApiKeysView.vue'),
          meta: { requiresAuth: true },
        },
        {
          path: 'api-usage',
          name: 'api-usage',
          component: () => import('@/views/ApiUsageView.vue'),
          meta: { requiresAuth: true },
        },
        // Admin routes
        {
          path: 'admin',
          component: () => import('@/layouts/AdminLayout.vue'),
          meta: { requiresAuth: true, requiresAdmin: true },
          children: [
            {
              path: '',
              name: 'admin.console',
              component: () => import('@/views/admin/ConsoleView.vue'),
            },
            {
              path: 'users',
              name: 'admin.users',
              component: () => import('@/views/admin/UsersView.vue'),
            },
            {
              path: 'images',
              name: 'admin.images',
              component: () => import('@/views/admin/ImagesView.vue'),
            },
            {
              path: 'groups',
              name: 'admin.groups',
              component: () => import('@/views/admin/GroupsView.vue'),
            },
            {
              path: 'strategies',
              name: 'admin.strategies',
              component: () => import('@/views/admin/StrategiesView.vue'),
            },
            {
              path: 'settings',
              name: 'admin.settings',
              component: () => import('@/views/admin/SettingsView.vue'),
            },
            {
              path: 'api-usage',
              name: 'admin.api-usage',
              component: () => import('@/views/admin/ApiUsageView.vue'),
            },
          ],
        },
      ],
    },
  ],
})

router.beforeEach((to, from) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.guest && authStore.isAuthenticated) {
    return { name: 'dashboard' }
  }
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    return { name: 'dashboard' }
  }
})

export default router
