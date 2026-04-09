<script setup lang="ts">
import type { UploadFile } from 'ant-design-vue/es/upload/interface';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

import { Alert, Modal, Switch, Upload } from 'ant-design-vue';

import { pluginRuntimeUpload } from '#/api/system/plugin';

const emit = defineEmits<{ reload: [] }>();

const UploadDragger = Upload.Dragger;

const [BasicModal, modalApi] = useVbenModal({
  onCancel: handleCancel,
  onConfirm: handleSubmit,
});

const fileList = ref<UploadFile[]>([]);
const overwriteSupport = ref(false);
const successMessage = ref('');

async function handleSubmit() {
  if (successMessage.value) {
    handleCancel();
    return;
  }
  try {
    modalApi.setState({ loading: true });
    if (fileList.value.length !== 1) {
      Modal.warning({ title: '请选择一个插件文件' });
      return;
    }

    const uploadItem = fileList.value[0]!;
    const rawFile = uploadItem.originFileObj as Blob | File;
    // Ant Design Upload may expose a Blob-like object here. Rebuilding a
    // concrete File preserves the original `.wasm` filename so the backend can
    // validate the extension and store the artifact under the expected name.
    const file =
      rawFile instanceof File
        ? rawFile
        : new File([rawFile], uploadItem.name || 'runtime-plugin.wasm', {
            type: rawFile.type || 'application/wasm',
          });
    await pluginRuntimeUpload(file, overwriteSupport.value);
    emit('reload');
    fileList.value = [];
    overwriteSupport.value = false;
    successMessage.value = '上传成功，请在插件列表中继续安装并启用。';
    modalApi.setState({
      confirmText: '知道了',
      showCancelButton: false,
    });
  } catch (error) {
    console.warn(error);
  } finally {
    modalApi.setState({ loading: false });
  }
}

function handleCancel() {
  modalApi.close();
  fileList.value = [];
  overwriteSupport.value = false;
  successMessage.value = '';
  modalApi.setState({
    confirmText: undefined,
    showCancelButton: true,
  });
}
</script>

<template>
  <BasicModal
    :close-on-click-modal="false"
    :fullscreen-button="false"
    title="上传插件"
  >
    <template v-if="!successMessage">
      <UploadDragger
        v-model:file-list="fileList"
        :before-upload="() => false"
        :max-count="1"
        :show-upload-list="true"
        accept=".wasm,application/wasm"
        data-testid="plugin-runtime-upload-dragger"
      >
        <p class="ant-upload-drag-icon flex items-center justify-center">
          <IconifyIcon
            class="text-primary text-5xl"
            icon="ant-design:inbox-outlined"
          />
        </p>
        <p class="ant-upload-text">点击或拖拽上传 .wasm 插件包</p>
        <p class="ant-upload-hint">
          仅支持单个 .wasm 文件，上传后可在列表中继续安装并启用。
        </p>
      </UploadDragger>
      <div class="mt-2 flex items-center gap-2">
        <span :class="{ 'text-red-500': overwriteSupport }">
          允许覆盖同 ID 且未安装的插件工作区文件
        </span>
        <div class="flex items-center gap-2">
          <Switch
            v-model:checked="overwriteSupport"
            data-testid="plugin-runtime-overwrite-switch"
          />
        </div>
      </div>
    </template>
    <Alert
      v-else
      :message="successMessage"
      data-testid="plugin-runtime-upload-success"
      show-icon
      type="success"
    />
  </BasicModal>
</template>
