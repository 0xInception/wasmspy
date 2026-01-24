<script lang="ts">
  export interface MenuItem {
    label: string;
    action: () => void;
    separator?: boolean;
    disabled?: boolean;
  }

  let { items, x, y, onClose }: { items: MenuItem[]; x: number; y: number; onClose: () => void } = $props();

  function handleClick(item: MenuItem) {
    if (!item.disabled) {
      item.action();
      onClose();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onClose();
  }
</script>

<svelte:window onkeydown={handleKeydown} onclick={onClose} />

<div
  class="fixed z-50 min-w-40 py-1 rounded shadow-xl text-sm"
  style="left: {x}px; top: {y}px; background: var(--panel-bg); border: 1px solid var(--panel-border);"
  onclick={(e) => e.stopPropagation()}
  role="menu"
>
  {#each items as item}
    {#if item.separator}
      <div class="my-1" style="border-top: 1px solid var(--panel-border);"></div>
    {:else}
      <button
        class="block w-full px-3 py-1 text-left"
        style="color: {item.disabled ? 'var(--sidebar-fg)' : 'var(--sidebar-fg)'}; opacity: {item.disabled ? 0.4 : 1};"
        onclick={() => handleClick(item)}
        onmouseenter={(e) => !item.disabled && ((e.target as HTMLElement).style.background = 'var(--sidebar-hover)')}
        onmouseleave={(e) => (e.target as HTMLElement).style.background = 'transparent'}
        disabled={item.disabled}
        role="menuitem"
      >
        {item.label}
      </button>
    {/if}
  {/each}
</div>
