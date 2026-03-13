<script setup lang="ts">
import { Page, useVbenDrawer, useVbenModal } from '@vben/common-ui';
import { useUserStore } from '@vben/stores';

import { computed, ref } from 'vue';

import {
  Dropdown,
  Menu,
  MenuItem,
  message,
  Modal,
  Popconfirm,
  Space,
  Switch,
} from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  userDelete,
  userExport,
  userList,
  userStatusChange,
} from '#/api/system/user';
import { downloadBlob } from '#/utils/download';

import { columns, querySchema } from './data';
import UserDrawer from './user-drawer.vue';
import UserImportModal from './user-import-modal.vue';
import UserResetPwdModal from './user-reset-pwd-modal.vue';

const [UserDrawerRef, userDrawerApi] = useVbenDrawer({
  connectedComponent: UserDrawer,
});

const [UserImportModalRef, userImportModalApi] = useVbenModal({
  connectedComponent: UserImportModal,
});

const [UserResetPwdModalRef, userResetPwdModalApi] = useVbenModal({
  connectedComponent: UserResetPwdModal,
});

const userStore = useUserStore();

function isSelf(row: any) {
  return row.id === Number(userStore.userInfo?.userId);
}

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
      checkMethod: ({ row }: any) => !isSelf(row),
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
  gridEvents: {
    checkboxChange: () => {
      checkedRows.value = gridApi.grid?.getCheckboxRecords() || [];
    },
    checkboxAll: () => {
      checkedRows.value = gridApi.grid?.getCheckboxRecords() || [];
    },
  },
});

const checkedRows = ref<any[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

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
      checkedRows.value = [];
      await gridApi.query();
    },
  });
}

async function handleStatusChange(row: any) {
  await userStatusChange(row.id, row.status);
}

function onReload() {
  gridApi.query();
}

async function handleExport() {
  try {
    const ids = checkedRows.value.map((row: any) => row.id);
    const data = await userExport({ ids });
    downloadBlob(data, 'users.xlsx');
    message.success('导出成功');
  } catch {
    message.error('导出失败');
  }
}

function handleImport() {
  userImportModalApi.open();
}

function handleResetPwd(row: any) {
  userResetPwdModalApi.setData({ record: row });
  userResetPwdModalApi.open();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid class="h-full" table-title="用户列表">
      <template #toolbar-tools>
        <Space>
          <a-button :disabled="!hasChecked" @click="handleExport">
            导 出
          </a-button>
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

      <template #status="{ row }">
        <Switch
          v-model:checked="row.status"
          :checked-value="1"
          :disabled="isSelf(row)"
          :un-checked-value="0"
          checked-children="启用"
          un-checked-children="禁用"
          @change="() => handleStatusChange(row)"
        />
      </template>

      <template #action="{ row }">
        <template v-if="!isSelf(row)">
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
                <MenuItem key="resetPwd" @click="handleResetPwd(row)">
                  重置密码
                </MenuItem>
              </Menu>
            </template>
            <a-button size="small" type="link">更多</a-button>
          </Dropdown>
        </template>
      </template>
    </Grid>

    <UserDrawerRef @success="onReload" />
    <UserImportModalRef @reload="onReload" />
    <UserResetPwdModalRef @reload="onReload" />
  </Page>
</template>
