<script setup lang="ts">
import type { DictData } from '#/api/system/dict/dict-data-model';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { message, Modal, Popconfirm, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  dictDataDelete,
  dictDataExport,
  dictDataList,
} from '#/api/system/dict/dict-data';
import { downloadBlob } from '#/utils/download';

import { emitter } from '../mitt';
import { columns, querySchema } from './data';
import dictDataDrawer from './dict-data-drawer.vue';

const dictType = ref('');

const [BasicTable, tableApi] = useVbenVxeGrid({
  formOptions: {
    schema: querySchema,
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
  },
  gridOptions: {
    checkboxConfig: {
      highlight: true,
      reserve: true,
    },
    columns,
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async ({ page }: { page: { currentPage: number; pageSize: number } }, formValues = {}) => {
          if (!dictType.value) {
            return { items: [], total: 0 };
          }
          return await dictDataList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            dictType: dictType.value,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    id: 'system-dict-data-index',
  },
  gridEvents: {
    checkboxChange: () => {
      checkedRows.value = tableApi.grid?.getCheckboxRecords() || [];
    },
    checkboxAll: () => {
      checkedRows.value = tableApi.grid?.getCheckboxRecords() || [];
    },
  },
});

const checkedRows = ref<any[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

const [DictDataDrawer, drawerApi] = useVbenDrawer({
  connectedComponent: dictDataDrawer,
});

function handleAdd() {
  drawerApi.setData({ dictType: dictType.value });
  drawerApi.open();
}

function handleEdit(row: DictData) {
  drawerApi.setData({ dictType: dictType.value, id: row.id });
  drawerApi.open();
}

async function handleDelete(row: DictData) {
  await dictDataDelete(row.id);
  message.success('删除成功');
  await tableApi.query();
}

function handleMultiDelete() {
  const rows = tableApi.grid.getCheckboxRecords();
  const ids = rows.map((row: DictData) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条记录吗？`,
    onOk: async () => {
      for (const id of ids) {
        await dictDataDelete(id);
      }
      checkedRows.value = [];
      await tableApi.query();
    },
  });
}

async function handleExport() {
  try {
    const data = await dictDataExport({
      dictType: dictType.value,
      ...tableApi.formApi.form.values,
    });
    downloadBlob(data, '字典数据.xlsx');
    message.success('导出成功');
  } catch {
    message.error('导出失败');
  }
}

function onReload() {
  tableApi.query();
}

emitter.on('rowClick', async (value: string) => {
  dictType.value = value;
  await tableApi.query();
});
</script>

<template>
  <div>
    <BasicTable id="dict-data" table-title="字典数据列表">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleExport">导 出</a-button>
          <a-button
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            删 除
          </a-button>
          <a-button
            :disabled="dictType === ''"
            type="primary"
            @click="handleAdd"
          >
            新 增
          </a-button>
        </Space>
      </template>
      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleEdit(row)">编辑</ghost-button>
          <Popconfirm
            placement="left"
            title="确认删除？"
            @confirm="handleDelete(row)"
          >
            <ghost-button danger @click.stop="">删除</ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </BasicTable>
    <DictDataDrawer @reload="onReload" />
  </div>
</template>
