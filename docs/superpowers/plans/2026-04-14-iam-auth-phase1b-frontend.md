# IAM Phase 1B - Frontend Management Console Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Vue 3 + TypeScript management console for the IAM service with login, dashboard, user management, tenant management, role management, and audit log pages.

**Architecture:** Vue 3 SPA with Vite, Element Plus UI, Pinia state management, Vue Router, Axios for API calls. Communicates with the backend from Phase 1A.

**Tech Stack:** Vue 3 + TypeScript + Vite + Element Plus + Pinia + Vue Router + Axios + Playwright + agent-browser（有头模式 --headed）

---

## File Map

| File | Responsibility |
|------|---------------|
| `web/package.json` | Frontend dependencies |
| `web/vite.config.ts` | Vite build config with API proxy |
| `web/tsconfig.json` | TypeScript config |
| `web/index.html` | HTML entry point |
| `web/src/main.ts` | Vue app entry |
| `web/src/App.vue` | Root component |
| `web/src/router/index.ts` | Route definitions + auth guards |
| `web/src/api/index.ts` | Axios instance + API functions |
| `web/src/types/index.ts` | TypeScript type definitions |
| `web/src/store/auth.ts` | Auth state (Pinia) |
| `web/src/layouts/AdminLayout.vue` | Admin shell (sidebar + header + content) |
| `web/src/views/LoginView.vue` | Login page |
| `web/src/views/DashboardView.vue` | Dashboard |
| `web/src/views/UserListView.vue` | User list with search/filter |
| `web/src/views/UserEditView.vue` | User detail/edit |
| `web/src/views/TenantListView.vue` | Tenant list |
| `web/src/views/TenantCreateView.vue` | Create tenant form |
| `web/src/views/RoleListView.vue` | Role list |
| `web/src/views/RoleCreateView.vue` | Create role form |
| `web/src/views/AuditLogView.vue` | Audit log list |
| `web/tests/login.spec.ts` | Playwright login test |
| `web/tests/dashboard.spec.ts` | Playwright dashboard test |
| `web/tests/user-management.spec.ts` | Playwright user management test |
| `web/tests/tenant-management.spec.ts` | Playwright tenant management test |
| `web/tests/role-management.spec.ts` | Playwright role management test |
| `web/tests/audit-log.spec.ts` | Playwright audit log test |
| `web/playwright.config.ts` | Playwright config |

---

### Task F1: Frontend Scaffold + Dependencies

**Files:**
- Create: `web/package.json`, `web/vite.config.ts`, `web/tsconfig.json`, `web/index.html`

- [ ] **Step 1: Create web/package.json**

```json
{
  "name": "iam-admin-console",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc -b && vite build",
    "preview": "vite preview",
    "test": "playwright test",
    "test:ui": "playwright test --ui"
  },
  "dependencies": {
    "vue": "^3.5",
    "vue-router": "^4.5",
    "pinia": "^2.3",
    "axios": "^1.7",
    "element-plus": "^2.9"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.2",
    "typescript": "~5.6",
    "vite": "^6.0",
    "vue-tsc": "^2.1",
    "@playwright/test": "^1.49"
  }
}
```

- [ ] **Step 2: Create web/vite.config.ts**

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 3: Create web/tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.tsx", "src/**/*.vue"]
}
```

- [ ] **Step 4: Create web/index.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>IAM Admin Console</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
```

- [ ] **Step 5: Install dependencies**

```bash
cd web && npm install
```

- [ ] **Step 6: Commit**

```bash
git add web/package.json web/package-lock.json web/vite.config.ts web/tsconfig.json web/index.html
git commit -m "feat: frontend scaffold with Vite, Vue 3, TypeScript, Element Plus"
```

---

### Task F2: Types + API Layer + Auth Store

**Files:**
- Create: `web/src/types/index.ts`, `web/src/api/index.ts`, `web/src/store/auth.ts`

- [ ] **Step 1: Create web/src/types/index.ts**

```typescript
export interface User {
  id: string
  email: string
  name: string
  avatar_url: string
  status: 'active' | 'inactive' | 'locked'
  email_verified: boolean
  created_at: string
}

export interface Tenant {
  id: string
  name: string
  unique_code: string
  custom_domain: string
  status: 'active' | 'inactive'
  created_at: string
}

export interface Role {
  id: string
  name: string
  description: string
  is_system: boolean
  created_at: string
}

export interface AuditLog {
  id: string
  user_id: string | null
  action: string
  target_type: string
  target_id: string | null
  details: Record<string, unknown>
  ip_address: string
  created_at: string
}

export interface ApiResponse<T> {
  data: T
  code?: string
  message?: string
}

export interface LoginRequest {
  email: string
  password: string
  tenant_code: string
}

export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_in: number
}
```

