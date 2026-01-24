<script lang="ts">
  import { onMount } from 'svelte';
  import Editor from './Editor.svelte';

  let {
    leftContent,
    rightContent,
    showLeft = true,
    showRight = true,
    showOffsets = false,
    disasmIndent = false,
    onCloseLeft,
    onCloseRight,
    onShowLeft,
    onShowRight,
    onToggleOffsets,
    onToggleDisasmIndent,
    onGotoAddress,
    onGotoFunction,
    onShowXRefs,
    functions,
    functionsByName,
    onLeftSelectionChange,
    onRightSelectionChange,
    leftHighlightLines,
    rightHighlightLines,
  }: {
    leftContent: string | null;
    rightContent: string | null;
    showLeft?: boolean;
    showRight?: boolean;
    showOffsets?: boolean;
    disasmIndent?: boolean;
    onCloseLeft?: () => void;
    onCloseRight?: () => void;
    onShowLeft?: () => void;
    onShowRight?: () => void;
    onToggleOffsets?: () => void;
    onToggleDisasmIndent?: () => void;
    onGotoAddress?: (addr: number) => void;
    onGotoFunction?: (index: number) => void;
    onShowXRefs?: (index: number) => void;
    functions?: { index: number; name: string }[];
    functionsByName?: Map<string, { index: number; name: string }>;
    onLeftSelectionChange?: (startLine: number, endLine: number) => void;
    onRightSelectionChange?: (startLine: number, endLine: number) => void;
    leftHighlightLines?: number[] | null;
    rightHighlightLines?: number[] | null;
  } = $props();

  let splitRatio = $state(0.5);
  let dragging = $state(false);
  let container: HTMLDivElement;

  onMount(() => {
    const saved = localStorage.getItem('editorSplitRatio');
    if (saved) splitRatio = Math.max(0.2, Math.min(0.8, parseFloat(saved)));
  });

  function startDrag(e: MouseEvent) {
    e.preventDefault();
    dragging = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
  }

  function onMouseMove(e: MouseEvent) {
    if (!dragging || !container) return;
    const rect = container.getBoundingClientRect();
    const newRatio = (e.clientX - rect.left) / rect.width;
    splitRatio = Math.max(0.2, Math.min(0.8, newRatio));
  }

  function stopDrag() {
    if (!dragging) return;
    dragging = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    localStorage.setItem('editorSplitRatio', String(splitRatio));
  }

  function stripOffsetComments(code: string): string {
    return code.replace(/\s*\/\/\s*@0x[0-9a-fA-F]+$/gm, '');
  }

  let displayLeftContent = $derived(
    leftContent && !showOffsets ? stripOffsetComments(leftContent) : leftContent
  );
</script>

<svelte:window onmousemove={onMouseMove} onmouseup={stopDrag} />

<div bind:this={container} class="flex flex-col h-full">
  <div class="flex h-full">
    {#if showLeft && leftContent !== null}
      <div class="flex flex-col" style="width: {showRight ? `${splitRatio * 100}%` : '100%'}">
        <div class="flex items-center justify-between px-3 py-1.5 text-xs" style="background: var(--panel-bg); border-bottom: 1px solid var(--panel-border);">
          <div class="flex items-center gap-2">
            <span style="color: var(--sidebar-fg);">Decompile</span>
            <label class="flex items-center gap-1 cursor-pointer" style="color: var(--sidebar-fg); opacity: 0.7;">
              <input type="checkbox" checked={showOffsets} onchange={onToggleOffsets} class="w-3 h-3" />
              <span>offsets</span>
            </label>
          </div>
          {#if onCloseLeft && showRight}
            <button
              class="opacity-60 hover:opacity-100 px-1"
              style="color: var(--sidebar-fg);"
              onclick={onCloseLeft}
            >×</button>
          {/if}
        </div>
        <div class="flex-1 overflow-hidden">
          <Editor
            content={displayLeftContent || ''}
            mode="decompile"
            {onGotoAddress}
            {onGotoFunction}
            {onShowXRefs}
            {functions}
            {functionsByName}
            onSelectionChange={onLeftSelectionChange}
            highlightLines={leftHighlightLines}
            onShowDisassembly={!showRight ? onShowRight : undefined}
          />
        </div>
      </div>

      {#if showRight && rightContent !== null}
        <div
          class="w-1 cursor-col-resize flex-shrink-0 transition-colors"
          style="background: {dragging ? 'var(--button-active)' : 'var(--sidebar-border)'};"
          onmousedown={startDrag}
          onmouseenter={(e) => !dragging && ((e.target as HTMLElement).style.background = 'var(--button-active)')}
          onmouseleave={(e) => !dragging && ((e.target as HTMLElement).style.background = 'var(--sidebar-border)')}
          role="separator"
        ></div>
      {/if}
    {/if}

    {#if showRight && rightContent !== null}
      <div class="flex flex-col flex-1">
        <div class="flex items-center justify-between px-3 py-1.5 text-xs" style="background: var(--panel-bg); border-bottom: 1px solid var(--panel-border);">
          <div class="flex items-center gap-2">
            <span style="color: var(--sidebar-fg);">Disassembly</span>
            <label class="flex items-center gap-1 cursor-pointer" style="color: var(--sidebar-fg); opacity: 0.7;">
              <input type="checkbox" checked={disasmIndent} onchange={onToggleDisasmIndent} class="w-3 h-3" />
              <span>indent</span>
            </label>
          </div>
          {#if onCloseRight && showLeft}
            <button
              class="opacity-60 hover:opacity-100 px-1"
              style="color: var(--sidebar-fg);"
              onclick={onCloseRight}
            >×</button>
          {/if}
        </div>
        <div class="flex-1 overflow-hidden">
          <Editor
            content={rightContent}
            mode="disasm"
            {onGotoAddress}
            {onGotoFunction}
            {onShowXRefs}
            {functions}
            {functionsByName}
            onSelectionChange={onRightSelectionChange}
            highlightLines={rightHighlightLines}
            onShowDecompile={!showLeft ? onShowLeft : undefined}
          />
        </div>
      </div>
    {/if}
  </div>
</div>
