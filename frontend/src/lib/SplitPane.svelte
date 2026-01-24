<script lang="ts">
  import { onMount, type Snippet } from 'svelte';

  let { left, right, leftWidth = 288, minLeft = 150, maxLeft = 500, storageKey = 'splitPane' }: {
    left: Snippet;
    right: Snippet;
    leftWidth?: number;
    minLeft?: number;
    maxLeft?: number;
    storageKey?: string;
  } = $props();

  let width = $state(leftWidth);
  let dragging = $state(false);
  let container: HTMLDivElement;

  onMount(() => {
    const saved = localStorage.getItem(storageKey);
    if (saved) width = Math.max(minLeft, Math.min(maxLeft, parseInt(saved, 10)));
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
    const newWidth = e.clientX - rect.left;
    width = Math.max(minLeft, Math.min(maxLeft, newWidth));
  }

  function stopDrag() {
    if (!dragging) return;
    dragging = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    localStorage.setItem(storageKey, String(width));
  }
</script>

<svelte:window onmousemove={onMouseMove} onmouseup={stopDrag} />

<div bind:this={container} class="flex h-full">
  <div style="width: {width}px" class="flex-shrink-0 h-full overflow-hidden">
    {@render left()}
  </div>
  <div
    class="w-1 cursor-col-resize flex-shrink-0 transition-colors"
    style="background: {dragging ? 'var(--button-active)' : 'var(--sidebar-border)'};"
    onmousedown={startDrag}
    onmouseenter={(e) => !dragging && ((e.target as HTMLElement).style.background = 'var(--button-active)')}
    onmouseleave={(e) => !dragging && ((e.target as HTMLElement).style.background = 'var(--sidebar-border)')}
    role="separator"
  ></div>
  <div class="flex-1 h-full overflow-hidden">
    {@render right()}
  </div>
</div>