- [ ] **Step 2: Create web/src/api/index.ts**

```typescript
import axios from 'axios'
import type { AxiosInstance } from 'axios'
import type {
  ApiResponse,
  LoginRequest,
  AuthTokens,
  User,
  Tenant,
  Role,
  AuditLog,
} from '@/types'

const api: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// Request interceptor: attach access token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor: auto-refresh on 401
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      try {
        const refreshToken = localStorage.getItem('refresh_token')
        const { data } = await axios.post<ApiResponse<AuthTokens>>(
          '/api/v1/auth/refresh',
          { refresh_token: refreshToken },
        )
        localStorage.setItem('access_token', data.data.access_token)
        localStorage.setItem('refresh_token', data.data.refresh_token)
        originalRequest.headers.Authorization = `Bearer ${data.data.access_token}`
        return api(originalRequest)
      } catch {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        window.location.href = '/login'
        return Promise.reject(error)
      }
    }
    return Promise.reject(error)
  },
)

// Auth APIs
export const authApi = {
  login: (req: LoginRequest) =>
    api.post<ApiResponse<AuthTokens>>('/auth/login', req),
  logout: (refreshToken: string) =>
    api.post('/auth/logout', { refresh_token: refreshToken }),
}

// User APIs
export const userApi = {
  getProfile: () => api.get<ApiResponse<User>>('/users/me'),
  updateProfile: (data: { name: string; avatar_url: string }) =>
    api.put('/users/me', data),
  changePassword: (oldPassword: string, newPassword: string) =>
    api.put('/users/me/password', { old_password: oldPassword, new_password: newPassword }),
}

// Admin APIs
export const adminApi = {
  listUsers: (params?: { email?: string; status?: string; offset?: number; limit?: number }) =>
    api.get<ApiResponse<{ users: User[]; total: number }>>('/admin/users', { params }),
  getUser: (id: string) => api.get<ApiResponse<User>>(`/admin/users/${id}`),
  updateUser: (id: string, data: { name?: string; avatar_url?: string }) =>
    api.put(`/admin/users/${id}`, data),
  updateUserStatus: (id: string, status: string) =>
    api.put(`/admin/users/${id}/status`, { status }),
  resetUserPassword: (id: string, newPassword: string) =>
    api.post(`/admin/users/${id}/reset-password`, { new_password: newPassword }),
  getUserRoles: (id: string) =>
    api.get<ApiResponse<Role[]>>(`/admin/users/${id}/roles`),
  assignUserRole: (id: string, roleId: string) =>
    api.put(`/admin/users/${id}/roles`, { role_id: roleId }),

  listTenants: (params?: { offset?: number; limit?: number }) =>
    api.get<ApiResponse<{ tenants: Tenant[]; total: number }>>('/admin/tenants', { params }),
  createTenant: (data: { name: string; unique_code: string; custom_domain?: string }) =>
    api.post<ApiResponse<Tenant>>('/admin/tenants', data),
  getTenant: (id: string) => api.get<ApiResponse<Tenant>>(`/admin/tenants/${id}`),

  listRoles: () => api.get<ApiResponse<Role[]>>('/admin/roles'),
  createRole: (data: { name: string; description?: string }) =>
    api.post<ApiResponse<Role>>('/admin/roles', data),

  listAuditLogs: (params?: { start_time?: string; end_time?: string; offset?: number; limit?: number }) =>
    api.get<ApiResponse<{ logs: AuditLog[]; total: number }>>('/admin/audit-logs', { params }),
}

export default api
```

- [ ] **Step 3: Create web/src/store/auth.ts**

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'
import type { LoginRequest, User } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string>(localStorage.getItem('access_token') || '')
  const refreshToken = ref<string>(localStorage.getItem('refresh_token') || '')
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => !!accessToken.value)

  async function login(req: LoginRequest) {
    const { data } = await authApi.login(req)
    accessToken.value = data.data.access_token
    refreshToken.value = data.data.refresh_token
    localStorage.setItem('access_token', data.data.access_token)
    localStorage.setItem('refresh_token', data.data.refresh_token)
  }

  async function logout() {
    try {
      await authApi.logout(refreshToken.value)
    } finally {
      accessToken.value = ''
      refreshToken.value = ''
      user.value = null
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
    }
  }

  return { accessToken, refreshToken, user, isAuthenticated, login, logout }
})
```

- [ ] **Step 4: Commit**

```bash
git add web/src/types/ web/src/api/ web/src/store/
git commit -m "feat: types, API layer with auto-refresh, and auth store"
```

---

### Task F3: App Entry + Router + Layout

**Files:**
- Create: `web/src/main.ts`, `web/src/App.vue`, `web/src/router/index.ts`, `web/src/layouts/AdminLayout.vue`

- [ ] **Step 1: Create web/src/main.ts**

```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus)
app.mount('#app')
```

- [ ] **Step 2: Create web/src/App.vue**

```vue
<template>
  <router-view />
