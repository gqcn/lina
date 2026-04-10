<script lang="ts" setup>
import type { NotificationItem } from '@vben/layouts';

import { computed, onBeforeUnmount, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';

import { AuthenticationLoginExpiredModal, useVbenModal } from '@vben/common-ui';
import { useWatermark } from '@vben/hooks';
import {
  BasicLayout,
  LockScreen,
  Notification,
  UserDropdown,
} from '@vben/layouts';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';

import { Modal } from 'ant-design-vue';

import PluginSlotOutlet from '#/components/plugin/plugin-slot-outlet.vue';
import { $t } from '#/locales';
import { pluginSlotKeys } from '#/plugins/plugin-slots';
import {
  notifyPluginRegistryChangedIfNeeded,
  onPluginRegistryChanged,
} from '#/plugins/slot-registry';
import { refreshAccessibleState } from '#/router/access-refresh';
import { useAuthStore } from '#/store';
import { useMessageStore } from '#/store/message';
import LoginForm from '#/views/_core/authentication/login.vue';
import NoticePreviewModal from '#/views/system/notice/notice-preview-modal.vue';

const router = useRouter();
const userStore = useUserStore();
const authStore = useAuthStore();
const accessStore = useAccessStore();
const messageStore = useMessageStore();
const { destroyWatermark, updateWatermark } = useWatermark();

const [PreviewModal, previewModalApi] = useVbenModal({
  connectedComponent: NoticePreviewModal,
});

let disposePluginRegistryListener: (() => void) | null = null;

// Map server messages to NotificationItem format
const notifications = computed<NotificationItem[]>(() =>
  messageStore.messages.map((msg) => ({
    id: msg.id,
    avatar: '',
    date: msg.createdAt,
    isRead: msg.isRead === 1,
    message: msg.title,
    title: msg.type === 1 ? '通知' : '公告',
    sourceType: msg.sourceType,
    sourceId: msg.sourceId,
  })),
);

const showDot = computed(() => messageStore.unreadCount > 0);

// Start polling on mount
onMounted(() => {
  messageStore.startPolling();
});

const menus = computed(() => [
  {
    handler: () => {
      router.push({ name: 'Profile' });
    },
    icon: 'lucide:user',
    text: $t('page.auth.profile'),
  },
]);

const avatar = computed(() => {
  return userStore.userInfo?.avatar || preferences.app.defaultAvatar;
});

async function handleLogout() {
  messageStore.stopPolling();
  await authStore.logout(false);
}

async function handleNoticeClear() {
  Modal.confirm({
    title: '提示',
    content: '确认清空所有消息通知？',
    onOk: async () => {
      await messageStore.clearAll();
    },
  });
}

async function handleRead(item: NotificationItem) {
  if (item.id) {
    await messageStore.markRead(item.id as number);
  }
}

async function handleRemove(item: NotificationItem) {
  if (item.id) {
    await messageStore.removeMessage(item.id as number);
  }
}

async function handleMakeAll() {
  await messageStore.markAllRead();
}

function handleViewAll() {
  router.push('/system/message');
}

function handleNotificationClick(item: NotificationItem) {
  const msg = messageStore.messages.find((m) => m.id === item.id);
  if (msg?.sourceType === 'notice' && msg?.sourceId) {
    previewModalApi.setData({ id: msg.sourceId });
    previewModalApi.open();
  }
}

async function refreshPluginAwareAccess() {
  await refreshAccessibleState(router, { showLoadingToast: false });
}

function handlePluginRegistryMaybeChanged() {
  if (
    typeof document !== 'undefined' &&
    document.visibilityState &&
    document.visibilityState !== 'visible'
  ) {
    return;
  }
  void notifyPluginRegistryChangedIfNeeded();
}

// Fetch messages when notification panel is likely to open
// The Notification component triggers @read when opened
// We fetch on mount to have data ready
onMounted(() => {
  messageStore.fetchMessages();
  disposePluginRegistryListener = onPluginRegistryChanged(() =>
    refreshPluginAwareAccess(),
  );
  window.addEventListener('focus', handlePluginRegistryMaybeChanged);
  document.addEventListener(
    'visibilitychange',
    handlePluginRegistryMaybeChanged,
  );
});

onBeforeUnmount(() => {
  disposePluginRegistryListener?.();
  disposePluginRegistryListener = null;
  window.removeEventListener('focus', handlePluginRegistryMaybeChanged);
  document.removeEventListener(
    'visibilitychange',
    handlePluginRegistryMaybeChanged,
  );
});

watch(
  () => ({
    enable: preferences.app.watermark,
    content: preferences.app.watermarkContent,
  }),
  async ({ enable, content }) => {
    if (enable) {
      await updateWatermark({
        content:
          content ||
          `${userStore.userInfo?.username} - ${userStore.userInfo?.realName}`,
      });
    } else {
      destroyWatermark();
    }
  },
  {
    immediate: true,
  },
);
</script>

<template>
  <BasicLayout @clear-preferences-and-logout="handleLogout">
    <template #header-right-45>
      <PluginSlotOutlet
        :slot-key="pluginSlotKeys.layoutHeaderActionsBefore"
        class="mr-2"
      />
    </template>
    <template #header-right-145>
      <PluginSlotOutlet
        :slot-key="pluginSlotKeys.layoutHeaderActionsAfter"
        class="mr-2"
      />
    </template>
    <template #user-dropdown>
      <div class="flex items-center">
        <PluginSlotOutlet
          :slot-key="pluginSlotKeys.layoutUserDropdownAfter"
          class="mr-2"
        />
        <UserDropdown
          :avatar
          :menus
          :text="userStore.userInfo?.realName"
          :description="userStore.userInfo?.email || ''"
          :tag-text="userStore.userInfo?.username"
          @logout="handleLogout"
        />
      </div>
    </template>
    <template #notification>
      <Notification
        :dot="showDot"
        :notifications="notifications"
        @clear="handleNoticeClear"
        @click="handleNotificationClick"
        @read="handleRead"
        @remove="handleRemove"
        @make-all="handleMakeAll"
        @view-all="handleViewAll"
      />
    </template>
    <template #extra>
      <AuthenticationLoginExpiredModal
        v-model:open="accessStore.loginExpired"
        :avatar
      >
        <LoginForm />
      </AuthenticationLoginExpiredModal>
    </template>
    <template #lock-screen>
      <LockScreen :avatar @to-login="handleLogout" />
    </template>
  </BasicLayout>
  <PreviewModal />
</template>
