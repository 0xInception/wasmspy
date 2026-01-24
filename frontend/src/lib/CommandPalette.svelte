<script lang="ts">
  import type { FunctionInfo } from './types';

  let { functions, onSelect, onClose }: {
    functions: FunctionInfo[];
    onSelect: (fn: FunctionInfo) => void;
    onClose: () => void;
  } = $props();

  let query = $state('');
  let selectedIndex = $state(0);
  let input: HTMLInputElement;

  let filtered = $derived(() => {
    if (!query) return functions.slice(0, 100);
    const q = query.toLowerCase();
    return functions.filter(fn => fn.name.toLowerCase().includes(q)).slice(0, 100);
  });

  let results = $derived(filtered());

  $effect(() => {
    query;
    selectedIndex = 0;
  });

  function handleKeydown(e: KeyboardEvent) {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        selectedIndex = Math.min(selectedIndex + 1, results.length - 1);
        break;
      case 'ArrowUp':
        e.preventDefault();
        selectedIndex = Math.max(selectedIndex - 1, 0);
        break;
      case 'Enter':
        e.preventDefault();
        if (results[selectedIndex]) {
          onSelect(results[selectedIndex]);
          onClose();
        }
        break;
      case 'Escape':
        e.preventDefault();
        onClose();
        break;
    }
  }

  function handleSelect(fn: FunctionInfo) {
    onSelect(fn);
    onClose();
  }
</script>

<div class="fixed inset-0 bg-black/60 flex items-start justify-center pt-24 z-50" onclick={onClose} role="dialog">
  <div class="w-[600px] rounded-lg shadow-2xl overflow-hidden" style="background: var(--panel-bg); border: 1px solid var(--panel-border);" onclick={(e) => e.stopPropagation()}>
    <input
      bind:this={input}
      bind:value={query}
      onkeydown={handleKeydown}
      placeholder="Search functions..."
      class="w-full px-4 py-3 outline-none"
      style="background: var(--input-bg); border-bottom: 1px solid var(--panel-border); color: var(--editor-fg);"
      autofocus
    />
    <div class="max-h-80 overflow-auto">
      {#each results as fn, i (fn.index)}
        <button
          class="flex items-center gap-2 w-full px-4 py-2 text-left text-sm"
          style="background: {i === selectedIndex ? 'var(--sidebar-active)' : 'transparent'}; color: var(--sidebar-fg);"
          onclick={() => handleSelect(fn)}
          onmouseenter={() => selectedIndex = i}
        >
          <span style="color: {fn.imported ? 'var(--icon-import)' : 'var(--icon-function)'};">f</span>
          <span class="truncate">{fn.name}</span>
          <span class="text-xs ml-auto" style="opacity: 0.5;">#{fn.index}</span>
        </button>
      {/each}
      {#if results.length === 0}
        <div class="px-4 py-8 text-center" style="opacity: 0.5;">No functions found</div>
      {/if}
    </div>
  </div>
</div>