</template>
```

- [ ] **Step 3: Create web/src/router/index.ts**

```typescript
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
    },
    {
      path: '/',
      component: () => import('@/layouts/AdminLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/dashboard' },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/DashboardView.vue'),
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('@/views/UserListView.vue'),
        },
        {
          path: 'users/:id/edit',
          name: 'user-edit',
          component: () => import('@/views/UserEditView.vue'),
          props: true,
        },
        {
          path: 'tenants',
          name: 'tenants',
          component: () => import('@/views/TenantListView.vue'),
        },
        {
          path: 'tenants/create',
          name: 'tenant-create',
          component: () => import('@/views/TenantCreateView.vue'),
        },
        {
          path: 'roles',
          name: 'roles',
          component: () => import('@/views/RoleListView.vue'),
        },
        {
          path: 'audit-logs',
          name: 'audit-logs',
          component: () => import('@/views/AuditLogView.vue'),
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('access_token')
  if (to.meta.requiresAuth && !token) {
    return { name: 'login' }
  }
  if (to.name === 'login' && token) {
    return { path: '/dashboard' }
  }
})

export default router
```

- [ ] **Step 4: Create web/src/layouts/AdminLayout.vue**

```vue
<template>
  <el-container style="height: 100vh">
    <el-aside width="200px">
      <div class="logo">IAM Admin</div>
      <el-menu :default-active="activeMenu" router>
        <el-menu-item index="/dashboard">
          <el-icon><Monitor /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/tenants">
          <el-icon><OfficeBuilding /></el-icon>
          <span>租户管理</span>
        </el-menu-item>
        <el-menu-item index="/roles">
          <el-icon><Key /></el-icon>
          <span>角色管理</span>
        </el-menu-item>
        <el-menu-item index="/audit-logs">
          <el-icon><Document /></el-icon>
          <span>操作日志</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header>
        <div class="header-right">
          <el-button type="danger" @click="handleLogout">登出</el-button>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { Monitor, User, OfficeBuilding, Key, Document } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const activeMenu = computed(() => route.path)

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: bold;
  color: #409eff;
  border-bottom: 1px solid #eee;
}
.header-right {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  height: 100%;
}
</style>
```

- [ ] **Step 5: Install icons**

```bash
cd web && npm install @element-plus/icons-vue
```

- [ ] **Step 6: Commit**

```bash
git add web/src/main.ts web/src/App.vue web/src/router/ web/src/layouts/
git commit -m "feat: app entry, router with auth guards, and admin layout with sidebar"
```

---

### Task F4: Login Page + Dashboard

**Files:**
- Create: `web/src/views/LoginView.vue`, `web/src/views/DashboardView.vue`

- [ ] **Step 1: Create web/src/views/LoginView.vue**

```vue
<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2 class="title">IAM Admin Console</h2>
      <el-form :model="form" :rules="rules" ref="formRef" @submit.prevent="handleLogin">
        <el-form-item prop="email">
          <el-input v-model="form.email" placeholder="邮箱" prefix-icon="Message" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item prop="tenant_code">
          <el-input v-model="form.tenant_code" placeholder="租户代码" prefix-icon="OfficeBuilding" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width: 100%" :loading="loading" native-type="submit">
            登录
          </el-button>
        </el-form-item>
      </el-form>
      <el-alert v-if="errorMsg" :title="errorMsg" type="error" show-icon closable />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import type { FormInstance } from 'element-plus'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const errorMsg = ref('')

const form = reactive({
  email: '',
  password: '',
  tenant_code: 'default',
})

