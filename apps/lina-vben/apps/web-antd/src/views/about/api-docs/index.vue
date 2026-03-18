<script setup lang="ts">
defineOptions({ name: 'ApiDocs' });
</script>

<template>
  <iframe class="api-docs-iframe" src="/stoplight/apidocs.html" />
</template>

<style scoped>
.api-docs-iframe {
  width: 100%;
  height: calc(100vh - 100px);
  border: none;
  display: block;
}
</style>
