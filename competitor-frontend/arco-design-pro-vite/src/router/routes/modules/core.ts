import DashboardWorkplace from '@/views/dashboard/workplace/index.vue';
import Competitors from '@/views/list/search-table/index.vue';
import AnalysisNew from '@/views/form/group/index.vue';
import AnalysisReport from '@/views/visualization/data-analysis/index.vue';

import { DEFAULT_LAYOUT } from '../base';
import type { AppRouteRecordRaw } from '../types';

const routes: AppRouteRecordRaw[] = [
  {
    path: '/dashboard',
    name: 'dashboard',
    component: DEFAULT_LAYOUT,
    redirect: '/dashboard/workbench',
    meta: {
      locale: 'menu.workbench',
      requiresAuth: true,
      icon: 'icon-dashboard',
      order: 0,
      hideChildrenInMenu: true,
      roles: ['*'],
    },
    children: [
      {
        path: 'workbench',
        name: 'workbench',
        component: DashboardWorkplace,
        meta: {
          locale: 'menu.workbench',
          requiresAuth: true,
          hideInMenu: true,
          roles: ['*'],
        },
      },
    ],
  },
  {
    path: '/competitors',
    name: 'competitors',
    component: DEFAULT_LAYOUT,
    redirect: '/competitors/list',
    meta: {
      locale: 'menu.competitors',
      requiresAuth: true,
      icon: 'icon-list',
      order: 1,
      hideChildrenInMenu: true,
      hideInMenu: true,
      roles: ['*'],
    },
    children: [
      {
        path: 'list',
        name: 'competitorsList',
        component: Competitors,
        meta: {
          locale: 'menu.competitors',
          requiresAuth: true,
          hideInMenu: true,
          roles: ['*'],
        },
      },
    ],
  },
  {
    path: '/analysis',
    name: 'analysis',
    component: DEFAULT_LAYOUT,
    redirect: '/analysis/new',
    meta: {
      locale: 'menu.analysis.new',
      requiresAuth: true,
      icon: 'icon-settings',
      order: 2,
      hideChildrenInMenu: true,
      hideInMenu: true,
      roles: ['*'],
    },
    children: [
      {
        path: 'new',
        name: 'analysisNew',
        component: AnalysisNew,
        meta: {
          locale: 'menu.analysis.new',
          requiresAuth: true,
          hideInMenu: true,
          roles: ['*'],
        },
      },
    ],
  },
  {
    path: '/analysis-report',
    name: 'analysisReport',
    component: DEFAULT_LAYOUT,
    redirect: '/analysis-report/view',
    meta: {
      locale: 'menu.analysis.report',
      requiresAuth: true,
      icon: 'icon-file',
      order: 3,
      hideChildrenInMenu: true,
      hideInMenu: true,
      roles: ['*'],
    },
    children: [
      {
        path: 'view',
        name: 'analysisReportView',
        component: AnalysisReport,
        meta: {
          locale: 'menu.analysis.report',
          requiresAuth: true,
          hideInMenu: true,
          roles: ['*'],
        },
      },
    ],
  },
];

export default routes;