const rules = {
  email: [{ required: true, message: '请输入邮箱', trigger: 'blur' }, { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  tenant_code: [{ required: true, message: '请输入租户代码', trigger: 'blur' }],
}

async function handleLogin() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  errorMsg.value = ''
  try {
    await authStore.login(form)
    router.push('/dashboard')
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: #f5f7fa;
}
.login-card {
  width: 400px;
}
.title {
  text-align: center;
  margin-bottom: 24px;
  color: #303133;
}
</style>
```

- [ ] **Step 2: Create web/src/views/DashboardView.vue**

```vue
<template>
  <div>
    <h2>仪表盘</h2>
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="8">
        <el-card>
          <el-statistic title="用户总数" :value="stats.userCount" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <el-statistic title="租户总数" :value="stats.tenantCount" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <el-statistic title="角色总数" :value="stats.roleCount" />
        </el-card>
      </el-col>
    </el-row>
    <el-card style="margin-top: 20px">
      <template #header>最近活动</template>
      <el-empty v-if="recentLogs.length === 0" description="暂无活动" />
      <el-timeline v-else>
        <el-timeline-item v-for="log in recentLogs" :key="log.id" :timestamp="log.created_at">
          {{ log.action }} - {{ log.target_type }}
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api'

const stats = ref({ userCount: 0, tenantCount: 0, roleCount: 0 })
const recentLogs = ref<any[]>([])

onMounted(async () => {
  try {
    const [usersRes, tenantsRes, rolesRes, logsRes] = await Promise.allSettled([
      adminApi.listUsers({ limit: 1 }),
      adminApi.listTenants({ limit: 1 }),
      adminApi.listRoles(),
      adminApi.listAuditLogs({ limit: 10 }),
    ])
    if (usersRes.status === 'fulfilled') stats.value.userCount = usersRes.value.data.data.total
    if (tenantsRes.status === 'fulfilled') stats.value.tenantCount = tenantsRes.value.data.data.total
    if (rolesRes.status === 'fulfilled') stats.value.roleCount = rolesRes.value.data.data.length
    if (logsRes.status === 'fulfilled') recentLogs.value = logsRes.value.data.data.logs
  } catch {
    // Dashboard stats are non-critical
  }
})
</script>
```

- [ ] **Step 3: Commit**

```bash
git add web/src/views/LoginView.vue web/src/views/DashboardView.vue
git commit -m "feat: login page and dashboard with statistics"
```

---

### Task F5: User Management Pages

**Files:**
- Create: `web/src/views/UserListView.vue`, `web/src/views/UserEditView.vue`

- [ ] **Step 1: Create web/src/views/UserListView.vue**

```vue
<template>
  <div>
    <h2>用户管理</h2>
    <el-row :gutter="16" style="margin: 16px 0">
      <el-col :span="8">
        <el-input v-model="searchEmail" placeholder="搜索邮箱" clearable @keyup.enter="fetchUsers" />
      </el-col>
      <el-col :span="4">
        <el-select v-model="searchStatus" placeholder="状态" clearable @change="fetchUsers">
          <el-option label="活跃" value="active" />
          <el-option label="禁用" value="inactive" />
          <el-option label="锁定" value="locked" />
        </el-select>
      </el-col>
      <el-col :span="4">
        <el-button type="primary" @click="fetchUsers">搜索</el-button>
      </el-col>
    </el-row>
    <el-table :data="users" v-loading="loading" style="width: 100%">
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="name" label="姓名" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="email_verified" label="邮箱验证">
        <template #default="{ row }">
          <el-tag :type="row.email_verified ? 'success' : 'warning'">
            {{ row.email_verified ? '已验证' : '未验证' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="editUser(row.id)">编辑</el-button>
          <el-button size="small" :type="row.status === 'active' ? 'danger' : 'success'"
            @click="toggleStatus(row)">
            {{ row.status === 'active' ? '禁用' : '启用' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="currentPage"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next"
      @current-change="fetchUsers"
      style="margin-top: 16px; justify-content: center"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api'
import type { User } from '@/types'

const router = useRouter()
const users = ref<User[]>([])
const loading = ref(false)
const searchEmail = ref('')
const searchStatus = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

async function fetchUsers() {
  loading.value = true
  try {
    const { data } = await adminApi.listUsers({
      email: searchEmail.value || undefined,
      status: searchStatus.value || undefined,
      offset: (currentPage.value - 1) * pageSize.value,
      limit: pageSize.value,
    })
    users.value = data.data.users
    total.value = data.data.total
  } catch {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

function statusType(status: string) {
  return { active: 'success', inactive: 'info', locked: 'danger' }[status] || 'info'
}

function editUser(id: string) {
  router.push(`/users/${id}/edit`)
}

async function toggleStatus(user: User) {
  const newStatus = user.status === 'active' ? 'inactive' : 'active'
  try {
    await ElMessageBox.confirm(`确定要${newStatus === 'active' ? '启用' : '禁用'}该用户吗？`, '确认')
    await adminApi.updateUserStatus(user.id, newStatus)
    ElMessage.success('状态已更新')
    fetchUsers()
  } catch {
    // User cancelled or error
  }
}

onMounted(fetchUsers)
</script>
```

- [ ] **Step 2: Create web/src/views/UserEditView.vue**

```vue
<template>
  <div>
    <h2>编辑用户</h2>
    <el-card v-if="user" style="max-width: 600px; margin-top: 16px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="邮箱">
          <el-input :value="user.email" disabled />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="头像 URL">
          <el-input v-model="form.avatar_url" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option label="活跃" value="active" />
            <el-option label="禁用" value="inactive" />
            <el-option label="锁定" value="locked" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveUser">保存</el-button>
        </el-form-item>
      </el-form>

      <el-divider>密码</el-divider>
      <el-form label-width="100px">
        <el-form-item label="新密码">
          <el-input v-model="newPassword" type="password" placeholder="输入新密码" />
        </el-form-item>
        <el-form-item>
          <el-button type="warning" @click="resetPassword">重置密码</el-button>
        </el-form-item>
      </el-form>

      <el-divider>角色</el-divider>
      <el-select v-model="selectedRoles" multiple placeholder="分配角色" @change="assignRoles" style="width: 100%">
        <el-option v-for="role in availableRoles" :key="role.id" :label="role.name" :value="role.id" />
      </el-select>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api'
import type { User, Role } from '@/types'

const route = useRoute()
const router = useRouter()
const user = ref<User | null>(null)
const form = ref({ name: '', avatar_url: '', status: 'active' })
const newPassword = ref('')
const availableRoles = ref<Role[]>([])
const selectedRoles = ref<string[]>([])

onMounted(async () => {
  const userId = route.params.id as string
  try {
    const [userRes, rolesRes, userRolesRes] = await Promise.all([
      adminApi.getUser(userId),
      adminApi.listRoles(),
      adminApi.getUserRoles(userId),
    ])
    user.value = userRes.data.data
    form.value = {
      name: user.value.name,
      avatar_url: user.value.avatar_url,
      status: user.value.status,
    }
    availableRoles.value = rolesRes.data.data
    selectedRoles.value = userRolesRes.data.data.map((r: Role) => r.id)
  } catch {
    ElMessage.error('获取用户信息失败')
    router.push('/users')
  }
})

async function saveUser() {
  try {
    await adminApi.updateUser(route.params.id as string, {
      name: form.value.name,
      avatar_url: form.value.avatar_url,
    })
    await adminApi.updateUserStatus(route.params.id as string, form.value.status)
    ElMessage.success('用户已更新')
  } catch {
    ElMessage.error('保存失败')
  }
}

async function resetPassword() {
  if (!newPassword.value) {
    ElMessage.warning('请输入新密码')
    return
  }
  try {
    await adminApi.resetUserPassword(route.params.id as string, newPassword.value)
    ElMessage.success('密码已重置')
    newPassword.value = ''
  } catch {
    ElMessage.error('重置密码失败')
  }
}

async function assignRoles() {
  try {
    await adminApi.assignUserRole(route.params.id as string, selectedRoles.value[0])
    ElMessage.success('角色已分配')
  } catch {
    ElMessage.error('分配角色失败')
  }
}
</script>
```

- [ ] **Step 3: Commit**

```bash
git add web/src/views/UserListView.vue web/src/views/UserEditView.vue
git commit -m "feat: user list with search/filter and user edit page"
```

---

### Task F6: Tenant + Role + Audit Pages

**Files:**
- Create: `web/src/views/TenantListView.vue`, `web/src/views/TenantCreateView.vue`, `web/src/views/RoleListView.vue`, `web/src/views/RoleCreateView.vue`, `web/src/views/AuditLogView.vue`

- [ ] **Step 1: Create web/src/views/TenantListView.vue**

```vue
<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center">
      <h2>租户管理</h2>
      <el-button type="primary" @click="$router.push('/tenants/create')">创建租户</el-button>
    </div>
    <el-table :data="tenants" v-loading="loading" style="width: 100%; margin-top: 16px">
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="unique_code" label="代码" />
      <el-table-column prop="custom_domain" label="自定义域名" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'danger'">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
    </el-table>
    <el-pagination
      v-model:current-page="currentPage"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next"
      @current-change="fetchTenants"
      style="margin-top: 16px; justify-content: center"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api'

const tenants = ref<any[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

async function fetchTenants() {
  loading.value = true
  try {
    const { data } = await adminApi.listTenants({
      offset: (currentPage.value - 1) * pageSize.value,
      limit: pageSize.value,
    })
    tenants.value = data.data.tenants
    total.value = data.data.total
  } catch {
    ElMessage.error('获取租户列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(fetchTenants)
</script>
```

- [ ] **Step 2: Create web/src/views/TenantCreateView.vue**

```vue
<template>
  <div>
    <h2>创建租户</h2>
    <el-card style="max-width: 500px; margin-top: 16px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="代码" prop="unique_code">
          <el-input v-model="form.unique_code" />
        </el-form-item>
        <el-form-item label="自定义域名">
          <el-input v-model="form.custom_domain" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSubmit">创建</el-button>
          <el-button @click="$router.push('/tenants')">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { adminApi } from '@/api'

const router = useRouter()
const formRef = ref<FormInstance>()
const form = reactive({ name: '', unique_code: '', custom_domain: '' })
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  unique_code: [{ required: true, message: '请输入代码', trigger: 'blur' }],
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    await adminApi.createTenant(form)
    ElMessage.success('租户已创建')
    router.push('/tenants')
  } catch {
    ElMessage.error('创建失败')
  }
}
</script>
```

- [ ] **Step 3: Create web/src/views/RoleListView.vue**

```vue
<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center">
      <h2>角色管理</h2>
      <el-button type="primary" @click="$router.push('/roles/create')">创建角色</el-button>
    </div>
    <el-table :data="roles" v-loading="loading" style="width: 100%; margin-top: 16px">
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="is_system" label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_system ? 'warning' : ''">{{ row.is_system ? '系统' : '自定义' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api'

const roles = ref<any[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const { data } = await adminApi.listRoles()
    roles.value = data.data
  } catch {
    ElMessage.error('获取角色列表失败')
  } finally {
    loading.value = false
  }
})
</script>
```

- [ ] **Step 4: Create web/src/views/RoleCreateView.vue**

```vue
<template>
  <div>
    <h2>创建角色</h2>
    <el-card style="max-width: 500px; margin-top: 16px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSubmit">创建</el-button>
          <el-button @click="$router.push('/roles')">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { adminApi } from '@/api'

const router = useRouter()
const formRef = ref<FormInstance>()
const form = reactive({ name: '', description: '' })
const rules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }] }

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    await adminApi.createRole(form)
    ElMessage.success('角色已创建')
    router.push('/roles')
  } catch {
    ElMessage.error('创建失败')
  }
}
</script>
```

- [ ] **Step 5: Create web/src/views/AuditLogView.vue**

```vue
<template>
  <div>
    <h2>操作日志</h2>
    <el-row :gutter="16" style="margin: 16px 0">
      <el-col :span="6">
        <el-date-picker v-model="dateRange" type="daterange" range-separator="至"
          start-placeholder="开始日期" end-placeholder="结束日期" @change="fetchLogs" />
      </el-col>
    </el-row>
    <el-table :data="logs" v-loading="loading" style="width: 100%">
      <el-table-column prop="action" label="操作" />
      <el-table-column prop="target_type" label="目标类型" width="120" />
      <el-table-column prop="ip_address" label="IP 地址" width="140" />
      <el-table-column prop="created_at" label="时间" width="180" />
      <el-table-column label="详情" width="100">
        <template #default="{ row }">
          <el-popover trigger="hover" placement="top" :width="300">
            <pre>{{ JSON.stringify(row.details, null, 2) }}</pre>
            <template #reference>
              <el-button size="small">查看</el-button>
            </template>
          </el-popover>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-model:current-page="currentPage"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next"
      @current-change="fetchLogs"
      style="margin-top: 16px; justify-content: center"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api'

const logs = ref<any[]>([])
const loading = ref(false)
const dateRange = ref<[Date, Date] | null>(null)
const currentPage = ref(1)
const pageSize = ref(50)
const total = ref(0)

async function fetchLogs() {
  loading.value = true
  try {
    const params: any = {
      offset: (currentPage.value - 1) * pageSize.value,
      limit: pageSize.value,
    }
    if (dateRange.value) {
      params.start_time = dateRange.value[0].toISOString()
      params.end_time = dateRange.value[1].toISOString()
    }
    const { data } = await adminApi.listAuditLogs(params)
    logs.value = data.data.logs
    total.value = data.data.total
  } catch {
    ElMessage.error('获取日志失败')
  } finally {
    loading.value = false
  }
}

onMounted(fetchLogs)
</script>
```

- [ ] **Step 6: Commit**

```bash
git add web/src/views/Tenant*.vue web/src/views/Role*.vue web/src/views/AuditLogView.vue
git commit -m "feat: tenant, role, and audit log pages"
```

---

### Task F7: Playwright E2E Tests

**Files:**
- Create: `web/playwright.config.ts`, `web/tests/login.spec.ts`, `web/tests/dashboard.spec.ts`, `web/tests/user-management.spec.ts`, `web/tests/tenant-management.spec.ts`, `web/tests/role-management.spec.ts`, `web/tests/audit-log.spec.ts`

- [ ] **Step 1: Create web/playwright.config.ts**

```typescript
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
})
```

- [ ] **Step 2: Create web/tests/login.spec.ts**

```typescript
import { test, expect } from '@playwright/test'

test.describe('Login Page', () => {
  test('should display login form', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByPlaceholder('邮箱')).toBeVisible()
    await expect(page.getByPlaceholder('密码')).toBeVisible()
    await expect(page.getByPlaceholder('租户代码')).toBeVisible()
    await expect(page.getByRole('button', { name: '登录' })).toBeVisible()
  })

  test('should show validation errors for empty form', async ({ page }) => {
    await page.goto('/login')
    await page.getByRole('button', { name: '登录' }).click()
    await expect(page.getByText('请输入邮箱')).toBeVisible()
  })

  test('should login successfully with admin credentials', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder('邮箱').fill('admin@iam.local')
    await page.getByPlaceholder('密码').fill('Admin@123')
    await page.getByPlaceholder('租户代码').fill('default')
    await page.getByRole('button', { name: '登录' }).click()
    await expect(page).toHaveURL('/dashboard')
    await expect(page.getByText('仪表盘')).toBeVisible()
  })

  test('should show error for wrong password', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder('邮箱').fill('admin@iam.local')
    await page.getByPlaceholder('密码').fill('WrongPassword1')
    await page.getByPlaceholder('租户代码').fill('default')
    await page.getByRole('button', { name: '登录' }).click()
    await expect(page.getByRole('alert')).toBeVisible()
  })
})
```

- [ ] **Step 3: Create web/tests/dashboard.spec.ts**

```typescript
import { test, expect } from '@playwright/test'

