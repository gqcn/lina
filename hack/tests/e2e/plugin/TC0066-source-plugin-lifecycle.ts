import { execFileSync } from "node:child_process";
import type { APIRequestContext, APIResponse, Page } from "@playwright/test";

import { request as playwrightRequest } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { config } from "../../fixtures/config";
import { LoginPage } from "../../pages/LoginPage";
import { PluginPage } from "../../pages/PluginPage";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const pluginID = "plugin-demo";
const pluginAfterAuthHeader = "x-lina-plugin-after-auth";
const mysqlBin = process.env.E2E_MYSQL_BIN ?? "mysql";
const mysqlUser = process.env.E2E_DB_USER ?? "root";
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? "12345678";
const mysqlDatabase = process.env.E2E_DB_NAME ?? "lina";

type PluginListItem = {
  id: string;
  enabled?: number;
  installed?: number;
  status?: number;
};

type UserMenuNode = {
  name: string;
  type: string;
  children?: UserMenuNode[];
};

type UserRouteNode = {
  children?: UserRouteNode[];
  meta?: {
    title?: string;
  };
};

function unwrapApiData(payload: any) {
  if (payload && typeof payload === "object" && "data" in payload) {
    return payload.data;
  }
  return payload;
}

function assertOk(response: APIResponse, message: string) {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
}

