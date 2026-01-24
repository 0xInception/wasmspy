<script lang="ts">
  export interface Tab {
    id: string;
    title: string;
    icon?: string;
  }

  let { tabs, activeId, onSelect, onClose }: {
    tabs: Tab[];
    activeId: string | null;
    onSelect: (id: string) => void;
    onClose: (id: string) => void;
  } = $props();

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
</script>

<div class="flex overflow-x-auto" style="background: var(--tab-bg); border-bottom: 1px solid var(--tab-border);">
  {#each tabs as tab (tab.id)}
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