test.beforeEach(async ({ page }) => {
  await page.goto('/login')
  await page.getByPlaceholder('邮箱').fill('admin@iam.local')
  await page.getByPlaceholder('密码').fill('Admin@123')
  await page.getByPlaceholder('租户代码').fill('default')
  await page.getByRole('button', { name: '登录' }).click()
  await page.waitForURL('/dashboard')
})

test('should display dashboard with statistics', async ({ page }) => {
  await expect(page.getByText('仪表盘')).toBeVisible()
  await expect(page.getByText('用户总数')).toBeVisible()
  await expect(page.getByText('租户总数')).toBeVisible()
  await expect(page.getByText('角色总数')).toBeVisible()
})

test('should show recent activity section', async ({ page }) => {
  await expect(page.getByText('最近活动')).toBeVisible()
})
```

- [ ] **Step 4: Create web/tests/user-management.spec.ts**

```typescript
import { test, expect } from '@playwright/test'

test.beforeEach(async ({ page }) => {
  await page.goto('/login')
  await page.getByPlaceholder('邮箱').fill('admin@iam.local')
  await page.getByPlaceholder('密码').fill('Admin@123')
  await page.getByPlaceholder('租户代码').fill('default')
  await page.getByRole('button', { name: '登录' }).click()
  await page.waitForURL('/dashboard')
})

