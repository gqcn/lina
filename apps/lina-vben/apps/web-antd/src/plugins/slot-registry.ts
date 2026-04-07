import type { PluginRuntimeState } from '#/api/system/plugin/model';
import type { PluginSlotKey } from '#/plugins/plugin-slots';
import type { Component } from 'vue';
import type { VirtualPluginSlotModuleEntry } from 'virtual:lina-plugin-slots';

import { pluginRuntimeList } from '#/api/system/plugin';
import { isPluginSlotKey } from '#/plugins/plugin-slots';
import { pluginSlotModules } from 'virtual:lina-plugin-slots';

type PluginRegistryListener = () => void | Promise<void>;

type PluginRegistryGlobal = typeof globalThis & {
  __linaPluginRegistryCheckPromise?: null | Promise<boolean>;
  __linaPluginRegistryListeners?: Set<PluginRegistryListener>;
  __linaPluginStateSignature?: null | string;
  __linaPluginStatePromise?: null | Promise<Map<string, PluginRuntimeState>>;
};

export interface PluginSlotMeta {
  order?: number;
  pluginId?: string;
  slotKey?: PluginSlotKey;
}

export interface RegisteredPluginSlotModule {
  component: Component;
  filePath: string;
  key: string;
  order: number;
  pluginId: string;
  slotKey: PluginSlotKey;
}

const registeredPluginSlotModules = pluginSlotModules
  .map((item: VirtualPluginSlotModuleEntry) => {
    const match = item.filePath.match(
      /\/lina-plugins\/([^/]+)\/frontend\/slots\/(.+)\.vue$/,
    );
    if (!match?.[1] || !match[2] || !item.module.default) {
      return null;
    }

    const pluginId = item.module.pluginSlotMeta?.pluginId || match[1];
    const relativePath = match[2];
    const segments = relativePath.split('/');
    const slotKey =
      item.module.pluginSlotMeta?.slotKey ||
      segments.slice(0, Math.max(segments.length - 1, 0)).join('/');
    const slotName = segments.at(-1) || relativePath;

    if (!slotKey || !isPluginSlotKey(slotKey)) {
      console.warn(
        `[plugin-slot] skip unpublished slot "${slotKey}" from ${item.filePath}`,
      );
      return null;
    }

    return {
      component: item.module.default as Component,
      filePath: item.filePath,
      key: `${pluginId}:${slotKey}:${slotName}`,
      order: item.module.pluginSlotMeta?.order ?? 0,
      pluginId,
      slotKey,
    } satisfies RegisteredPluginSlotModule;
  })
  .filter((item): item is RegisteredPluginSlotModule => item !== null)
  .sort((a, b) => {
    if (a.order !== b.order) {
      return a.order - b.order;
    }
    return a.key.localeCompare(b.key);
  });

function normalizePluginKeys(item: PluginRuntimeState): string[] {
  const keys = [item.id];
  if (item.statusKey?.startsWith('sys_plugin.status:')) {
    keys.push(item.statusKey.substring('sys_plugin.status:'.length));
  }
  return keys.filter((key): key is string => !!key);
}

function getPluginRegistryGlobal() {
  return globalThis as PluginRegistryGlobal;
}

function getPluginRegistryListeners() {
  const registryGlobal = getPluginRegistryGlobal();
  registryGlobal.__linaPluginRegistryListeners ??= new Set();
  return registryGlobal.__linaPluginRegistryListeners;
}

function getPluginStatePromise() {
  return getPluginRegistryGlobal().__linaPluginStatePromise ?? null;
}

function getPluginStateSignature() {
  return getPluginRegistryGlobal().__linaPluginStateSignature ?? null;
}

function setPluginStatePromise(
  promise: null | Promise<Map<string, PluginRuntimeState>>,
) {
  getPluginRegistryGlobal().__linaPluginStatePromise = promise;
}

function setPluginStateSignature(signature: null | string) {
  getPluginRegistryGlobal().__linaPluginStateSignature = signature;
}

function buildPluginStateMap(items: PluginRuntimeState[]) {
  const map = new Map<string, PluginRuntimeState>();
  for (const item of items) {
    for (const key of normalizePluginKeys(item)) {
      map.set(key, item);
    }
  }
  return map;
}

function buildPluginStateSignature(items: PluginRuntimeState[]) {
  return items
    .map((item) => `${item.id}:${item.installed}:${item.enabled}:${item.statusKey}`)
    .sort()
    .join('|');
}

function setPluginStateSnapshot(items: PluginRuntimeState[]) {
  const pluginStateMap = buildPluginStateMap(items);
  setPluginStateSignature(buildPluginStateSignature(items));
  setPluginStatePromise(Promise.resolve(pluginStateMap));
  return pluginStateMap;
}

async function loadPluginStateMap(force = false) {
  let pluginStatePromise = getPluginStatePromise();
  if (!pluginStatePromise || force) {
    pluginStatePromise = pluginRuntimeList()
      .then((items) => {
        return setPluginStateSnapshot(items);
      })
      .catch((error) => {
        console.error('[plugin-slot] failed to load plugin state map', error);
        return new Map<string, PluginRuntimeState>();
      });
    setPluginStatePromise(pluginStatePromise);
  }
  return pluginStatePromise;
}

/**
 * Returns plugin slot definitions for a given slot key.
 */
export function getPluginSlots(
  slotKey: PluginSlotKey,
): RegisteredPluginSlotModule[] {
  return registeredPluginSlotModules.filter((item) => item.slotKey === slotKey);
}

/**
 * Queries current plugin runtime states from host backend.
 */
export async function getPluginStateMap(force = false) {
  return await loadPluginStateMap(force);
}

/**
 * Notifies plugin-aware UI that plugin registry state changed.
 */
export async function notifyPluginRegistryChanged() {
  setPluginStatePromise(null);
  setPluginStateSignature(null);
  await Promise.allSettled(
    Array.from(getPluginRegistryListeners(), (listener) =>
      Promise.resolve(listener()),
    ),
  );
}

/**
 * Queries latest plugin runtime state and only notifies listeners when it actually changed.
 */
export async function notifyPluginRegistryChangedIfNeeded() {
  const registryGlobal = getPluginRegistryGlobal();
  if (registryGlobal.__linaPluginRegistryCheckPromise) {
    return await registryGlobal.__linaPluginRegistryCheckPromise;
  }

  registryGlobal.__linaPluginRegistryCheckPromise = (async () => {
    try {
      const items = await pluginRuntimeList();
      const nextSignature = buildPluginStateSignature(items);

      if (nextSignature === getPluginStateSignature()) {
        return false;
      }

      setPluginStateSnapshot(items);
      await Promise.allSettled(
        Array.from(getPluginRegistryListeners(), (listener) =>
          Promise.resolve(listener()),
        ),
      );
      return true;
    } catch (error) {
      console.error(
        '[plugin-slot] failed to check plugin registry changes',
        error,
      );
      return false;
    } finally {
      registryGlobal.__linaPluginRegistryCheckPromise = null;
    }
  })();

  return await registryGlobal.__linaPluginRegistryCheckPromise;
}

/**
 * Subscribes to plugin registry changes.
 */
export function onPluginRegistryChanged(listener: () => void | Promise<void>) {
  const pluginRegistryListeners = getPluginRegistryListeners();
  pluginRegistryListeners.add(listener);
  return () => {
    pluginRegistryListeners.delete(listener);
  };
}
