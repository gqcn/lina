import type { RouteRecordRaw, Router } from 'vue-router';

import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { resetStaticRoutes } from '@vben/utils';

import { useAuthStore } from '#/store';

import { generateAccess } from './access';
import { routes } from './routes';
import { accessRoutes } from './routes';

let accessRefreshTask: null | Promise<void> = null;

function collectAccessibleRouteNames(
  routeList: RouteRecordRaw[],
  names: Set<string> = new Set(),
) {
  for (const route of routeList) {
    if (typeof route.name === 'string' && route.name) {
      names.add(route.name);
    }
    if (route.children?.length) {
      collectAccessibleRouteNames(route.children, names);
    }
  }
  return names;
}

/**
 * Refreshes menus and dynamic routes for the current logged-in user.
 */
async function refreshAccessibleState(
  router: Router,
  { showLoadingToast = false }: { showLoadingToast?: boolean } = {},
) {
  if (accessRefreshTask) {
    return accessRefreshTask;
  }

  accessRefreshTask = (async () => {
    const accessStore = useAccessStore();
    const authStore = useAuthStore();
    const userStore = useUserStore();

    if (!accessStore.accessToken) {
      return;
    }

    const currentFullPath = router.currentRoute.value.fullPath;
    const userInfo = await authStore.fetchUserInfo();
    const userRoles = userStore.userInfo?.roles ?? [];

    resetStaticRoutes(router, routes);
    accessStore.setIsAccessChecked(false);

    const { accessibleMenus, accessibleRoutes } = await generateAccess(
      {
        roles: userRoles,
        router,
        routes: accessRoutes,
      },
      {
        showLoadingToast,
      },
    );

    accessStore.setAccessMenus(accessibleMenus);
    accessStore.setAccessRoutes(accessibleRoutes);
    accessStore.setIsAccessChecked(true);

    const accessibleNames = collectAccessibleRouteNames(accessibleRoutes);
    const resolved = router.resolve(currentFullPath);
    const hasAccessibleMatch = resolved.matched.some((route) => {
      return typeof route.name === 'string' && accessibleNames.has(route.name);
    });

    if (hasAccessibleMatch) {
      await router.replace(currentFullPath);
      return;
    }

    const fallbackPath =
      userInfo.homePath || preferences.app.defaultHomePath || '/';
    if (router.currentRoute.value.fullPath !== fallbackPath) {
      await router.replace(fallbackPath);
    }
  })().finally(() => {
    accessRefreshTask = null;
  });

  return accessRefreshTask;
}

export { refreshAccessibleState };
