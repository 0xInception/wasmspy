<script lang="ts">
  let {
    onCancel,
    onQuitWithoutSave,
    onSaveAndQuit,
  }: {
    onCancel: () => void;
    onQuitWithoutSave: () => void;
    onSaveAndQuit: () => void;
  } = $props();
</script>

<div class="fixed inset-0 flex items-center justify-center z-50" style="background: rgba(0,0,0,0.5);">
  <div class="p-4 rounded-lg shadow-xl" style="background: var(--sidebar-bg); border: 1px solid var(--panel-border); min-width: 320px;">
    <div class="text-sm mb-4" style="color: var(--sidebar-fg);">
      You have unsaved changes. Do you want to save before quitting?
    </div>
    <div class="flex justify-end gap-2">
      <button
        class="px-3 py-1.5 text-xs rounded"
        style="background: var(--panel-bg); color: var(--sidebar-fg); border: 1px solid var(--panel-border);"
        onclick={onCancel}
      >Cancel</button>
      <button
        class="px-3 py-1.5 text-xs rounded"
        style="background: var(--color-error); color: white;"
        onclick={onQuitWithoutSave}
      >Don't Save</button>
      <button
        class="px-3 py-1.5 text-xs rounded"
        style="background: var(--button-active); color: white;"
        onclick={onSaveAndQuit}
      >Save & Quit</button>
    </div>
  </div>
</div>
