<script setup lang="ts">
import type { OperLog } from '#/api/monitor/operlog/model';

import { onMounted } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';

import { message, Modal, Space } from 'ant-design-vue';

import { useVbenVxeGrid, vxeCheckboxChecked } from '#/adapter/vxe-table';
import {
  operLogClean,
  operLogDelete,
  operLogExport,
  operLogList,
} from '#/api/monitor/operlog';
import { downloadBlob } from '#/utils/download';
import { useDictStore } from '#/store/dict';

import { columns, querySchema } from './data';
import OperlogDetailDrawer from './operlog-detail-drawer.vue';

const dictStore = useDictStore();

onMounted(() => {
  dictStore.getDictOptions('sys_oper_type');
  dictStore.getDictOptions('sys_oper_status');
});

const [DetailDrawerRef, detailDrawerApi] = useVbenDrawer({
  connectedComponent: OperlogDetailDrawer,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: querySchema,
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
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
    sortConfig: {
      remote: true,
      trigger: 'cell',
    },
    proxyConfig: {
      sort: true,
      ajax: {
        query: async ({ page, sorts }: any, formValues: Record<string, any> = {}) => {
          const sortParams: Record<string, string> = {};
          if (sorts && sorts.length > 0) {
            const sort = sorts[0];
            if (sort && sort.order) {
              sortParams.orderBy = sort.field;
              sortParams.orderDirection = sort.order;
            }
          }

          const params: Record<string, any> = {
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
            ...sortParams,
          };

          // Handle operTime date range
          if (params.operTime && Array.isArray(params.operTime)) {
            params.beginTime = params.operTime[0];
            params.endTime = params.operTime[1];
            delete params.operTime;
          }

          return await operLogList(params);
        },
      },
    },
    headerCellConfig: {
      height: 44,
    },
    cellConfig: {
      height: 48,
    },
    rowConfig: {
      keyField: 'id',
    },
    id: 'monitor-operlog-index',
  },
});

function handlePreview(row: OperLog) {
  detailDrawerApi.setData({ record: row });
  detailDrawerApi.open();
}

function handleClean() {
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: '确认要清空所有操作日志数据吗？',
    onOk: async () => {
      await operLogClean();
      message.success('清空成功');
      await gridApi.reload();
    },
  });
}

function handleDelete() {
  const rows = gridApi.grid.getCheckboxRecords();
  const ids = rows.map((row: OperLog) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条操作日志吗？`,
    onOk: async () => {
      await operLogDelete(ids);
      message.success('删除成功');
      await gridApi.query();
    },
  });
}

async function handleExport() {
  try {
    const formValues = gridApi.formApi.form.values;
    const params: Record<string, any> = { ...formValues };

    // Handle operTime date range
    if (params.operTime && Array.isArray(params.operTime)) {
      params.beginTime = params.operTime[0];
      params.endTime = params.operTime[1];
      delete params.operTime;
    }

    const data = await operLogExport(params);
    downloadBlob(data, '操作日志.xlsx');
    message.success('导出成功');
  } catch {
    message.error('导出失败');
  }
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="操作日志列表">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleClean">清 空</a-button>
          <a-button @click="handleExport">导 出</a-button>
          <a-button
            :disabled="!vxeCheckboxChecked(gridApi)"
            danger
            type="primary"
            @click="handleDelete"
          >
            删 除
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <ghost-button @click.stop="handlePreview(row)">详情</ghost-button>
      </template>
    </Grid>

    <DetailDrawerRef />
  </Page>
</template>
