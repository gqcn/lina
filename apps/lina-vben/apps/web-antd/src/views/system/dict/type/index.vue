<script setup lang="ts">
import type { DictType } from '#/api/system/dict/dict-type-model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message, Modal, Popconfirm, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  dictTypeDelete,
  dictTypeExport,
  dictTypeList,
} from '#/api/system/dict/dict-type';
import { downloadBlob } from '#/utils/download';

import { emitter } from '../mitt';
import { columns, querySchema } from './data';
import DictTypeImportModal from './dict-type-import-modal.vue';
import dictTypeModal from './dict-type-modal.vue';

const [DictTypeModal, modalApi] = useVbenModal({
  connectedComponent: dictTypeModal,
});

const [ImportModal, importModalApi] = useVbenModal({
  connectedComponent: DictTypeImportModal,
});

const lastDictType = ref('');

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
          return await dictTypeList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
      isCurrent: true,
    },
    id: 'system-dict-type-index',
    rowClassName: 'hover:cursor-pointer',
  },
  gridEvents: {
    cellClick: (e: any) => {
      const { row } = e;
      if (lastDictType.value === row.type) {
        return;
      }
      emitter.emit('rowClick', row.type);
      lastDictType.value = row.type;
    },
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

function handleAdd() {
  modalApi.setData({});
  modalApi.open();
}

function handleEdit(record: DictType) {
  modalApi.setData({ id: record.id });
  modalApi.open();
}

async function handleDelete(row: DictType) {
  await dictTypeDelete(row.id);
  message.success('删除成功');
  await tableApi.query();
}

function handleMultiDelete() {
  const rows = tableApi.grid.getCheckboxRecords();
  const ids = rows.map((row: DictType) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条记录吗？`,
    onOk: async () => {
      for (const id of ids) {
        await dictTypeDelete(id);
      }
      checkedRows.value = [];
      await tableApi.query();
    },
  });
}

function onReload() {
  tableApi.query();
}

function onImportReload() {
  tableApi.query();
}

async function handleExport() {
  try {
    const formValues = tableApi.formApi.form.values;
    const data = await dictTypeExport(formValues);
    downloadBlob(data, '字典类型.xlsx');
    message.success('导出成功');
  } catch {
    message.error('导出失败');
  }
}

function handleImport() {
  importModalApi.open();
}
</script>

<template>
  <div>
    <BasicTable id="dict-type" table-title="字典类型列表">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleExport">导 出</a-button>
          <a-button @click="handleImport">导 入</a-button>
          <a-button
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            删 除
          </a-button>
          <a-button type="primary" @click="handleAdd">新 增</a-button>
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
    <DictTypeModal @reload="onReload" />
    <ImportModal @reload="onImportReload" />
  </div>
</template>

<style lang="scss">
div#dict-type {
  .vxe-body--row {
    &.row--current {
      @apply font-semibold;
    }
  }
}
</style>
