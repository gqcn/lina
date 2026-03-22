<script setup lang="ts">
import type { ServerNodeInfo } from '#/api/monitor/server/model';

import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Col,
  Descriptions,
  DescriptionsItem,
  Progress,
  Row,
  Select,
  SelectOption,
  Table,
} from 'ant-design-vue';

import { getServerMonitor } from '#/api/monitor/server';

const nodes = ref<ServerNodeInfo[]>([]);
const selectedNode = ref<string>('');
const loading = ref(false);

const currentNode = computed(() => {
  if (!selectedNode.value || nodes.value.length === 0) {
    return nodes.value[0] ?? null;
  }
  return (
    nodes.value.find(
      (n) => `${n.nodeName}|${n.nodeIp}` === selectedNode.value,
    ) ?? nodes.value[0]
  );
});

const showNodeSelector = computed(() => nodes.value.length > 1);

onMounted(async () => {
  await loadData();
});

async function loadData() {
  loading.value = true;
  try {
    const resp = await getServerMonitor();
    nodes.value = resp.nodes ?? [];
    if (nodes.value.length > 0 && !selectedNode.value) {
      const first = nodes.value[0]!;
      selectedNode.value = `${first.nodeName}|${first.nodeIp}`;
    }
  } finally {
    loading.value = false;
  }
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / k ** i).toFixed(2)} ${sizes[i]}`;
}

function formatRate(bytesPerSec: number): string {
  return `${formatBytes(bytesPerSec)}/s`;
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const parts: string[] = [];
  if (days > 0) parts.push(`${days}天`);
  if (hours > 0) parts.push(`${hours}小时`);
  if (mins > 0) parts.push(`${mins}分钟`);
  return parts.join(' ') || '刚启动';
}

function getProgressColor(percent: number): string {
  if (percent >= 90) return '#ff4d4f';
  if (percent >= 70) return '#faad14';
  return '#52c41a';
}

const diskColumns = [
  { title: '盘符', dataIndex: 'path', key: 'path' },
  { title: '文件系统', dataIndex: 'fsType', key: 'fsType' },
  {
    title: '总容量',
    dataIndex: 'total',
    key: 'total',
    customRender: ({ text }: any) => formatBytes(text),
  },
  {
    title: '已用',
    dataIndex: 'used',
    key: 'used',
    customRender: ({ text }: any) => formatBytes(text),
  },
  {
    title: '可用',
    dataIndex: 'free',
    key: 'free',
    customRender: ({ text }: any) => formatBytes(text),
  },
  {
    title: '使用率',
    dataIndex: 'usagePercent',
    key: 'usagePercent',
    width: 200,
  },
];
</script>

<template>
  <Page>
    <div v-if="currentNode" class="flex flex-col gap-4">
      <!-- Node Selector + Refresh -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <span v-if="showNodeSelector" class="text-sm text-gray-500">
            节点：
          </span>
          <Select
            v-if="showNodeSelector"
            v-model:value="selectedNode"
            style="width: 260px"
          >
            <SelectOption
              v-for="node in nodes"
              :key="`${node.nodeName}|${node.nodeIp}`"
              :value="`${node.nodeName}|${node.nodeIp}`"
            >
              {{ node.nodeName }} ({{ node.nodeIp }})
            </SelectOption>
          </Select>
        </div>
        <div class="flex items-center gap-2 text-xs text-gray-400">
          <span>采集时间：{{ currentNode.collectAt }}</span>
          <Button size="small" @click="loadData">
            <span class="icon-[charm--refresh]"></span>
          </Button>
        </div>
      </div>

      <!-- Server Info -->
      <Card size="small" title="服务器信息">
        <Descriptions :column="{ xs: 1, sm: 2, md: 3, lg: 4 }" size="small">
          <DescriptionsItem label="主机名">
            {{ currentNode.server?.hostname }}
          </DescriptionsItem>
          <DescriptionsItem label="操作系统">
            {{ currentNode.server?.os }}
          </DescriptionsItem>
          <DescriptionsItem label="系统架构">
            {{ currentNode.server?.arch }}
          </DescriptionsItem>
          <DescriptionsItem label="系统运行时长">
            {{ formatUptime(currentNode.server?.uptime ?? 0) }}
          </DescriptionsItem>
          <DescriptionsItem label="系统启动时间">
            {{ currentNode.server?.bootTime }}
          </DescriptionsItem>
          <DescriptionsItem label="服务启动时间">
            {{ currentNode.server?.startTime }}
          </DescriptionsItem>
          <DescriptionsItem label="节点IP">
            {{ currentNode.nodeIp }}
          </DescriptionsItem>
        </Descriptions>
      </Card>

      <!-- CPU + Memory + Go Runtime -->
      <Row :gutter="[16, 16]">
        <Col :xs="24" :sm="24" :md="8">
          <Card size="small" title="CPU">
            <div class="flex flex-col items-center gap-3 py-2">
              <Progress
                :percent="
                  Number((currentNode.cpu?.usagePercent ?? 0).toFixed(1))
                "
                :stroke-color="
                  getProgressColor(currentNode.cpu?.usagePercent ?? 0)
                "
                :width="120"
                type="circle"
              />
              <div class="text-center text-xs text-gray-500">
                <div>{{ currentNode.cpu?.cores }} 核</div>
                <div class="mt-1 max-w-[200px] truncate">
                  {{ currentNode.cpu?.modelName }}
                </div>
              </div>
            </div>
          </Card>
        </Col>
        <Col :xs="24" :sm="24" :md="8">
          <Card size="small" title="内存">
            <div class="flex flex-col items-center gap-3 py-2">
              <Progress
                :percent="
                  Number((currentNode.memory?.usagePercent ?? 0).toFixed(1))
                "
                :stroke-color="
                  getProgressColor(currentNode.memory?.usagePercent ?? 0)
                "
                :width="120"
                type="circle"
              />
              <div class="text-center text-xs text-gray-500">
                <div>
                  {{ formatBytes(currentNode.memory?.used ?? 0) }} /
                  {{ formatBytes(currentNode.memory?.total ?? 0) }}
                </div>
                <div class="mt-1">
                  可用：{{ formatBytes(currentNode.memory?.available ?? 0) }}
                </div>
              </div>
            </div>
          </Card>
        </Col>
        <Col :xs="24" :sm="24" :md="8">
          <Card size="small" title="Go 运行时">
            <Descriptions :column="1" size="small" class="py-2">
              <DescriptionsItem label="Go版本">
                {{ currentNode.goInfo?.version }}
              </DescriptionsItem>
              <DescriptionsItem label="GoFrame版本">
                {{ currentNode.goInfo?.gfVersion }}
              </DescriptionsItem>
              <DescriptionsItem label="Goroutines">
                {{ currentNode.goInfo?.goroutines }}
              </DescriptionsItem>
              <DescriptionsItem label="堆内存分配">
                {{ formatBytes(currentNode.goInfo?.heapAlloc ?? 0) }}
              </DescriptionsItem>
              <DescriptionsItem label="堆内存系统">
                {{ formatBytes(currentNode.goInfo?.heapSys ?? 0) }}
              </DescriptionsItem>
              <DescriptionsItem label="GC暂停">
                {{
                  ((currentNode.goInfo?.gcPauseNs ?? 0) / 1_000_000).toFixed(2)
                }}
                ms
              </DescriptionsItem>
            </Descriptions>
          </Card>
        </Col>
      </Row>

      <!-- Disk Usage -->
      <Card size="small" title="磁盘使用">
        <Table
          :columns="diskColumns"
          :data-source="currentNode.disks"
          :pagination="false"
          row-key="path"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'usagePercent'">
              <Progress
                :percent="Number(record.usagePercent.toFixed(1))"
                :stroke-color="getProgressColor(record.usagePercent)"
                size="small"
              />
            </template>
          </template>
        </Table>
      </Card>

      <!-- Network -->
      <Card size="small" title="网络流量">
        <Descriptions :column="{ xs: 1, sm: 2, md: 4 }" size="small">
          <DescriptionsItem label="总发送">
            {{ formatBytes(currentNode.network?.bytesSent ?? 0) }}
          </DescriptionsItem>
          <DescriptionsItem label="总接收">
            {{ formatBytes(currentNode.network?.bytesRecv ?? 0) }}
          </DescriptionsItem>
          <DescriptionsItem label="发送速率">
            {{ formatRate(currentNode.network?.sendRate ?? 0) }}
          </DescriptionsItem>
          <DescriptionsItem label="接收速率">
            {{ formatRate(currentNode.network?.recvRate ?? 0) }}
          </DescriptionsItem>
        </Descriptions>
      </Card>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="!loading"
      class="flex h-[300px] items-center justify-center text-gray-400"
    >
      暂无监控数据，请等待数据采集...
    </div>
  </Page>
</template>
