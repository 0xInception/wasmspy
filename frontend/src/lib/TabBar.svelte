<script lang="ts">
  import ContextMenu, { type MenuItem } from './ContextMenu.svelte';

  export interface Tab {
    id: string;
    title: string;
    icon?: string;
  }

  let { tabs, activeId, onSelect, onClose, onCloseAll, onCloseOthers, onCloseToLeft, onCloseToRight }: {
    tabs: Tab[];
    activeId: string | null;
    onSelect: (id: string) => void;
    onClose: (id: string) => void;
    onCloseAll?: () => void;
    onCloseOthers?: (id: string) => void;
    onCloseToLeft?: (id: string) => void;
    onCloseToRight?: (id: string) => void;
  } = $props();

  let contextMenu: { x: number; y: number; items: MenuItem[] } | null = $state(null);

  function handleClose(e: MouseEvent, id: string) {
    e.stopPropagation();
    onClose(id);
  }

  function handleAuxClick(e: MouseEvent, id: string) {
    if (e.button === 1) {
      e.preventDefault();
      onClose(id);
    }
  }

  function handleContextMenu(e: MouseEvent, tab: Tab, index: number) {
    e.preventDefault();
    const items: MenuItem[] = [
      { label: 'Close', action: () => onClose(tab.id) },
    ];
    if (onCloseOthers && tabs.length > 1) {
      items.push({ label: 'Close others', action: () => onCloseOthers(tab.id) });
    }
    if (onCloseToLeft && index > 0) {
      items.push({ label: 'Close to the left', action: () => onCloseToLeft(tab.id) });
    }
    if (onCloseToRight && index < tabs.length - 1) {
      items.push({ label: 'Close to the right', action: () => onCloseToRight(tab.id) });
    }
    if (onCloseAll && tabs.length > 1) {
      items.push({ label: 'Close all', action: () => onCloseAll() });
    }
    contextMenu = { x: e.clientX, y: e.clientY, items };
  }
</script>

<div class="flex items-center" style="background: var(--tab-bg); border-bottom: 1px solid var(--tab-border);">
  <div class="flex overflow-x-auto flex-1">
    {#each tabs as tab, index (tab.id)}
      {@const isActive = activeId === tab.id}
      <button
        class="flex items-center gap-2 px-3 py-1.5 text-xs min-w-0 max-w-48 group"
        style="
          background: {isActive ? 'var(--tab-active-bg)' : 'transparent'};
          color: {isActive ? 'var(--tab-active-fg)' : 'var(--tab-fg)'};
          border-right: 1px solid var(--tab-border);
        "
        onclick={() => onSelect(tab.id)}
        onauxclick={(e) => handleAuxClick(e, tab.id)}
        oncontextmenu={(e) => handleContextMenu(e, tab, index)}
      >
        {#if tab.icon}
          <span class="flex-shrink-0" style="color: {tab.icon === 'f' ? 'var(--icon-function)' : tab.icon === 'm' ? 'var(--icon-memory)' : 'inherit'};">{tab.icon}</span>
        {/if}
        <span class="truncate">{tab.title}</span>
        <span
          class="flex-shrink-0 w-4 h-4 flex items-center justify-center rounded {isActive ? 'opacity-100' : 'opacity-0'} group-hover:opacity-100"
          style="background: transparent;"
          onclick={(e) => handleClose(e, tab.id)}
          onmouseenter={(e) => (e.target as HTMLElement).style.background = 'var(--button-hover)'}
          onmouseleave={(e) => (e.target as HTMLElement).style.background = 'transparent'}
        >×</span>
      </button>
    {/each}
  </div>
  {#if tabs.length > 5}
    <div class="px-2 text-xs flex-shrink-0" style="color: var(--tab-fg); opacity: 0.6;">
      {tabs.length} tabs
    </div>
  {/if}
</div>

{#if contextMenu}
  <ContextMenu items={contextMenu.items} x={contextMenu.x} y={contextMenu.y} onClose={() => contextMenu = null} />
{/if}
