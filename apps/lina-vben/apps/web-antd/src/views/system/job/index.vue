<script setup lang="ts">
import { Page } from '@vben/common-ui';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { useVbenModal } from '@vben/common-ui';
import { jobApi, type JobApi } from '#/api/system/job';
import { useDictStore } from '#/store/dict';
import { message, Popconfirm } from 'ant-design-vue';
import JobForm from './form.vue';

const dictStore = useDictStore();

const [Modal, modalApi] = useVbenModal({
  connectedComponent: JobForm,
  onConfirm: async (values) => {
    if (values.id) {
      await jobApi.update(values);
      message.success('更新成功');
    } else {
      await jobApi.create(values);
      message.success('创建成功');
    }
    gridApi.reload();
  },
});

const [Grid, gridApi] = useVbenVxeGrid({
  columns: [
    { field: 'name', title: '任务名称', width: 150 },
    { field: 'group', title: '任务分组', width: 120 },
    { field: 'command', title: '执行指令', width: 200 },
    { field: 'cronExpr', title: 'Cron表达式', width: 150 },
    {
      field: 'status',
      title: '状态',
      width: 100,
      cellRender: { name: 'DictTag', props: { dictType: 'sys_job_status' } },
    },
    {
      field: 'singleton',
      title: '执行模式',
      width: 120,
      cellRender: { name: 'DictTag', props: { dictType: 'sys_job_singleton' } },
    },
    {
      field: 'execTimes',
      title: '执行次数',
      width: 120,
      formatter: ({ row }) => `${row.execTimes}/${row.maxTimes || '∞'}`,
    },
    {
      field: 'action',
      title: '操作',
      width: 280,
      fixed: 'right',
      slots: { default: 'action' },
    },
  ],
  proxyConfig: {
    ajax: {
      query: async ({ page }) => {
        const { items, total } = await jobApi.list({
          name: gridApi.getSearchFormValues().name,
          group: gridApi.getSearchFormValues().group,
          status: gridApi.getSearchFormValues().status,
          page: page.currentPage,
          pageSize: page.pageSize,
        });
        return { items, total };
      },
    },
  },
  formConfig: {
    items: [
      { field: 'name', title: '任务名称', itemRender: { name: 'AInput' } },
      { field: 'group', title: '任务分组', itemRender: { name: 'AInput' } },
      {
        field: 'status',
        title: '状态',
        itemRender: {
          name: 'ASelect',
          props: { options: dictStore.getDictOptions('sys_job_status') },
        },
      },
    ],
  },
  toolbarConfig: {
    buttons: [
      { code: 'insert', name: '新增', status: 'primary' },
      { code: 'reload', name: '刷新' },
    ],
  },
});

function handleAdd() {
  modalApi.open({ title: '新增任务' });
}

function handleEdit(row: JobApi.Job) {
  modalApi.open({ title: '编辑任务', values: row });
}

async function handleDelete(row: JobApi.Job) {
  await jobApi.delete([row.id]);
  message.success('删除成功');
  gridApi.reload();
}

async function handleToggleStatus(row: JobApi.Job) {
  const newStatus = row.status === 1 ? 0 : 1;
  await jobApi.updateStatus(row.id, newStatus);
  message.success(newStatus === 1 ? '已启用' : '已禁用');
  gridApi.reload();
}

async function handleRun(row: JobApi.Job) {
  await jobApi.run(row.id);
  message.success('任务已触发执行');
}

gridApi.on('toolbar-button-click', ({ code }) => {
  if (code === 'insert') handleAdd();
});
</script>

<template>
  <Page>
    <Grid>
      <template #action="{ row }">
        <a-button type="link" size="small" @click="handleEdit(row)">编辑</a-button>
        <Popconfirm
          v-if="row.isSystem === 0"
          title="确认删除？"
          @confirm="handleDelete(row)"
        >
          <a-button type="link" size="small" danger>删除</a-button>
        </Popconfirm>
        <a-button type="link" size="small" @click="handleToggleStatus(row)">
          {{ row.status === 1 ? '禁用' : '启用' }}
        </a-button>
        <Popconfirm title="确认立即执行？" @confirm="handleRun(row)">
          <a-button type="link" size="small">执行</a-button>
        </Popconfirm>
        <a-button type="link" size="small" @click="$router.push(`/system/job/log?jobName=${row.name}`)">
          日志
        </a-button>
      </template>
    </Grid>
    <Modal />
  </Page>
</template>
