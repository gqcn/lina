<script setup lang="ts">
import { Descriptions, DescriptionsItem } from 'ant-design-vue';
import { useDictStore } from '#/store/dict';

interface Props {
  formApi: any;
}

const props = defineProps<Props>();
const dictStore = useDictStore();

const log = props.formApi?.values || {};
</script>

<template>
  <Descriptions bordered :column="2">
    <DescriptionsItem label="任务名称">{{ log.jobName }}</DescriptionsItem>
    <DescriptionsItem label="任务分组">{{ log.jobGroup }}</DescriptionsItem>
    <DescriptionsItem label="执行指令" :span="2">{{ log.command }}</DescriptionsItem>
    <DescriptionsItem label="开始时间">{{ log.startTime }}</DescriptionsItem>
    <DescriptionsItem label="结束时间">{{ log.endTime || '-' }}</DescriptionsItem>
    <DescriptionsItem label="执行耗时">{{ log.duration ? `${log.duration}ms` : '-' }}</DescriptionsItem>
    <DescriptionsItem label="执行状态">
      <a-tag :color="log.status === 1 ? 'success' : 'error'">
        {{ dictStore.getDictLabel('sys_job_log_status', log.status) }}
      </a-tag>
    </DescriptionsItem>
    <DescriptionsItem v-if="log.errorMsg" label="错误信息" :span="2">
      <pre style="white-space: pre-wrap; word-break: break-all;">{{ log.errorMsg }}</pre>
    </DescriptionsItem>
  </Descriptions>
</template>
