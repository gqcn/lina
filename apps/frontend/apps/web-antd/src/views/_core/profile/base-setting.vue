<script setup lang="ts">
import type { SysUser } from '#/api/system/user';
import type { VbenFormSchema } from '#/adapter/form';

import { computed, onMounted } from 'vue';

import { useVbenForm } from '#/adapter/form';
import { updateProfile } from '#/api/system/user';

import { message } from 'ant-design-vue';

const props = defineProps<{ profile: SysUser }>();

const emit = defineEmits<{ updated: [] }>();

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      fieldName: 'nickname',
      component: 'Input',
      label: '昵称',
      componentProps: {
        placeholder: '请输入昵称',
      },
    },
    {
      fieldName: 'email',
      component: 'Input',
      label: '邮箱',
      componentProps: {
        placeholder: '请输入邮箱',
      },
    },
    {
      fieldName: 'phone',
      component: 'Input',
      label: '手机号',
      componentProps: {
        placeholder: '请输入手机号',
      },
    },
  ];
});

function buttonLoading(loading: boolean) {
  formApi.setState({ submitButtonOptions: { loading } });
}

const [Form, formApi] = useVbenForm({
  schema: formSchema,
  commonConfig: {
    labelWidth: 80,
    componentProps: {
      class: 'w-full',
    },
  },
  wrapperClass: 'grid-cols-1',
  resetButtonOptions: { show: false },
  submitButtonOptions: { content: '更新信息' },
  async handleSubmit(values) {
    buttonLoading(true);
    try {
      await updateProfile(values);
      message.success('更新成功');
      emit('updated');
    } finally {
      buttonLoading(false);
    }
  },
});

onMounted(() => {
  formApi.setValues({
    nickname: props.profile.nickname,
    email: props.profile.email,
    phone: props.profile.phone,
  });
});
</script>
<template>
  <div class="mt-[16px] md:w-full lg:w-1/2 2xl:w-2/5">
    <Form />
  </div>
</template>
