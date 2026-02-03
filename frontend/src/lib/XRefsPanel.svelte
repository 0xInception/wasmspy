<script lang="ts">
  import type { XRefInfo } from './types';

  let {
    xrefs,
    funcName,
    onClose,
    onGotoFunction,
  }: {
    xrefs: XRefInfo;
    funcName: string;
    onClose: () => void;
    onGotoFunction: (index: number) => void;
  } = $props();
</script>

<div class="w-64 flex-shrink-0 text-xs overflow-auto" style="background: var(--sidebar-bg); border-left: 1px solid var(--panel-border);">
  <div class="flex items-center justify-between px-3 py-2" style="border-bottom: 1px solid var(--panel-border);">
    <span style="color: var(--sidebar-fg);">References</span>
    <button class="opacity-60 hover:opacity-100" style="color: var(--sidebar-fg);" onclick={onClose}>×</button>
  </div>
  <div class="px-3 py-2" style="color: var(--sidebar-fg);">
    <div class="truncate mb-3" style="color: var(--syntax-function);">{funcName}</div>
    {#if xrefs.callers.length > 0}
      <div class="mb-3">
        <div class="opacity-60 mb-1">Called by ({xrefs.callers.length})</div>
        {#each xrefs.callers as caller}
          <button
            class="block hover:underline truncate text-left w-full py-0.5"
            style="color: var(--syntax-function);"
            onclick={() => onGotoFunction(caller.index)}
          >{caller.name}</button>
        {/each}
      </div>
    {/if}
    {#if xrefs.callees.length > 0}
      <div>
        <div class="opacity-60 mb-1">Calls ({xrefs.callees.length})</div>
        {#each xrefs.callees as callee}
          <button
            class="block hover:underline truncate text-left w-full py-0.5"
            style="color: var(--syntax-function);"
            onclick={() => onGotoFunction(callee.index)}
          >{callee.name}</button>
        {/each}
      </div>
    {/if}
    {#if xrefs.callers.length === 0 && xrefs.callees.length === 0}
      <div class="opacity-60">No references found</div>
    {/if}
  </div>
</div>