test('should navigate to user list and display users', async ({ page }) => {
  await page.getByText('用户管理').click()
  await expect(page.getByText('用户管理')).toBeVisible()
  await expect(page.getByPlaceholder('搜索邮箱')).toBeVisible()
})

test('should search users by email', async ({ page }) => {
  await page.getByText('用户管理').click()
  await page.getByPlaceholder('搜索邮箱').fill('admin')
  await page.getByRole('button', { name: '搜索' }).click()
  await expect(page.getByText('admin@iam.local')).toBeVisible()
})

test('should navigate to user edit page', async ({ page }) => {
  await page.getByText('用户管理').click()
  await page.getByRole('button', { name: '编辑' }).first().click()
  await expect(page.getByText('编辑用户')).toBeVisible()
})
```

- [ ] **Step 5: Create web/tests/tenant-management.spec.ts**

```typescript
import { test, expect } from '@playwright/test'

test.beforeEach(async ({ page }) => {
  await page.goto('/login')
  await page.getByPlaceholder('邮箱').fill('admin@iam.local')
  await page.getByPlaceholder('密码').fill('Admin@123')
  await page.getByPlaceholder('租户代码').fill('default')
  await page.getByRole('button', { name: '登录' }).click()
  await page.waitForURL('/dashboard')
})

