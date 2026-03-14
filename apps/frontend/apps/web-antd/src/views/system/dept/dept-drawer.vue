<script setup lang="ts">
import type { DeptTree } from '#/api/system/dept/model';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import {
  deptAdd,
  deptExclude,
  deptInfo,
  deptTree,
  deptUpdate,
  deptUsers,
} from '#/api/system/dept';

import { drawerSchema } from './data';

const emit = defineEmits<{ reload: [] }>();

interface DrawerProps {
  id?: number;
  update: boolean;
}

const isUpdate = ref(false);
const deptId = ref<number>(0);
const title = computed(() => (isUpdate.value ? '编辑部门' : '新增部门'));

const [BasicForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-2',
    labelWidth: 80,
  },
  schema: drawerSchema(),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

/** 为树节点添加 fullName（显示完整路径） */
function addFullName(
  tree: DeptTree[],
  parentPath = '',
  separator = ' / ',
) {
  for (const node of tree) {
    const fullName = parentPath
      ? `${parentPath}${separator}${node.label}`
      : node.label;
    (node as any).fullName = fullName;
    if (node.children && node.children.length > 0) {
      addFullName(node.children, fullName, separator);
    }
  }
}

/** 初始化部门树选择 */
async function initDeptSelect(id?: number) {
  let treeData: DeptTree[];
  if (isUpdate.value && id) {
    // 编辑时排除自身及子节点
    treeData = await deptExclude(id);
  } else {
    treeData = await deptTree();
  }
  // 添加完整路径名
  addFullName(treeData);
  formApi.updateSchema([
    {
      componentProps: {
        fieldNames: { label: 'label', value: 'id' },
        showSearch: true,
        treeData,
        treeDefaultExpandAll: true,
        treeLine: { showLeafIcon: false },
        treeNodeFilterProp: 'label',
        treeNodeLabelProp: 'fullName',
      },
      fieldName: 'parentId',
    },
  ]);
}

/** 初始化部门负责人下拉 */
async function initDeptUsers(id: number) {
  const ret = await deptUsers(id);
  const options = ret.map((user) => ({
    label: `${user.username} | ${user.nickname}`,
    value: user.id,
  }));
  formApi.updateSchema([
    {
      componentProps: {
        disabled: ret.length === 0,
        options,
        placeholder: ret.length === 0 ? '该部门暂无用户' : '请选择部门负责人',
      },
      fieldName: 'leader',
    },
  ]);
}

/** 新增时禁用负责人选择 */
function setLeaderDisabled() {
  formApi.updateSchema([
    {
      componentProps: {
        disabled: true,
        options: [],
        placeholder: '仅在更新时可选部门负责人',
      },
      fieldName: 'leader',
    },
  ]);
}

const [BasicDrawer, drawerApi] = useVbenDrawer({
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  async onOpenChange(isOpen) {
    if (!isOpen) {
      return;
    }
    drawerApi.setState({ loading: true });

    const { id, update } = drawerApi.getData() as DrawerProps;
    isUpdate.value = update;

    if (id) {
      await formApi.setFieldValue('parentId', id);
      if (update) {
        deptId.value = id;
        const record = await deptInfo(id);
        await formApi.setValues(record);
      }
    }

    await (update && id ? initDeptUsers(id) : setLeaderDisabled());
    await initDeptSelect(id);

    drawerApi.setState({ loading: false });
  },
});

async function handleConfirm() {
  try {
    drawerApi.lock(true);
    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const data = await formApi.getValues();
    if (isUpdate.value) {
      await deptUpdate(deptId.value, data);
      message.success('更新成功');
    } else {
      await deptAdd(data);
      message.success('创建成功');
    }
    emit('reload');
    drawerApi.close();
  } catch (error) {
    console.error(error);
  } finally {
    drawerApi.lock(false);
  }
}

async function handleClosed() {
  await formApi.resetForm();
}
</script>

<template>
  <BasicDrawer :title="title" class="w-[600px]">
    <BasicForm />
  </BasicDrawer>
</template>
