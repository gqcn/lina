import { ref } from 'vue';

import { defineStore } from 'pinia';

import { dictDataByType } from '#/api/system/dict/dict-data';

interface DictOption {
  label: string;
  value: string;
  tagStyle?: string;
  cssClass?: string;
}

export const useDictStore = defineStore('dict', () => {
  const dictOptionsMap = ref(new Map<string, DictOption[]>());
  const dictRequestCache = new Map<string, Promise<DictOption[]>>();

  async function getDictOptions(dictName: string): Promise<DictOption[]> {
    // Return from cache if exists
    if (dictOptionsMap.value.has(dictName)) {
      return dictOptionsMap.value.get(dictName)!;
    }
    // Dedup concurrent requests
    if (dictRequestCache.has(dictName)) {
      return dictRequestCache.get(dictName)!;
    }
    const promise = dictDataByType(dictName).then((list) => {
      const options = (list || []).map((item) => ({
        label: item.label,
        value: item.value,
        tagStyle: item.tagStyle,
        cssClass: item.cssClass,
      }));
      dictOptionsMap.value.set(dictName, options);
      dictRequestCache.delete(dictName);
      return options;
    });
    dictRequestCache.set(dictName, promise);
    return promise;
  }

  function resetCache() {
    dictOptionsMap.value.clear();
    dictRequestCache.clear();
  }

  function $reset() {
    resetCache();
  }

  return { dictOptionsMap, getDictOptions, resetCache, $reset };
});