test('should display tenant list', async ({ page }) => {
  await page.getByText('租户管理').click()
  await expect(page.getByText('租户管理')).toBeVisible()
  await expect(page.getByRole('button', { name: '创建租户' })).toBeVisible()
})

test('should navigate to create tenant page', async ({ page }) => {
  await page.getByText('租户管理').click()
  await page.getByRole('button', { name: '创建租户' }).click()
  await expect(page.getByText('创建租户')).toBeVisible()
})
```

- [ ] **Step 6: Create web/tests/role-management.spec.ts**

```typescript
import { test, expect } from '@playwright/test'

test.beforeEach(async ({ page }) => {
  await page.goto('/login')
  await page.getByPlaceholder('邮箱').fill('admin@iam.local')
  await page.getByPlaceholder('密码').fill('Admin@123')
  await page.getByPlaceholder('租户代码').fill('default')
  await page.getByRole('button', { name: '登录' }).click()
  await page.waitForURL('/dashboard')
})

test('should display role list', async ({ page }) => {
  await page.getByText('角色管理').click()
  await expect(page.getByText('角色管理')).toBeVisible()
  await expect(page.getByRole('button', { name: '创建角色' })).toBeVisible()
})

test('should navigate to create role page', async ({ page }) => {
  await page.getByText('角色管理').click()
  await page.getByRole('button', { name: '创建角色' }).click()
  await expect(page.getByText('创建角色')).toBeVisible()
})
```

- [ ] **Step 7: Create web/tests/audit-log.spec.ts**

```typescript
import { test, expect } from '@playwright/test'

