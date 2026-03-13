<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { userAdd, userInfo, userUpdate } from '#/api/system/user';

import { drawerSchema } from './data';

const emit = defineEmits<{ success: [] }>();

const isEdit = ref(false);
const userId = ref<number>(0);
const title = computed(() => (isEdit.value ? '编辑用户' : '新增用户'));

const [Form, formApi] = useVbenForm({
  schema: drawerSchema(false),
  showDefaultActions: false,
});

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (open) {
      const data = drawerApi.getData<{ isEdit: boolean; row?: any }>();
      isEdit.value = data?.isEdit ?? false;

      // Update schema based on mode
      formApi.setState({
        schema: drawerSchema(isEdit.value),
      });

      if (isEdit.value && data?.row) {
        userId.value = data.row.id;
        // Load user info
        const user = await userInfo(data.row.id);
        await formApi.setValues({
          username: user.username,
          nickname: user.nickname,
          email: user.email,
          phone: user.phone,
          status: user.status,
          remark: user.remark,
        });
      } else {
        userId.value = 0;
        await formApi.resetForm();
      }
    }
  },
  async onConfirm() {
    const values = await formApi.getValues();

    if (isEdit.value) {
      await userUpdate({
        id: userId.value,
        ...values,
      });
      message.success('更新成功');
    } else {
      await userAdd(values);
      message.success('创建成功');
    }

    emit('success');
    drawerApi.close();
  },
});
</script>

<template>
  <Drawer :title="title">
    <Form />
  </Drawer>
</template>
