<script setup lang="ts">
import type { SystemPlugin } from '#/api/system/plugin/model';

import { Page } from '@vben/common-ui';

import { message, Popconfirm, Space, Switch, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  pluginDisable,
  pluginEnable,
  pluginInstall,
  pluginList,
  pluginSync,
  pluginUninstall,
} from '#/api/system/plugin';
import { notifyPluginRegistryChanged } from '#/plugins/slot-registry';

const typeColorMap: Record<string, string> = {
  runtime: 'green',
  source: 'blue',
};

const lifecycleStateLabelMap: Record<string, string> = {
  runtime_enabled: '运行时已启用',
  runtime_installed: '运行时已安装',
  runtime_uninstalled: '运行时未安装',
  source_disabled: '源码已禁用',
  source_enabled: '源码已启用',
};

const lifecycleStateColorMap: Record<string, string> = {
  runtime_enabled: 'success',
  runtime_installed: 'warning',
  runtime_uninstalled: 'default',
  source_disabled: 'warning',
  source_enabled: 'processing',
};

const nodeStateLabelMap: Record<string, string> = {
  enabled: '节点已启用',
  installed: '节点已接入',
  uninstalled: '节点未接入',
};

const nodeStateColorMap: Record<string, string> = {
  enabled: 'success',
  installed: 'processing',
  uninstalled: 'default',
};

const migrationStateLabelMap: Record<string, string> = {
  failed: '迁移失败',
  none: '无迁移',
  succeeded: '迁移成功',
};

const migrationStateColorMap: Record<string, string> = {
  failed: 'error',
  none: 'default',
  succeeded: 'success',
};

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: 'Input',
        fieldName: 'id',
        label: '插件标识',
      },
      {
        component: 'Input',
        fieldName: 'name',
        label: '插件名称',
      },
      {
        component: 'Select',
        fieldName: 'type',
        label: '插件类型',
        componentProps: {
          options: [
            { label: '源码插件', value: 'source' },
            { label: '运行时插件', value: 'runtime' },
          ],
        },
      },
      {
        component: 'Select',
        fieldName: 'installed',
        label: '接入态',
        componentProps: {
          options: [
            { label: '已接入', value: 1 },
            { label: '未安装', value: 0 },
          ],
        },
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: '状态',
        componentProps: {
          options: [
            { label: '启用', value: 1 },
            { label: '禁用', value: 0 },
          ],
        },
      },
    ],
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
  },
  gridOptions: {
    columns: [
      { field: 'id', minWidth: 160, title: '插件标识' },
      { field: 'name', minWidth: 160, title: '插件名称' },
      {
        field: 'type',
        slots: { default: 'type' },
        title: '插件类型',
        width: 120,
      },
      { field: 'version', title: '版本', width: 120 },
      {
        className: 'plugin-description-column',
        field: 'description',
        minWidth: 260,
        showOverflow: false,
        slots: { default: 'description' },
        title: '描述',
      },
      {
        field: 'enabled',
        slots: { default: 'enabled' },
        title: '状态',
        width: 130,
      },
      {
        field: 'lifecycleState',
        slots: { default: 'lifecycleState' },
        title: '生命周期',
        width: 150,
      },
      {
        field: 'governance',
        minWidth: 280,
        slots: { default: 'governance' },
        title: '治理摘要',
      },
      {
        field: 'action',
        fixed: 'right',
        slots: { default: 'action' },
        title: '操作',
        width: 180,
      },
      { field: 'installedAt', title: '安装时间', width: 180 },
      { field: 'updatedAt', title: '更新时间', width: 180 },
    ],
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    showOverflow: 'ellipsis',
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues = {},
        ) => {
          return await pluginList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    id: 'system-plugin-index',
  },
});

function getPluginTypeLabel(type: string) {
  return type === 'source' ? '源码插件' : '运行时插件';
}

function getPluginTypeColor(type: string) {
  return typeColorMap[type === 'source' ? 'source' : 'runtime'] || 'default';
}