async function createAdminApiContext(): Promise<APIRequestContext> {
  const loginApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await loginApi.post("auth/login", {
    data: {
      username: config.adminUser,
      password: config.adminPass,
    },
  });
  assertOk(loginResponse, "管理员登录 API 失败");

  const loginResult = unwrapApiData(await loginResponse.json());
  const accessToken = loginResult?.accessToken;
  expect(accessToken, "未获取到 accessToken").toBeTruthy();
  await loginApi.dispose();

  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

async function syncPlugins(adminApi: APIRequestContext) {
  const response = await adminApi.post("plugins/sync");
  assertOk(response, "同步源码插件失败");
}

async function listPlugins(
  adminApi: APIRequestContext,
): Promise<PluginListItem[]> {
  const response = await adminApi.get("plugins");
  assertOk(response, "查询插件列表失败");
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

async function fetchCurrentUserMenus(
  adminApi: APIRequestContext,
): Promise<UserMenuNode[]> {
  const response = await adminApi.get("user/info");
  assertOk(response, "查询当前用户信息失败");
  const payload = unwrapApiData(await response.json());
  return payload?.menus ?? [];
}

async function fetchCurrentUserRoutes(
  adminApi: APIRequestContext,
): Promise<UserRouteNode[]> {
  const response = await adminApi.get("menus/all");
  assertOk(response, "查询当前用户动态路由失败");
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

async function fetchPluginSummary(adminApi: APIRequestContext) {
  return await adminApi.get(`plugins/${pluginID}/summary`);
}

function hasMenuName(list: UserMenuNode[], name: string): boolean {
  return list.some((item) => {
    if (item.name === name) {
      return true;
    }
    return hasMenuName(item.children ?? [], name);
  });
}

function hasButtonMenuNode(list: UserMenuNode[]): boolean {
  return list.some((item) => {
    if (item.type === "B") {
      return true;
    }
    return hasButtonMenuNode(item.children ?? []);
  });
}

function hasRouteTitle(list: UserRouteNode[], title: string): boolean {
  return list.some((item) => {
    if (item?.meta?.title === title) {
      return true;
    }
    return hasRouteTitle(item?.children ?? [], title);
  });
}

async function findPlugin(adminApi: APIRequestContext, id = pluginID) {
  const list = await listPlugins(adminApi);
  return list.find((item) => item.id === id) ?? null;
}

async function updatePluginStatus(
  adminApi: APIRequestContext,
  id: string,
  enabled: boolean,
) {
  const url = enabled ? `plugins/${id}/enable` : `plugins/${id}/disable`;
  const response = await adminApi.put(url);
  assertOk(response, `更新插件状态失败: enabled=${enabled}`);
}

function resetPluginRegistryRow(id: string) {
  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-e",
      `DELETE FROM sys_plugin WHERE plugin_id = '${id.replaceAll("'", "''")}';`,
    ],
    {
      stdio: "ignore",
    },
  );
}

async function loginAsAdmin(page: Page) {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
}

test.describe("TC-66 源码插件生命周期", () => {
  let adminApi: APIRequestContext | null = null;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    if (adminApi) {
      await adminApi.dispose();
    }
  });

  test("TC-66a: 同步 source 插件后自动处于已集成且默认启用态", async ({
    page,
  }) => {
    resetPluginRegistryRow(pluginID);
    await syncPlugins(adminApi!);

    const pluginAfterSync = await findPlugin(adminApi!);
    expect(pluginAfterSync, "同步后应发现 plugin-demo").toBeTruthy();
    expect(pluginAfterSync?.installed, "源码插件同步后应直接处于已集成态").toBe(
      1,
    );
    expect(
      pluginAfterSync?.enabled ?? pluginAfterSync?.status,
      "源码插件首次同步后应默认启用",
    ).toBe(1);

    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await expect(loginPage.pluginLoginSlot).toBeVisible();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await expect(pluginPage.pluginRow(pluginID)).toBeVisible();
    await expect(pluginPage.pluginIntegratedTag(pluginID)).toBeVisible();
    await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
      "aria-checked",
      "true",
    );
    await expect(pluginPage.pluginInstallButton(pluginID)).toHaveCount(0);
    await expect(pluginPage.pluginUninstallButton(pluginID)).toHaveCount(0);
    await pluginPage.expectCrudSlotsVisible();
  });

  test("TC-66b: 启用后工作台卡片与左侧菜单页可正常展示", async ({ page }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, true);

    const pluginAfterEnable = await findPlugin(adminApi!);
    expect(pluginAfterEnable?.enabled ?? pluginAfterEnable?.status).toBe(1);

    const pluginPage = new PluginPage(page);
    await loginAsAdmin(page);

    await pluginPage.gotoWorkspace();
    await pluginPage.expectHeaderSlotsVisible();
    await pluginPage.expectWorkspaceSlotVisible();
    await pluginPage.openSidebarExampleFromMenu();
  });

  test("TC-66c: 启用后可验证插件路由与鉴权后回调", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, true);

    const summaryResponse = await fetchPluginSummary(adminApi!);
    assertOk(summaryResponse, "查询插件摘要路由失败");
    expect(
      summaryResponse.headers()[pluginAfterAuthHeader.toLowerCase()],
      "鉴权后回调应向受保护请求追加插件响应头",
    ).toBe(pluginID);
    const summaryPayload = unwrapApiData(await summaryResponse.json());
    expect(
      summaryPayload?.message,
      "插件摘要应仅返回页面实际使用的简介文案",
    ).toBe(
      "这是一条来自 plugin-demo 接口的简要介绍，用于验证插件页面可读取插件后端数据。",
    );

    await loginAsAdmin(page);
  });

  test("TC-66d: 禁用后不渲染插件 slot 且隐藏菜单", async ({ page }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, false);

    const summaryResponse = await fetchPluginSummary(adminApi!);
    expect(summaryResponse.status(), "插件禁用后插件自有路由应返回 404").toBe(
      404,
    );

    const pluginAfterDisable = await findPlugin(adminApi!);
    expect(pluginAfterDisable?.enabled ?? pluginAfterDisable?.status ?? 0).toBe(
      0,
    );

    const pluginPage = new PluginPage(page);
    await loginAsAdmin(page);

    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectSidebarMenuHidden("插件示例");
  });

  test("TC-66e: 禁用后源码插件仍保留已集成态且无需重新安装", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, false);

    const pluginAfterDisable = await findPlugin(adminApi!);
    expect(
      pluginAfterDisable,
      "禁用后仍应可在清单中发现 source 插件",
    ).toBeTruthy();
    expect(
      pluginAfterDisable?.installed ?? 0,
      "源码插件禁用后仍应保持已集成态",
    ).toBe(1);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await expect(pluginPage.pluginIntegratedTag(pluginID)).toBeVisible();
    await expect(pluginPage.pluginInstallButton(pluginID)).toHaveCount(0);
    await expect(pluginPage.pluginUninstallButton(pluginID)).toHaveCount(0);
  });

  test("TC-66f: 登录态在线启用后立即刷新左侧菜单与工作台卡片", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, false);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectSidebarMenuHidden("插件示例");

    await pluginPage.setPluginEnabled(pluginID, true);

    await pluginPage.expectSidebarMenuVisible("插件示例");
    await pluginPage.expectHeaderSlotsVisible();
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotVisible();
  });

  test("TC-66g: 登录态在线禁用后立即隐藏左侧菜单与工作台卡片", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, true);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotVisible();
    await pluginPage.gotoManage();
    await pluginPage.expectHeaderSlotsVisible();
    await pluginPage.expectSidebarMenuVisible("插件示例");

    await pluginPage.setPluginEnabled(pluginID, false);

    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectSidebarMenuHidden("插件示例");
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();
  });

  test("TC-66h: 当前会话重新获得焦点后自动同步外部插件状态变更", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, false);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectSidebarMenuHidden("插件示例");

    await updatePluginStatus(adminApi!, pluginID, true);
    await page.evaluate(() => {
      window.dispatchEvent(new Event("focus"));
      document.dispatchEvent(new Event("visibilitychange"));
    });

    await pluginPage.expectSidebarMenuVisible("插件示例");
    await pluginPage.expectHeaderSlotsVisible();
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotVisible();
  });

  test("TC-66i: 按钮权限不会被返回为左侧导航菜单或动态路由", async ({
    page,
  }) => {
    const currentUserMenus = await fetchCurrentUserMenus(adminApi!);
    expect(
      hasButtonMenuNode(currentUserMenus),
      "user/info 不应再返回按钮类型菜单",
    ).toBeFalsy();
    expect(
      hasMenuName(currentUserMenus, "插件查询"),
      "user/info 不应包含插件查询按钮菜单",
    ).toBeFalsy();

    const currentUserRoutes = await fetchCurrentUserRoutes(adminApi!);
    expect(
      hasRouteTitle(currentUserRoutes, "插件查询"),
      "menus/all 不应包含插件查询按钮路由",
    ).toBeFalsy();
    expect(
      hasRouteTitle(currentUserRoutes, "用户查询"),
      "menus/all 不应包含用户查询按钮路由",
    ).toBeFalsy();

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await expect(page).toHaveURL(/\/system\/plugin$/);
    await expect(page).toHaveTitle("插件管理 - Lina");
    await pluginPage.expectSidebarMenuHidden("插件查询");
    await pluginPage.expectSidebarMenuHidden("用户查询");
  });

  test("TC-66j: 当前会话重新获得焦点但插件状态未变化时不重复刷新菜单", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, true);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.expectSidebarMenuVisible("插件示例");

    const menuResponses: string[] = [];
    page.on("response", (response) => {
      if (
        response.request().method() === "GET" &&
        response.url().includes("/api/v1/menus/all")
      ) {
        menuResponses.push(response.url());
      }
    });

    await page.evaluate(() => {
      window.dispatchEvent(new Event("focus"));
      document.dispatchEvent(new Event("visibilitychange"));
    });

    await page.waitForTimeout(1200);
    await pluginPage.expectSidebarMenuVisible("插件示例");
    expect(
      menuResponses,
      "插件状态未变化时，焦点恢复不应重复拉取菜单",
    ).toHaveLength(0);
  });
});
