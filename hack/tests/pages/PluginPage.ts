import { Page, Locator, expect } from "@playwright/test";

export class PluginPage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  get tableTitle(): Locator {
    return this.page.getByText("插件列表").first();
  }

  get runtimeUploadTrigger(): Locator {
    return this.page.getByTestId("plugin-runtime-upload-trigger").first();
  }

  get runtimeUploadDragger(): Locator {
    return this.page.getByTestId("plugin-runtime-upload-dragger").first();
  }

  get runtimeOverwriteSwitch(): Locator {
    return this.page.getByTestId("plugin-runtime-overwrite-switch").first();
  }

  get sidebarMenu(): Locator {
    return this.page.getByRole("menu").first();
  }

  sidebarMenuItem(menuName: string): Locator {
    return this.sidebarMenu.getByText(menuName, { exact: true }).first();
  }

  async clickSidebarMenuItem(menuName: string) {
    await this.expectSidebarMenuVisible(menuName);
    await this.sidebarMenuItem(menuName).click();
  }

  pluginIframeFrame() {
    return this.page.frameLocator("iframe");
  }

  pluginRuntimeEmbeddedHost(): Locator {
    return this.page.getByTestId("plugin-runtime-embedded-host").first();
  }

  pluginDemoRuntimeTitle(): Locator {
    return this.page
      .getByRole("heading", { name: "运行时插件示例已生效" })
      .first();
  }

  pluginDemoRuntimeDescription(): Locator {
    return this.page.getByText(
      "该页面来自 plugin-demo-runtime 的运行时挂载入口，用于验证宿主主内容区展示与独立静态页面跳转。",
    );
  }

  pluginDemoRuntimeOpenStandaloneButton(): Locator {
    return this.page.getByTestId("plugin-demo-runtime-open-standalone").first();
  }

  runtimeUploadDialog(): Locator {
    return this.page.getByRole("dialog", { name: "上传插件" }).last();
  }

  runtimeUploadTriggerLabel(): Locator {
    return this.runtimeUploadTrigger.getByText("上传插件", { exact: true });
  }

  runtimeUploadHint(): Locator {
    return this.runtimeUploadDialog().getByText(
      "仅支持单个 .wasm 文件，上传后可在列表中继续安装并启用。",
      { exact: true },
    );
  }

  runtimeOverwriteHint(): Locator {
    return this.runtimeUploadDialog().getByText(
      "允许覆盖同 ID 且未安装的插件工作区文件",
      { exact: true },
    );
  }

  runtimeUploadConfirmButton(): Locator {
    return this.runtimeUploadDialog()
      .getByRole("button", { name: /确\s*认|知\s*道了|知\s*道|ok/i })
      .last();
  }

  runtimeUploadCancelButton(): Locator {
    return this.runtimeUploadDialog()
      .getByRole("button", { name: /取\s*消|cancel/i })
      .last();
  }

  runtimeUploadCloseButton(): Locator {
    return this.runtimeUploadDialog()
      .locator(".ant-modal-close")
      .last();
  }

  uploadSuccessDialog(): Locator {
    return this.runtimeUploadDialog()
      .getByTestId("plugin-runtime-upload-success")
      .first();
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
      .getByRole("button", { name: /安\s*装/ })
      .first();
  }

  pluginUninstallButton(pluginId: string): Locator {
    return this.pluginRow(pluginId)
      .getByRole("button", { name: /卸\s*载/ })
      .first();
  }

  pluginSourceDisabledUninstallTrigger(pluginId: string): Locator {
    return this.page.getByTestId(
      `plugin-source-uninstall-disabled-${pluginId}`,
    );
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
        hasText:
          "这是一条来自 plugin-demo 接口的简要介绍，用于验证插件页面可读取插件后端数据。",
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

  async uploadRuntimePlugin(
    filePath: string,
    overwrite = false,
    expectedSuccessText?: string,
  ) {
    await this.runtimeUploadTrigger.click();
    await expect(this.runtimeUploadDialog()).toBeVisible();
    await expect(this.runtimeUploadDragger).toBeVisible();
    if (overwrite) {
      const isChecked =
        (await this.runtimeOverwriteSwitch.getAttribute("aria-checked")) ===
        "true";
      if (!isChecked) {
        await this.runtimeOverwriteSwitch.click();
      }
    }
    const [fileChooser] = await Promise.all([
      this.page.waitForEvent("filechooser"),
      this.runtimeUploadDragger.click(),
    ]);
    await fileChooser.setFiles(filePath);

    // Ant Design Upload updates the modal state asynchronously after the file
    // chooser closes. Waiting for the rendered upload item avoids clicking the
    // confirm button before the file is committed into the reactive file list.
    await expect(
      this.runtimeUploadDialog().locator(".ant-upload-list-item"),
    ).toBeVisible();
    await this.page.waitForTimeout(1500);

    const uploadResponsePromise = this.page.waitForResponse(
      (response) =>
        response.url().includes("/plugins/runtime/package") &&
        response.request().method() === "POST",
      { timeout: 30000 },
    );

    await this.runtimeUploadConfirmButton().click();

    const uploadResponse = await uploadResponsePromise;
    expect(uploadResponse.status()).toBe(200);

    await expect(this.uploadSuccessDialog()).toBeVisible();
    await expect(this.uploadSuccessDialog()).toContainText(
      expectedSuccessText ?? "上传成功，请在插件列表中继续安装并启用。",
    );
    await expect(this.runtimeUploadConfirmButton()).toContainText("知道了");
    await expect(this.runtimeUploadCancelButton()).toHaveCount(0);
    await expect(this.runtimeUploadCloseButton()).toHaveCount(0);
    await this.runtimeUploadConfirmButton().click();
    await expect(this.runtimeUploadDialog()).not.toBeVisible();

    // The Vite dev server keeps HMR-related requests alive, so waiting for
    // `networkidle` here can hang even after the upload flow already finished.
    // Use stable UI signals instead of transport-level idleness.
    await expect(this.runtimeUploadTrigger).toBeVisible();
    await expect(this.tableTitle).toBeVisible();
  }

  async installPlugin(pluginId: string) {
    const row = this.pluginRow(pluginId);
    await expect(row).toBeVisible();
    await this.pluginInstallButton(pluginId).click();
    const confirmPopover = this.page.locator(".ant-popover:visible").last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover
      .getByRole("button", { name: /确\s*定|确\s*认/i })
      .click();
    await expect(this.pluginUninstallButton(pluginId)).toBeVisible();
  }

  async uninstallPlugin(pluginId: string) {
    const row = this.pluginRow(pluginId);
    await expect(row).toBeVisible();
    await this.pluginUninstallButton(pluginId).click();
    const confirmPopover = this.page.locator(".ant-popover:visible").last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover
      .getByRole("button", { name: /确\s*定|确\s*认/i })
      .click();
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
    const menuItem = this.sidebarMenuItem(menuName);
    const visible = await menuItem.isVisible().catch(() => false);
    if (!visible) {
      await this.sidebarMenuItem("插件管理").click();
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
    const headerTitles = (
      await this.page
        .locator(".vxe-table--header .vxe-cell--title")
        .allTextContents()
    )
      .map((title) => title.trim())
      .filter(Boolean);

    const targetIndex = headerTitles.indexOf(targetTitle);
    const previousIndex = headerTitles.indexOf(previousTitle);
    const nextIndex = headerTitles.indexOf(nextTitle);

    expect(targetIndex, `未找到列表列: ${targetTitle}`).toBeGreaterThanOrEqual(
      0,
    );
    expect(
      previousIndex,
      `未找到列表列: ${previousTitle}`,
    ).toBeGreaterThanOrEqual(0);
    expect(nextIndex, `未找到列表列: ${nextTitle}`).toBeGreaterThanOrEqual(0);
    expect(
      targetIndex,
      `${targetTitle} 应位于 ${previousTitle} 之后`,
    ).toBeGreaterThan(previousIndex);
    expect(targetIndex, `${targetTitle} 应位于 ${nextTitle} 之前`).toBeLessThan(
      nextIndex,
    );
  }

  async expectDescriptionUsesNativeTooltip(pluginId: string) {
    const descriptionTestId = `plugin-description-${pluginId}`;
    const descriptionCell = this.pluginDescriptionCell(pluginId);
    const descriptionText =
      ((await descriptionCell.textContent()) || "").trim() || "-";
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

  async expectSourcePluginDisabledUninstall(pluginId: string) {
    const uninstallButton = this.pluginSourceDisabledUninstallTrigger(pluginId);
    const tooltipText =
      "源码插件不支持页面动态卸载，如需移除请在源码中取消注册后重新构建宿主。";

    const hasVisibleDisabledButton = await uninstallButton.evaluateAll(
      (elements, expectedTitle) => {
        return elements.some((element) => {
          if (!(element instanceof HTMLButtonElement)) {
            return false;
          }
          const style = window.getComputedStyle(element);
          const rect = element.getBoundingClientRect();
          const isVisible =
            style.display !== "none" &&
            style.visibility !== "hidden" &&
            rect.width > 0 &&
            rect.height > 0;
          return (
            isVisible &&
            element.disabled &&
            element.getAttribute("title") === expectedTitle
          );
        });
      },
      tooltipText,
    );

    expect(
      hasVisibleDisabledButton,
      "源码插件应显示一个可见的灰态卸载按钮，并携带动态卸载提示",
    ).toBeTruthy();
  }

  async openSidebarExampleFromMenu() {
    await this.clickSidebarMenuItem("插件示例");
    await expect(this.pluginSidebarSimpleTitle()).toBeVisible();
    await expect(this.pluginSidebarBriefDescription()).toBeVisible();
    await expect(this.pluginSidebarLegacyDescription()).toHaveCount(0);
    await expect(this.pluginSummaryMessage()).toHaveCount(0);
    await expect(this.pluginSummaryErrorToast()).toHaveCount(0);
  }
}
