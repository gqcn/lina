<script setup lang="ts">
import type { Notice } from '#/api/system/notice/model';

import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Page } from '@vben/common-ui';

import { Button, Descriptions, DescriptionsItem } from 'ant-design-vue';

import { noticeInfo } from '#/api/system/notice';
import { DictTag } from '#/components/dict';
import { useDictStore } from '#/store/dict';

const route = useRoute();
const router = useRouter();
const notice = ref<Notice | null>(null);
const loading = ref(true);

const dictStore = useDictStore();
const noticeTypeDicts = ref<any[]>([]);

onMounted(async () => {
  noticeTypeDicts.value = await dictStore.getDictOptions('sys_notice_type');
  const id = Number(route.params.id);
  if (!id) {
    router.back();
    return;
  }
  try {
    notice.value = await noticeInfo(id);
  } finally {
    loading.value = false;
  }
});

function goBack() {
  router.back();
}
</script>

<template>
  <Page>
    <div v-if="loading" class="p-8 text-center">加载中...</div>
    <div v-else-if="notice" class="mx-auto max-w-[900px] p-6">
      <div class="mb-6">
        <Button type="link" class="p-0 mb-4" @click="goBack">
          &larr; 返回
        </Button>
        <h1 class="text-2xl font-bold mb-4">{{ notice.title }}</h1>
        <Descriptions :column="3" size="small" bordered>
          <DescriptionsItem label="公告类型">
            <DictTag :dicts="noticeTypeDicts" :value="String(notice.type)" />
          </DescriptionsItem>
          <DescriptionsItem label="创建人">
            {{ notice.createdByName || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="创建时间">
            {{ notice.createdAt }}
          </DescriptionsItem>
        </Descriptions>
      </div>
      <div
        class="notice-content prose max-w-none"
        v-html="notice.content"
      />
    </div>
    <div v-else class="p-8 text-center text-gray-400">通知公告不存在</div>
  </Page>
</template>

<style scoped>
.notice-content :deep(img) {
  max-width: 100%;
  height: auto;
}

.notice-content :deep(h1) {
  font-size: 2em;
  font-weight: bold;
  margin: 0.67em 0;
}

.notice-content :deep(h2) {
  font-size: 1.5em;
  font-weight: bold;
  margin: 0.83em 0;
}

.notice-content :deep(h3) {
  font-size: 1.17em;
  font-weight: bold;
  margin: 1em 0;
}

.notice-content :deep(ul),
.notice-content :deep(ol) {
  padding-left: 1.5em;
  margin: 0.5em 0;
}

.notice-content :deep(ul) {
  list-style-type: disc;
}

.notice-content :deep(ol) {
  list-style-type: decimal;
}

.notice-content :deep(blockquote) {
  border-left: 3px solid #d9d9d9;
  padding-left: 1em;
  margin: 0.5em 0;
  color: #666;
}
</style>
