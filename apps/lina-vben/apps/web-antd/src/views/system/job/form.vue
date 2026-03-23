<script setup lang="ts">
import { useVbenForm } from '#/adapter/form';
import { useDictStore } from '#/store/dict';
import { computed } from 'vue';

interface Props {
  formApi: any;
}

const props = defineProps<Props>();
const dictStore = useDictStore();

const isSystemJob = computed(() => props.formApi?.values?.isSystem === 1);

const [Form] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
  },
  schema: [
    {
      component: 'Input',
      fieldName: 'name',
      label: '任务名称',
      rules: 'required|max:64',
    },
    {
      component: 'Input',
      fieldName: 'group',
      label: '任务分组',
      rules: 'required|max:64',
    },
    {
      component: 'Input',
      fieldName: 'command',
      label: '执行指令',
      rules: 'required|max:500',
      componentProps: {
        disabled: isSystemJob.value,
      },
    },
    {
      component: 'Input',
      fieldName: 'cronExpr',
      label: 'Cron表达式',
      rules: 'required',
    },
    {
      component: 'Textarea',
      fieldName: 'description',
      label: '任务描述',
    },
    {
      component: 'RadioGroup',
      fieldName: 'status',
      label: '任务状态',
      defaultValue: 1,
      componentProps: {
        options: dictStore.getDictOptions('sys_job_status'),
        optionType: 'button',
        buttonStyle: 'solid',
      },
    },
    {
      component: 'RadioGroup',
      fieldName: 'singleton',
      label: '执行模式',
      defaultValue: 1,
      componentProps: {
        options: dictStore.getDictOptions('sys_job_singleton'),
        optionType: 'button',
        buttonStyle: 'solid',
      },
    },
    {
      component: 'InputNumber',
      fieldName: 'maxTimes',
      label: '最大执行次数',
      defaultValue: 0,
      help: '0表示无限制',
      componentProps: {
        min: 0,
        class: 'w-full',
      },
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: '备注',
    },
  ],
});
</script>

<template>
  <Form />
</template>
