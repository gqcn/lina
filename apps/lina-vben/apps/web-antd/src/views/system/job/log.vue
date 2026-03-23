<script setup lang="ts">
import { Page } from '@vben/common-ui';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { useVbenDrawer } from '@vben/common-ui';
import { jobApi, type JobApi } from '#/api/system/job';
import { useDictStore } from '#/store/dict';
import { useRoute } from 'vue-router';
import JobLogDetail from './log-detail.vue';

const route = useRoute();
const dictStore = useDictStore();

const [Drawer, drawerApi] = useVbenDrawer({
  title: '日志详情',
  connectedComponent: JobLogDetail,
});

const [Grid, gridApi] = useVbenVxeGrid({
  columns: [
    { field: 'jobName', title: '任务名称', width: 150 },
    { field: 'jobGroup', title: '任务分组', width: 120 },
    { field: 'command', title: '执行指令', width: 200 },
    { field: 'startTime', title: '开始时间', width: 160 },
    { field: 'endTime', title: '结束时间', width: 160 },
    {
      field: 'duration',
      title: '执行耗时',
      width: 120,
      formatter: ({ cellValue }) => (cellValue ? `${cellValue}ms` : '-'),
    },
    {
      field: 'status',
      title: '执行状态',
      width: 100,
      cellRender: { name: 'DictTag', props: { dictType: 'sys_job_log_status' } },
    },
    {
      field: 'action',
      title: '操作',
      width: 100,
      fixed: 'right',
      slots: { default: 'action' },
    },
  ],
  proxyConfig: {
    ajax: {
      query: async ({ page }) => {
        const { items, total } = await jobApi.logList({
          jobName: gridApi.getSearchFormValues().jobName,
          status: gridApi.getSearchFormValues().status,
          startTime: gridApi.getSearchFormValues().startTime,
          endTime: gridApi.getSearchFormValues().endTime,
          page: page.currentPage,
          pageSize: page.pageSize,
        });
        return { items, total };
      },
    },
  },
  formConfig: {
    items: [
      {
        field: 'jobName',
        title: '任务名称',
        itemRender: { name: 'AInput' },
        defaultValue: route.query.jobName || '',
      },
      {
        field: 'status',
        title: '执行状态',
        itemRender: {
          name: 'ASelect',
          props: { options: dictStore.getDictOptions('sys_job_log_status') },
        },
      },
      {
        field: 'startTime',
        title: '开始时间',
        itemRender: { name: 'ARangePicker' },
      },
    ],
  },
  toolbarConfig: {
    buttons: [{ code: 'reload', name: '刷新' }],
  },
});

function handleDetail(row: JobApi.JobLog) {
  drawerApi.open({ values: row });
}
</script>

<template>
  <Page>
    <Grid>
      <template #action="{ row }">
        <a-button type="link" size="small" @click="handleDetail(row)">详情</a-button>
      </template>
    </Grid>
    <Drawer />
  </Page>
</template>
