<script setup lang="ts">
import { ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';

import {
  Dropdown,
  Menu,
  MenuItem,
  message,
  Modal,
  Popconfirm,
  Space,
  Switch,
  Upload,
} from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  userDelete,
  userExport,
  userImport,
  userImportTemplate,
  userList,
  userStatusChange,
} from '#/api/system/user';

import { columns, querySchema } from './data';
import UserDrawer from './user-drawer.vue';

const [UserDrawerRef, userDrawerApi] = useVbenDrawer({
  connectedComponent: UserDrawer,
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
        query: async ({ page, sorts }, formValues = {}) => {
          const sortParams: Record<string, string> = {};
          if (sorts && sorts.length > 0) {
            const sort = sorts[0];
            if (sort && sort.order) {
              sortParams.orderBy = sort.field;
              sortParams.orderDirection = sort.order;
            }
          }
          // Handle createdAt date range
          const params: Record<string, any> = {
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
            ...sortParams,
          };
          if (params.createdAt && Array.isArray(params.createdAt)) {
            params.beginTime = params.createdAt[0];
            params.endTime = params.createdAt[1];
            delete params.createdAt;
          }
          return await userList(params);
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
    id: 'system-user-index',
  },
});

function handleAdd() {
  userDrawerApi.setData({ isEdit: false });
  userDrawerApi.open();
}

function handleEdit(row: any) {
  userDrawerApi.setData({ isEdit: true, row });
  userDrawerApi.open();
}

async function handleDelete(row: any) {
  await userDelete(row.id);
  message.success('删除成功');
  await gridApi.query();
}

function handleMultiDelete() {
  const rows = gridApi.grid.getCheckboxRecords();
  const ids = rows.map((row: any) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条记录吗？`,
    onOk: async () => {
      for (const id of ids) {
        await userDelete(id);
      }
      await gridApi.query();
    },
  });
}

async function handleStatusChange(row: any) {
  await userStatusChange(row.id, row.status);
}

function onDrawerSuccess() {
  gridApi.query();
}

const importModalVisible = ref(false);

async function handleExport() {
  try {
    const formValues = gridApi.formApi.form.values || {};
    const params: Record<string, any> = { ...formValues };
    if (params.createdAt && Array.isArray(params.createdAt)) {
      params.beginTime = params.createdAt[0];
      params.endTime = params.createdAt[1];
      delete params.createdAt;
    }
    const data = await userExport(params);
    const blob = new Blob([data as any], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'users.xlsx';
    link.click();
    window.URL.revokeObjectURL(url);
    message.success('导出成功');
  } catch {
    message.error('导出失败');
  }
}

async function handleDownloadTemplate() {
  try {
    const data = await userImportTemplate();
    const blob = new Blob([data as any], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'user-import-template.xlsx';
    link.click();
    window.URL.revokeObjectURL(url);
  } catch {
    message.error('下载模板失败');
  }
}

async function handleImportUpload(info: any) {
  const file = info.file;
  if (!file) return;
  try {
    const result = await userImport(file);
    const res = result as any;
    if (res.fail > 0) {
      const failReasons = res.failList
        .slice(0, 5)
        .map((item: any) => `第${item.row}行: ${item.reason}`)
        .join('\n');
      Modal.info({
        title: '导入结果',
        content: `成功 ${res.success} 条，失败 ${res.fail} 条\n${failReasons}${res.failList.length > 5 ? '\n...' : ''}`,
      });
    } else {
      message.success(`成功导入 ${res.success} 条用户数据`);
    }
    importModalVisible.value = false;
    gridApi.query();
  } catch {
    message.error('导入失败');
  }
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid class="h-full" table-title="用户列表">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleExport">导 出</a-button>
          <a-button @click="importModalVisible = true">导 入</a-button>
          <a-button danger type="primary" @click="handleMultiDelete">
            删 除
          </a-button>
          <a-button type="primary" @click="handleAdd">新 增</a-button>
        </Space>
      </template>

      <template #status="{ row }">
        <Switch
          v-model:checked="row.status"
          :checked-value="1"
          :un-checked-value="0"
          checked-children="启用"
          un-checked-children="禁用"
          @change="() => handleStatusChange(row)"
        />
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
        <Dropdown placement="bottomRight">
          <template #overlay>
            <Menu>
              <MenuItem key="resetPwd">重置密码</MenuItem>
              <MenuItem key="assignRole">分配角色</MenuItem>
            </Menu>
          </template>
          <a-button size="small" type="link">更多</a-button>
        </Dropdown>
      </template>
    </Grid>

    <UserDrawerRef @success="onDrawerSuccess" />

    <Modal
      v-model:open="importModalVisible"
      :footer="null"
      title="导入用户"
      width="480px"
    >
      <div class="py-4">
        <p class="mb-4 text-gray-500">
          请先下载导入模板，按模板格式填写数据后上传。
        </p>
        <div class="mb-4">
          <a-button type="link" @click="handleDownloadTemplate">
            下载导入模板
          </a-button>
        </div>
        <Upload
          :before-upload="() => false"
          :max-count="1"
          :show-upload-list="false"
          accept=".xlsx,.xls"
          @change="handleImportUpload"
        >
          <a-button type="primary">选择文件并导入</a-button>
        </Upload>
      </div>
    </Modal>
  </Page>
</template>
