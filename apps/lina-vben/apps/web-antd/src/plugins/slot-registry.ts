import type { SystemPlugin } from '#/api/system/plugin/model';
import type { PluginSlotKey } from '#/plugins/plugin-slots';
import type { Component } from 'vue';
import type { VirtualPluginSlotModuleEntry } from 'virtual:lina-plugin-slots';

import { pluginList } from '#/api/system/plugin';
import { isPluginSlotKey } from '#/plugins/plugin-slots';
import { pluginSlotModules } from 'virtual:lina-plugin-slots';

const pluginRegistryChangedEvent = 'lina:plugin-registry-changed';

let pluginStatePromise: null | Promise<Map<string, SystemPlugin>> = null;

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
      /\/lina-plugins\/([^/]+)\/frontend\/src\/slots\/(.+)\.vue$/,
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

function normalizePluginKeys(item: SystemPlugin): string[] {
  const keys = [item.id];
  if (item.statusKey?.startsWith('plugin:')) {
    keys.push(item.statusKey.substring('plugin:'.length));
  }
  return keys.filter((key): key is string => !!key);
}

async function loadPluginStateMap(force = false) {
  if (!pluginStatePromise || force) {
    pluginStatePromise = pluginList({ pageNum: 1, pageSize: 1000 })
      .then(({ items }) => {
        const map = new Map<string, SystemPlugin>();
        for (const item of items) {
          for (const key of normalizePluginKeys(item)) {
            map.set(key, item);
          }
        }
        return map;
      })
      .catch((error) => {
        console.error('[plugin-slot] failed to load plugin state map', error);
        return new Map<string, SystemPlugin>();
      });
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
export function notifyPluginRegistryChanged() {
  pluginStatePromise = null;
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new Event(pluginRegistryChangedEvent));
  }
}

/**
 * Subscribes to plugin registry changes.
 */
export function onPluginRegistryChanged(listener: () => void) {
  if (typeof window === 'undefined') {
    return () => {};
  }
  window.addEventListener(pluginRegistryChangedEvent, listener);
  return () => {
    window.removeEventListener(pluginRegistryChangedEvent, listener);
  };
}
