<script lang="ts">
  import { EditorState } from '@codemirror/state';
  import { EditorView, lineNumbers, keymap } from '@codemirror/view';
  import { HighlightStyle, syntaxHighlighting, StreamLanguage } from '@codemirror/language';
  import { tags } from '@lezer/highlight';
  import { search, searchKeymap } from '@codemirror/search';
  import { GetMemoryData } from '../../wailsjs/go/main/App';
  import ContextMenu, { type MenuItem } from './ContextMenu.svelte';

  let { modulePath, memIndex, initialAddress }: { modulePath: string; memIndex: number; initialAddress?: number | null } = $props();
  let contextMenu: { x: number; y: number; items: MenuItem[] } | null = $state(null);

  let container: HTMLDivElement;
  let view: EditorView | null = null;
  let content = $state('');
  let loading = $state(true);
  let segments: { offset: number; size: number }[] = $state([]);
  let gotoAddr = $state('');

  const BYTES_PER_ROW = 16;

  const hexLanguage = StreamLanguage.define({
    token(stream) {
      if (stream.sol() && stream.match(/[0-9a-f]+/i)) return 'lineNumber';
      if (stream.match(/\|.*\|/)) return 'string';
      if (stream.match(/[0-9a-f]{2}/i)) return 'number';
      stream.next();
      return null;
    }
  });

  const hexTheme = EditorView.theme({
    '&': { height: '100%', fontSize: '13px', backgroundColor: '#0d1117' },
    '.cm-scroller': { overflow: 'auto' },
    '.cm-content': { fontFamily: 'ui-monospace, monospace', padding: '8px 0' },
    '.cm-line': { padding: '0 8px' },
    '.cm-gutters': { backgroundColor: '#0d1117', border: 'none' },
    '.cm-lineNumbers .cm-gutterElement': { color: '#484f58', minWidth: '3em' },
    '.cm-activeLine': { backgroundColor: '#161b2280' },
    '.cm-activeLineGutter': { backgroundColor: '#161b2280' },
  }, { dark: true });

  const hexHighlight = HighlightStyle.define([
    { tag: tags.lineComment, color: '#6e7681' },
    { tag: tags.number, color: '#79c0ff' },
    { tag: tags.string, color: '#56d364' },
  ]);

  function toHex(b: number): string {
    return b.toString(16).padStart(2, '0');
  }

  function toAscii(b: number): string {
    return b >= 0x20 && b < 0x7f ? String.fromCharCode(b) : '.';
  }

  function decodeData(data: unknown): number[] {
    if (typeof data === 'string') {
      const binary = atob(data);
      const bytes = new Array(binary.length);
      for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
      return bytes;
    }
    return Array.from(data as ArrayLike<number>);
  }

  function formatHexDump(data: number[], startOffset: number): string {
    const lines: string[] = [];
    for (let i = 0; i < data.length; i += BYTES_PER_ROW) {
      const addr = startOffset + i;
      const row = data.slice(i, i + BYTES_PER_ROW);
      const hex1 = row.slice(0, 8).map(toHex).join(' ');
      const hex2 = row.slice(8, 16).map(toHex).join(' ');
      const ascii = row.map(toAscii).join('');
      lines.push(`${addr.toString(16).padStart(8, '0')}  ${hex1.padEnd(23)}  ${hex2.padEnd(23)}  |${ascii}|`);
    }
    return lines.join('\n');
  }

  async function loadData(offset: number = 0) {
    if (!modulePath) return;
    loading = true;
    try {
      const result = await GetMemoryData(modulePath, memIndex, offset, 100000);
      segments = result.segments || [];
      const data = decodeData(result.data);
      content = data.length === 0 ? '; No data at this offset' : formatHexDump(data, offset);
    } catch (e) {
      content = '; Failed to load memory: ' + e;
    } finally {
      loading = false;
    }
  }

  function goToAddress() {
    const addr = parseInt(gotoAddr, 16);
    if (!isNaN(addr)) loadData(addr);
  }

  function handleContextMenu(e: MouseEvent) {
    e.preventDefault();
    const selection = view?.state.selection.main;
    const hasSelection = selection && selection.from !== selection.to;
    const selectedText = hasSelection ? view?.state.sliceDoc(selection.from, selection.to) : null;

    const items: MenuItem[] = [];
    if (selectedText) {
      items.push({ label: 'Copy', action: () => navigator.clipboard.writeText(selectedText) });
    }
    items.push({ label: 'Copy all', action: () => navigator.clipboard.writeText(content) });
    contextMenu = { x: e.clientX, y: e.clientY, items };
  }

  $effect(() => {
    const addr = initialAddress ?? 0;
    if (modulePath) {
      gotoAddr = addr > 0 ? addr.toString(16) : '';
      loadData(addr);
    }
  });

  $effect(() => {
    if (!container || loading) return;
    if (view) view.destroy();
    view = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions: [
          lineNumbers(),
          hexTheme,
          hexLanguage,
          syntaxHighlighting(hexHighlight),
          EditorState.readOnly.of(true),
          search(),
          keymap.of(searchKeymap),
        ],
      }),
      parent: container,
    });
  });
</script>

<div class="h-full flex flex-col">
  <div class="flex items-center gap-2 p-2 border-b border-gray-700 bg-gray-900 text-xs">
    <span class="text-gray-400">Go to:</span>
    <input
      type="text"
      bind:value={gotoAddr}
      onkeydown={(e) => e.key === 'Enter' && goToAddress()}
      placeholder="hex address"
      class="w-24 px-2 py-1 bg-gray-800 border border-gray-700 rounded text-gray-300 font-mono"
    />
    <button onclick={goToAddress} class="px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded">Go</button>
    {#if segments.length > 0}
      <span class="text-gray-500 ml-2">|</span>
      <span class="text-gray-400">Segments:</span>
      {#each segments.slice(0, 5) as seg}
        <button onclick={() => loadData(seg.offset)} class="px-2 py-1 bg-gray-800 hover:bg-gray-700 rounded text-cyan-400">
          {seg.offset.toString(16)}
        </button>
      {/each}
      {#if segments.length > 5}
        <span class="text-gray-500">+{segments.length - 5} more</span>
      {/if}
    {/if}
  </div>
  <div bind:this={container} class="flex-1 overflow-hidden" oncontextmenu={handleContextMenu}>
    {#if loading}
      <div class="p-4 text-gray-500">Loading...</div>
    {/if}
  </div>
</div>

{#if contextMenu}
  <ContextMenu items={contextMenu.items} x={contextMenu.x} y={contextMenu.y} onClose={() => contextMenu = null} />
{/if}
