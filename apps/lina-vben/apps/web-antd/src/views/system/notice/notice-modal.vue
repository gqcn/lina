<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenForm, useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { noticeAdd, noticeInfo, noticeUpdate } from '#/api/system/notice';
import { TiptapEditor } from '#/components/tiptap';

const emit = defineEmits<{ reload: [] }>();

const isEdit = computed(() => !!formData.value.id);
const formData = ref<Record<string, any>>({});
const content = ref('');

const [Form, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
  },
  schema: [
    {
      component: 'Input',
      fieldName: 'title',
      label: '公告标题',
      rules: 'required',
    },
    {
      component: 'RadioGroup',
      fieldName: 'status',
      label: '公告状态',
      defaultValue: 0,
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
        options: [
          { label: '草稿', value: 0 },
          { label: '已发布', value: 1 },
        ],
      },
    },
    {
      component: 'RadioGroup',
      fieldName: 'type',
      label: '公告类型',
      defaultValue: 1,
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
        options: [
          { label: '通知', value: 1 },
          { label: '公告', value: 2 },
        ],
      },
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: '备注',
    },
  ],
  layout: 'horizontal',
  wrapperClass: 'grid-cols-1 md:grid-cols-2',
});

const [Modal, modalApi] = useVbenModal({
  class: 'w-[800px]',
  fullscreenButton: true,
  title: computed(() => (isEdit.value ? '编辑通知公告' : '新增通知公告')),
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    const data = modalApi.getData();
    if (data?.id) {
      modalApi.setState({ confirmLoading: true });
      try {
        const record = await noticeInfo(data.id);
        formData.value = { id: record.id };
        await formApi.setValues({
          title: record.title,
          type: record.type,
          status: record.status,
          remark: record.remark,
        });
        content.value = record.content || '';
      } finally {
        modalApi.setState({ confirmLoading: false });
      }
    } else {
      formData.value = {};
      content.value = '';
      await formApi.resetForm();
    }
  },
});

async function handleConfirm() {
  const { valid } = await formApi.validate();
  if (!valid) return;

  if (!content.value || content.value === '<p></p>') {
    message.error('请输入公告内容');
    return;
  }

  const values = await formApi.getValues();
  modalApi.setState({ confirmLoading: true });

  try {
    if (isEdit.value) {
      await noticeUpdate(formData.value.id, {
        ...values,
        content: content.value,
      });
      message.success('更新成功');
    } else {
      await noticeAdd({
        ...values,
        content: content.value,
      });
      message.success('新增成功');
    }
    emit('reload');
    modalApi.close();
  } finally {
    modalApi.setState({ confirmLoading: false });
  }
}
</script>

<template>
  <Modal>
    <Form />
    <div class="px-4 pb-2">
      <div class="mb-2 font-medium">公告内容</div>
      <TiptapEditor v-model="content" :height="300" />
    </div>
  </Modal>
</template>
