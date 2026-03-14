import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:settings',
      order: 10,
      title: '系统管理',
    },
    name: 'System',
    path: '/system',
    children: [
      {
        name: 'UserManagement',
        path: '/system/user',
        component: () => import('#/views/system/user/index.vue'),
        meta: {
          icon: 'lucide:users',
          title: '用户管理',
        },
      },
      {
        name: 'DictManagement',
        path: '/system/dict',
        component: () => import('#/views/system/dict/index.vue'),
        meta: {
          icon: 'lucide:book-open',
          title: '字典管理',
        },
      },
      {
        name: 'DeptManagement',
        path: '/system/dept',
        component: () => import('#/views/system/dept/index.vue'),
        meta: {
          icon: 'lucide:network',
          title: '部门管理',
        },
      },
      {
        name: 'PostManagement',
        path: '/system/post',
        component: () => import('#/views/system/post/index.vue'),
        meta: {
          icon: 'lucide:briefcase',
          title: '岗位管理',
        },
      },
    ],
  },
  {
    name: 'Profile',
    path: '/profile',
    component: () => import('#/views/_core/profile/index.vue'),
    meta: {
      icon: 'lucide:user',
      hideInMenu: true,
      title: $t('page.auth.profile'),
    },
  },
];

export default routes;