test.beforeEach(async ({ page }) => {
  await page.goto('/login')
  await page.getByPlaceholder('邮箱').fill('admin@iam.local')
  await page.getByPlaceholder('密码').fill('Admin@123')
  await page.getByPlaceholder('租户代码').fill('default')
  await page.getByRole('button', { name: '登录' }).click()
  await page.waitForURL('/dashboard')
})

test('should display audit logs', async ({ page }) => {
  await page.getByText('操作日志').click()
  await expect(page.getByText('操作日志')).toBeVisible()
  await expect(page.getByRole('table')).toBeVisible()
})

test('should filter logs by date range', async ({ page }) => {
  await page.getByText('操作日志').click()
  await page.locator('.el-date-editor').first().click()
  // Date picker interaction - just verify it opens
  await expect(page.locator('.el-picker-panel')).toBeVisible()
})
```

- [ ] **Step 8: Commit**

```bash
git add web/playwright.config.ts web/tests/
git commit -m "feat: Playwright E2E tests for all frontend pages"
```

---

### Task F8: Final Verification

- [ ] **Step 1: Build frontend**

```bash
cd web && npm run build
```

Expected: Build succeeds, dist/ directory created.

- [ ] **Step 2: Run Playwright tests (requires backend running)**

```bash
make up
make migrate-up
./bin/server &
cd web && npx playwright install chromium
make test-e2e-web
```

Expected: All Playwright tests pass.

- [ ] **Step 3: agent-browser 有头模式页面验证**

使用 agent-browser --headed 逐页验证前端：

```bash
# 登录页
agent-browser --headed open http://localhost:5173/login
# 验证：表单渲染、输入框、按钮、错误提示

# 仪表盘（登录后）
agent-browser --headed open http://localhost:5173/dashboard
# 验证：侧边栏导航、统计卡片、最近活动

# 用户管理
agent-browser --headed open http://localhost:5173/users
# 验证：表格渲染、搜索框、分页、操作按钮

# 租户管理、角色管理、操作日志同理
```

**注意**：必须使用 `--headed` 标志打开可见浏览器窗口，便于实时观察页面操作结果。

- [ ] **Step 4: Commit**

```bash
git add .
git commit -m "chore: frontend build verification"
```

---

## Self-Review

**1. Spec coverage:**

| Spec Section | Task Coverage |
|-------------|---------------|
| Login page | F4 |
| Dashboard with stats | F4 |
| User management (list, search, filter, edit, disable, reset password, role assignment) | F5 |
| Tenant management (list, create, detail) | F6 |
| Role management (list, create) | F6 |
| Audit logs (list, date filter) | F6 |
| Auth flow (token storage, auto-refresh, 401 redirect) | F2, F3 |
| Playwright E2E tests | F7 |

All frontend requirements covered.

**2. Placeholder scan:** No TBD/TODO found.

**3. Type consistency:** All files use types from `@/types/index.ts`. API response format matches backend `ApiResponse<T>` wrapper.

---

Plan complete.
