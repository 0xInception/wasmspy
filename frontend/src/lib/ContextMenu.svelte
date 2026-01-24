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
  class="fixed z-50 min-w-40 py-1 bg-gray-800 border border-gray-700 rounded shadow-xl text-sm"
  style="left: {x}px; top: {y}px"
  onclick={(e) => e.stopPropagation()}
  role="menu"
>
  {#each items as item}
    {#if item.separator}
      <div class="my-1 border-t border-gray-700"></div>
    {:else}
      <button
        class="block w-full px-3 py-1 text-left {item.disabled ? 'text-gray-600 cursor-default' : 'text-gray-200 hover:bg-gray-700'}"
        onclick={() => handleClick(item)}
        disabled={item.disabled}
        role="menuitem"
      >
        {item.label}
      </button>
    {/if}
  {/each}
</div>