function getLifecycleStateLabel(state: string) {
  return lifecycleStateLabelMap[state] || state || '-';
}

function getLifecycleStateColor(state: string) {
  return lifecycleStateColorMap[state] || 'default';
}

function getNodeStateLabel(state: string) {
  return nodeStateLabelMap[state] || state || '-';
}

function getNodeStateColor(state: string) {
  return nodeStateColorMap[state] || 'default';
}

function getMigrationStateLabel(state: string) {
  return migrationStateLabelMap[state] || state || '-';
}

function getMigrationStateColor(state: string) {
  return migrationStateColorMap[state] || 'default';
}

function isSourcePlugin(row: SystemPlugin) {
  return row.type === 'source';
}

async function handleStatusChange(row: SystemPlugin, checked: boolean) {
  if (row.installed !== 1) {
    message.warning('请先完成插件接入');
    return;
  }
  await (checked ? pluginEnable : pluginDisable)(row.id);
  row.enabled = checked ? 1 : 0;
  await notifyPluginRegistryChanged();
  message.success(checked ? '插件已启用' : '插件已禁用');
}

async function handleInstall(row: SystemPlugin) {
  await pluginInstall(row.id);
  row.installed = 1;
  row.enabled = 0;
  await notifyPluginRegistryChanged();
  message.success('运行时插件已安装');
  await gridApi.query();
}

async function handleUninstall(row: SystemPlugin) {
  await pluginUninstall(row.id);
  row.installed = 0;
  row.enabled = 0;
  await notifyPluginRegistryChanged();
  message.success('运行时插件已卸载');
  await gridApi.query();
}

async function handleSync() {
  const res = await pluginSync();
  await notifyPluginRegistryChanged();
  const total = typeof res?.total === 'number' ? res.total : 0;
  message.success(`已同步 ${total} 个源码插件`);
  await gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="插件列表">
      <template #toolbar-tools>
        <Space>
          <a-button type="primary" @click="handleSync">同步插件</a-button>
        </Space>
      </template>

      <template #type="{ row }">
        <Tag :color="getPluginTypeColor(row.type)">
          {{ getPluginTypeLabel(row.type) }}
        </Tag>
      </template>

      <template #description="{ row, isHidden }">
        <div
          v-if="!isHidden"
          :data-testid="`plugin-description-${row.id}`"
          class="max-w-full truncate"
          :title="row.description || '-'"
        >
          {{ row.description || '-' }}
        </div>
        <span v-else aria-hidden="true" class="sr-only"></span>
      </template>

      <template #enabled="{ row }">
        <Switch
          :checked="row.enabled === 1"
          :disabled="row.installed !== 1"
          checked-children="启用"
          un-checked-children="禁用"
          @change="(checked) => handleStatusChange(row, !!checked)"
        />
      </template>

      <template #lifecycleState="{ row }">
        <Tag :color="getLifecycleStateColor(row.lifecycleState)">
          {{ getLifecycleStateLabel(row.lifecycleState) }}
        </Tag>
      </template>

      <template #governance="{ row }">
        <Space wrap size="small">
          <Tag :color="getNodeStateColor(row.nodeState)">
            {{ getNodeStateLabel(row.nodeState) }}
          </Tag>
          <Tag :color="getMigrationStateColor(row.migrationState)">
            {{ getMigrationStateLabel(row.migrationState) }}
          </Tag>
          <Tag color="default">生效版本 {{ row.releaseVersion || '-' }}</Tag>
          <Tag color="default">资源 {{ row.resourceCount ?? 0 }}</Tag>
        </Space>
      </template>

      <template #action="{ row }">
        <Space v-if="!isSourcePlugin(row)">
          <Popconfirm
            v-if="row.installed !== 1"
            title="确认安装该插件？"
            @confirm="handleInstall(row)"
          >
            <ghost-button @click.stop="">安装</ghost-button>
          </Popconfirm>
          <Popconfirm
            v-else
            title="确认卸载该插件？"
            @confirm="handleUninstall(row)"
          >
            <ghost-button danger @click.stop="">卸载</ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>
  </Page>
</template>
