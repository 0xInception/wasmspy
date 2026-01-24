<script lang="ts" generics="T">
  import type { Snippet } from 'svelte';

  let { items, itemHeight = 24, overscan = 5, renderItem }: {
    items: T[];
    itemHeight?: number;
    overscan?: number;
    renderItem: Snippet<[T, number]>;
  } = $props();

  let container: HTMLDivElement;
  let scrollTop = $state(0);
  let containerHeight = $state(0);

  let totalHeight = $derived(items.length * itemHeight);
  let startIndex = $derived(Math.max(0, Math.floor(scrollTop / itemHeight) - overscan));
  let endIndex = $derived(Math.min(items.length, Math.ceil((scrollTop + containerHeight) / itemHeight) + overscan));
  let visibleItems = $derived(items.slice(startIndex, endIndex).map((item, i) => ({ item, index: startIndex + i })));
  let offsetY = $derived(startIndex * itemHeight);

  function onScroll(e: Event) {
    const target = e.target as HTMLDivElement;
    scrollTop = target.scrollTop;
  }

  function onResize() {
    if (container) containerHeight = container.clientHeight;
  }

  $effect(() => {
    if (container) containerHeight = container.clientHeight;
  });
</script>

<svelte:window onresize={onResize} />

<div bind:this={container} class="h-full overflow-auto" onscroll={onScroll}>
  <div style="height: {totalHeight}px; position: relative;">
    <div style="transform: translateY({offsetY}px);">
      {#each visibleItems as { item, index } (index)}
        <div style="height: {itemHeight}px;">
          {@render renderItem(item, index)}
        </div>
      {/each}
    </div>
  </div>
</div>
