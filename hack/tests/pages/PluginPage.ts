import { Page, Locator, expect } from "@playwright/test";

export class PluginPage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  get tableTitle(): Locator {
    return this.page.getByText("插件列表").first();
  }

  get sidebarMenu(): Locator {
    return this.page.getByRole("menu").first();
  }

  pluginRow(pluginId: string): Locator {
    return this.page.locator(".vxe-body--row", { hasText: pluginId }).first();
  }

  pluginInstallButton(pluginId: string): Locator {
    return this.pluginRow(pluginId)
      .getByText(/安\s*装/)
      .first();
  }

  pluginUninstallButton(pluginId: string): Locator {
    return this.pluginRow(pluginId)
      .getByText(/卸\s*载/)
      .first();
  }

  pluginIntegratedTag(pluginId: string): Locator {
    return this.pluginRow(pluginId).getByText("已集成").first();
  }

  pluginEnabledSwitch(pluginId: string): Locator {
    return this.pluginRow(pluginId).locator(".ant-switch").first();
  }

  headerActionBeforeSlot(): Locator {
    return this.page.getByText("plugin-demo 头部前置扩展").first();
  }

  headerActionAfterSlot(): Locator {
    return this.page.getByText("plugin-demo 头部后置扩展").first();
  }

  pluginSidebarSimpleTitle(): Locator {
    return this.page.getByRole("heading", { name: "插件示例已生效" }).first();
  }

  pluginSidebarSimpleDescription(): Locator {
    return this.page.getByText(
      "当前页面来自 plugin-demo 的左侧菜单入口，用于验证源码插件可以向宿主左侧导航插入页面，并在后台主内容区正常打开。",
    );
  }

  pluginSummaryMessage(): Locator {
    return this.page.getByText(
      "plugin-demo 仅演示最小源码插件接入，不包含数据库读写示例。",
    );
  }

  workspaceBeforeSlot(): Locator {
    return this.page.getByText(
      "plugin-demo 正在通过 `dashboard.workspace.before` 在工作台顶部插入横幅内容。",
    );
  }

  workspaceAfterSlot(): Locator {
    return this.page.getByText("插件示例工作台卡片").first();
  }

  crudToolbarSlot(): Locator {
    return this.page.getByText("plugin-demo CRUD 扩展").first();
  }

  async gotoManage() {
    await this.page.goto("/system/plugin");
    await expect(this.tableTitle).toBeVisible();
  }

  async syncPlugins() {
    await this.page.getByRole("button", { name: "同步插件" }).click();
    await this.page.waitForLoadState("networkidle");
  }

  async installPlugin(pluginId: string) {
    const row = this.pluginRow(pluginId);
    await expect(row).toBeVisible();
    await this.pluginInstallButton(pluginId).click();
    await this.page.getByRole("button", { name: /确\s*定|确\s*认/i }).click();
    await expect(this.pluginUninstallButton(pluginId)).toBeVisible();
  }

  async uninstallPlugin(pluginId: string) {
    const row = this.pluginRow(pluginId);
    await expect(row).toBeVisible();
    await this.pluginUninstallButton(pluginId).click();
    await this.page.getByRole("button", { name: /确\s*定|确\s*认/i }).click();
    await expect(this.pluginInstallButton(pluginId)).toBeVisible();
  }

  async setPluginEnabled(pluginId: string, enabled: boolean) {
    const row = this.pluginRow(pluginId);
    await expect(row).toBeVisible();
    const switcher = row.locator(".ant-switch").first();
    const isChecked = (await switcher.getAttribute("aria-checked")) === "true";
    if (isChecked !== enabled) {
      await switcher.click();
      await expect(switcher).toHaveAttribute(
        "aria-checked",
        enabled ? "true" : "false",
      );
      await expect(
        this.page.getByText(enabled ? "插件已启用" : "插件已禁用").last(),
      ).toBeVisible();
    }
  }

  async expectSidebarMenuVisible(menuName: string) {
    const menuItem = this.sidebarMenu
      .getByText(menuName, { exact: true })
      .first();
    const visible = await menuItem.isVisible().catch(() => false);
    if (!visible) {
      await this.sidebarMenu
        .getByText("插件管理", { exact: true })
        .first()
        .click();
    }
    await expect(menuItem).toBeVisible();
  }

  async expectSidebarMenuHidden(menuName: string) {
    const visible = await this.sidebarMenu
      .getByText(menuName, { exact: true })
      .first()
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    expect(visible).toBeFalsy();
  }

  async gotoWorkspace() {
    await this.page.goto("/dashboard/workspace");
    await expect(
      this.page.getByText("开始您一天的工作吧！").first(),
    ).toBeVisible();
  }

  async expectWorkspaceSlotVisible() {
    await expect(this.workspaceBeforeSlot()).toBeVisible();
    await expect(this.workspaceAfterSlot()).toBeVisible();
  }

  async expectWorkspaceSlotHidden() {
    await expect(this.workspaceBeforeSlot()).toHaveCount(0);
    await expect(this.workspaceAfterSlot()).toHaveCount(0);
  }

  async expectHeaderSlotsVisible() {
    await expect(this.headerActionBeforeSlot()).toBeVisible();
    await expect(this.headerActionAfterSlot()).toBeVisible();
  }

  async expectHeaderSlotsHidden() {
    await expect(this.headerActionBeforeSlot()).toHaveCount(0);
    await expect(this.headerActionAfterSlot()).toHaveCount(0);
  }

  async expectCrudSlotsVisible() {
    await expect(this.crudToolbarSlot()).toBeVisible();
  }

  async expectCrudSlotsHidden() {
    await expect(this.crudToolbarSlot()).toHaveCount(0);
  }

  async openSidebarExampleFromMenu() {
    await this.expectSidebarMenuVisible("插件示例");
    await this.sidebarMenu
      .getByText("插件示例", { exact: true })
      .first()
      .click();
    await expect(this.pluginSidebarSimpleTitle()).toBeVisible();
    await expect(this.pluginSidebarSimpleDescription()).toBeVisible();
    await expect(this.pluginSummaryMessage()).toBeVisible();
  }
}
