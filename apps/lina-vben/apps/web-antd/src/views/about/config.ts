/**
 * 系统信息模块配置
 * 外链地址和项目信息集中在此定义，修改时无需改动页面组件代码
 */

/** 项目信息 */
export const PROJECT_INFO = {
  name: 'Lina',
  description:
    '一个基于 GoFrame + Vue 3 + Vben5 的现代化管理后台系统，采用前后端分离架构，提供高效、灵活的企业级管理解决方案。',
  version: 'v0.5.0',
  license: 'MIT',
  homepage: 'https://github.com/gqcn/lina',
};

/** 后端组件列表 */
export const BACKEND_COMPONENTS = [
  {
    name: 'GoFrame',
    version: 'v2.10.0',
    url: 'https://goframe.org',
    description: 'Go 语言 Web 框架',
  },
  {
    name: 'MySQL',
    version: '8.0+',
    url: 'https://www.mysql.com',
    description: '关系型数据库',
  },
  {
    name: 'JWT',
    version: '-',
    url: 'https://jwt.io',
    description: '身份认证',
  },
];

/** 前端组件列表 */
export const FRONTEND_COMPONENTS = [
  {
    name: 'Vue',
    version: '3.x',
    url: 'https://vuejs.org',
    description: '渐进式 JavaScript 框架',
  },
  {
    name: 'Vben Admin',
    version: '5.x',
    url: 'https://www.vben.pro',
    description: '中后台管理框架',
  },
  {
    name: 'Ant Design Vue',
    version: '4.x',
    url: 'https://antdv.com',
    description: 'UI 组件库',
  },
  {
    name: 'TypeScript',
    version: '5.x',
    url: 'https://www.typescriptlang.org',
    description: '类型安全的 JavaScript',
  },
  {
    name: 'Vite',
    version: '6.x',
    url: 'https://vite.dev',
    description: '前端构建工具',
  },
  {
    name: 'Pinia',
    version: '3.x',
    url: 'https://pinia.vuejs.org',
    description: '状态管理',
  },
  {
    name: 'TailwindCSS',
    version: '4.x',
    url: 'https://tailwindcss.com',
    description: '原子化 CSS 框架',
  },
];
