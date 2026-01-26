<script lang="ts">
  let { currentName, onConfirm, onCancel }: {
    currentName: string;
    onConfirm: (newName: string) => void;
    onCancel: () => void;
  } = $props();

  let inputValue = $state(currentName);
  let inputEl: HTMLInputElement;

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && inputValue.trim()) {
      onConfirm(inputValue.trim());
    } else if (e.key === 'Escape') {
      onCancel();
    }
  }

  function handleSubmit() {
    if (inputValue.trim() && inputValue.trim() !== currentName) {
      onConfirm(inputValue.trim());
    } else {
      onCancel();
    }
  }

  $effect(() => {
    inputEl?.focus();
    inputEl?.select();
  });
</script>

<div class="fixed inset-0 flex items-center justify-center z-50" onclick={onCancel} onkeydown={handleKeydown} role="dialog">
  <div class="absolute inset-0 bg-black/50"></div>
  <div class="relative p-4 rounded shadow-lg min-w-[300px]" style="background: var(--panel-bg); border: 1px solid var(--panel-border);" onclick={(e) => e.stopPropagation()} role="document">
    <div class="text-sm mb-3" style="color: var(--sidebar-fg);">Add nickname</div>
    <input
      bind:this={inputEl}
      bind:value={inputValue}
      type="text"
      class="w-full px-2 py-1.5 text-sm rounded outline-none"
      style="background: var(--editor-bg); color: var(--editor-fg); border: 1px solid var(--panel-border);"
      onkeydown={handleKeydown}
    />
    <div class="flex justify-end gap-2 mt-3">
      <button
        class="px-3 py-1 text-sm rounded"
        style="background: var(--sidebar-bg); color: var(--sidebar-fg);"
        onclick={onCancel}
      >Cancel</button>
      <button
        class="px-3 py-1 text-sm rounded"
        style="background: var(--button-active); color: var(--editor-fg);"
        onclick={handleSubmit}
      >Save</button>
    </div>
  </div>
</div>
