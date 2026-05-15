import type { Router, LocationQueryRaw } from 'vue-router';
import NProgress from 'nprogress'; // progress bar

import { useUserStore } from '@/store';
import { isLogin } from '@/utils/auth';
import type { RoleType } from '@/store/modules/user/types';

export default function setupUserLoginInfoGuard(router: Router) {
  router.beforeEach(async (to, from, next) => {
    NProgress.start();
    const userStore = useUserStore();
    if (isLogin()) {
      if (userStore.role) {
        next();
      } else {
        const cachedRole = window.localStorage.getItem('userRole');
        const isValidRole = (value: string): value is RoleType => {
          return value === 'admin' || value === 'user' || value === '*' || value === '';
        };
        if (cachedRole && isValidRole(cachedRole)) {
          userStore.setInfo({
            role: cachedRole,
            name: userStore.name || '分析师',
          });
          next();
          return;
        }
        try {
          await userStore.info();
          next();
        } catch (error) {
          await userStore.logout();
          next({
            name: 'login',
            query: {
              redirect: to.name,
              ...to.query,
            } as LocationQueryRaw,
          });
        }
      }
    } else {
      if (to.name === 'login') {
        next();
        return;
      }
      next({
        name: 'login',
        query: {
          redirect: to.name,
          ...to.query,
        } as LocationQueryRaw,
      });
    }
  });
}
