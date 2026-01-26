<script lang="ts">
  import { type Settings, type WindowSize, defaultSettings, windowSizePresets } from './settings';

  let {
    settings,
    onSave,
    onClose,
  }: {
    settings: Settings;
    onSave: (settings: Settings) => void;
    onClose: () => void;
  } = $props();

  let localSettings = $state({ ...settings });

  function handleSave() {
    onSave(localSettings);
    onClose();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onClose();
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) handleSave();
  }

  function resetToDefaults() {
    localSettings = { ...defaultSettings };
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="fixed inset-0 flex items-center justify-center z-50" role="dialog">
  <div class="absolute inset-0 bg-black/50" onclick={onClose}></div>
  <div class="relative rounded-lg shadow-xl w-[480px]" style="background: var(--sidebar-bg); border: 1px solid var(--panel-border);">
    <div class="flex items-center justify-between px-4 py-3" style="border-bottom: 1px solid var(--panel-border);">
      <span class="font-medium" style="color: var(--sidebar-fg);">Settings</span>
      <button class="opacity-60 hover:opacity-100 text-lg leading-none" style="color: var(--sidebar-fg);" onclick={onClose}>×</button>
    </div>

    <div class="p-4 space-y-4">
      <div>
        <label class="block text-xs mb-1.5" style="color: var(--sidebar-fg); opacity: 0.7;">
          Virtualization threshold (lines)
        </label>
        <input
          type="number"
          min="500"
          max="10000"
          step="500"
          bind:value={localSettings.virtualizationThreshold}
          class="w-full px-3 py-2 text-sm rounded outline-none"
          style="background: var(--input-bg); border: 1px solid var(--input-border); color: var(--editor-fg);"
        />
        <p class="text-xs mt-1" style="color: var(--sidebar-fg); opacity: 0.5;">
          Use fast rendering for files under this line count. Higher = smoother but more memory.
        </p>
      </div>

      <div>
        <label class="block text-xs mb-1.5" style="color: var(--sidebar-fg); opacity: 0.7;">
          Font size (px)
        </label>
        <input
          type="number"
          min="10"
          max="24"
          step="1"
          bind:value={localSettings.fontSize}
          class="w-full px-3 py-2 text-sm rounded outline-none"
          style="background: var(--input-bg); border: 1px solid var(--input-border); color: var(--editor-fg);"
        />
      </div>

      <div>
        <label class="block text-xs mb-2" style="color: var(--sidebar-fg); opacity: 0.7;">
          Default panels
        </label>
        <div class="space-y-2">
          <label class="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" bind:checked={localSettings.defaultShowDecompile} class="w-4 h-4" />
            <span class="text-sm" style="color: var(--sidebar-fg);">Show Decompile panel</span>
          </label>
          <label class="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" bind:checked={localSettings.defaultShowDisassembly} class="w-4 h-4" />
            <span class="text-sm" style="color: var(--sidebar-fg);">Show Disassembly panel</span>
          </label>
        </div>
      </div>

      <div>
        <label class="block text-xs mb-1.5" style="color: var(--sidebar-fg); opacity: 0.7;">
          Default window size
        </label>
        <select
          bind:value={localSettings.windowSize}
          class="w-full px-3 py-2 text-sm rounded outline-none"
          style="background: var(--input-bg); border: 1px solid var(--input-border); color: var(--editor-fg);"
        >
          <option value="4k">4K (3840 × 2160)</option>
          <option value="2k">2K (2560 × 1440)</option>
          <option value="1080p">1080p (1920 × 1080)</option>
          <option value="720p">720p (1280 × 720)</option>
        </select>
        <p class="text-xs mt-1" style="color: var(--sidebar-fg); opacity: 0.5;">
          Applied on next app launch
        </p>
      </div>
    </div>

    <div class="flex items-center justify-between px-4 py-3" style="border-top: 1px solid var(--panel-border);">
      <button
        class="px-3 py-1.5 text-xs rounded"
        style="color: var(--sidebar-fg); opacity: 0.7;"
        onclick={resetToDefaults}
      >Reset to defaults</button>
      <div class="flex gap-2">
        <button
          class="px-3 py-1.5 text-xs rounded"
          style="background: var(--panel-bg); color: var(--sidebar-fg); border: 1px solid var(--panel-border);"
          onclick={onClose}
        >Cancel</button>
        <button
          class="px-3 py-1.5 text-xs rounded"
          style="background: var(--button-active); color: white;"
          onclick={handleSave}
        >Save</button>
      </div>
    </div>
  </div>
</div>
