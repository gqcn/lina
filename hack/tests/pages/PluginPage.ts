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

  tableColumn(title: string): Locator {
    return this.page
      .locator(".vxe-table--header .vxe-cell--title", { hasText: title })
      .first();
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

  pluginEnabledSwitch(pluginId: string): Locator {
    return this.pluginRow(pluginId).locator(".ant-switch").first();
  }

  pluginDescriptionCell(pluginId: string): Locator {
    return this.pluginRow(pluginId)
      .getByTestId(`plugin-description-${pluginId}`)
      .first();
  }

  antTooltip(): Locator {
    return this.page.locator(".ant-tooltip:visible");
  }

  vxeTooltip(): Locator {
    return this.page.locator(".vxe-table--tooltip-wrapper:visible");
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

  pluginSidebarBriefDescription(): Locator {
    return this.page.getByText(
      "这是一条来自 plugin-demo 接口的简要介绍，用于验证插件页面可读取插件后端数据。",
    );
  }

  pluginSidebarLegacyDescription(): Locator {
    return this.page.getByText(
      "当前页面用于验证 plugin-demo 已成功接入宿主左侧菜单，并能在后台主内容区正常打开。",
    );
  }

  pluginSummaryMessage(): Locator {
    return this.page.getByText(
      "plugin-demo 仅演示最小源码插件接入，不包含数据库读写示例。",
    );
  }

  pluginSummaryErrorToast(): Locator {
    return this.page
      .locator(".ant-message-notice")
      .filter({
        hasText: "这是一条来自 plugin-demo 接口的简要介绍，用于验证插件页面可读取插件后端数据。",
      })
      .first();
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

  async expectTableColumnVisible(title: string) {
    await expect(this.tableColumn(title)).toBeVisible();
  }

  async expectTableColumnHidden(title: string) {
    await expect(this.tableColumn(title)).toHaveCount(0);
  }

  async expectTableColumnBetween(
    targetTitle: string,
    previousTitle: string,
    nextTitle: string,
  ) {
    const headerTitles = (await this.page
      .locator(".vxe-table--header .vxe-cell--title")
      .allTextContents())
      .map((title) => title.trim())
      .filter(Boolean);

    const targetIndex = headerTitles.indexOf(targetTitle);
    const previousIndex = headerTitles.indexOf(previousTitle);
    const nextIndex = headerTitles.indexOf(nextTitle);

    expect(targetIndex, `未找到列表列: ${targetTitle}`).toBeGreaterThanOrEqual(0);
    expect(previousIndex, `未找到列表列: ${previousTitle}`).toBeGreaterThanOrEqual(0);
    expect(nextIndex, `未找到列表列: ${nextTitle}`).toBeGreaterThanOrEqual(0);
    expect(targetIndex, `${targetTitle} 应位于 ${previousTitle} 之后`).toBeGreaterThan(
      previousIndex,
    );
    expect(targetIndex, `${targetTitle} 应位于 ${nextTitle} 之前`).toBeLessThan(
      nextIndex,
    );
  }

  async expectDescriptionUsesNativeTooltip(pluginId: string) {
    const descriptionTestId = `plugin-description-${pluginId}`;
    const descriptionCell = this.pluginDescriptionCell(pluginId);
    const descriptionText = ((await descriptionCell.textContent()) || "").trim() || "-";
    await expect(descriptionCell).toBeVisible();
    await expect(this.page.getByTestId(descriptionTestId)).toHaveCount(1);
    await expect(descriptionCell).toHaveAttribute("title", descriptionText);
    await descriptionCell.hover();
    await expect(this.vxeTooltip()).toHaveCount(0);
    await expect(this.antTooltip()).toHaveCount(0);
    await this.page.waitForTimeout(5000);
    await expect(this.vxeTooltip()).toHaveCount(0);
    await expect(this.antTooltip()).toHaveCount(0);
    const delayedTitleCount = await this.page
      .locator("[title]")
      .evaluateAll((elements, text) => {
        return elements.filter((element) =>
          (element.getAttribute("title") || "").includes(text),
        ).length;
      }, descriptionText);
    expect(delayedTitleCount, "描述列应只保留单一系统默认提示来源").toBe(1);
  }

  async openSidebarExampleFromMenu() {
    await this.expectSidebarMenuVisible("插件示例");
    await this.sidebarMenu
      .getByText("插件示例", { exact: true })
      .first()
      .click();
    await expect(this.pluginSidebarSimpleTitle()).toBeVisible();
    await expect(this.pluginSidebarBriefDescription()).toBeVisible();
    await expect(this.pluginSidebarLegacyDescription()).toHaveCount(0);
    await expect(this.pluginSummaryMessage()).toHaveCount(0);
    await expect(this.pluginSummaryErrorToast()).toHaveCount(0);
  }
}
