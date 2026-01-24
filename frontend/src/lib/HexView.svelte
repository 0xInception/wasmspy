<script lang="ts">
  import { EditorState, RangeSetBuilder, StateField, Prec, type RangeSet } from '@codemirror/state';
  import { EditorView, lineNumbers, keymap, Decoration, drawSelection } from '@codemirror/view';
  import { defaultKeymap, selectAll } from '@codemirror/commands';
  import { search, searchKeymap } from '@codemirror/search';
  import { GetMemoryData } from '../../wailsjs/go/main/App';
  import ContextMenu, { type MenuItem } from './ContextMenu.svelte';
  import { getTheme } from './themes';

  let { modulePath, memIndex, initialAddress }: { modulePath: string; memIndex: number; initialAddress?: number | null } = $props();
  let contextMenu: { x: number; y: number; items: MenuItem[] } | null = $state(null);

  let container: HTMLDivElement;
  let view: EditorView | null = null;
  let content = $state('');
  let loading = $state(true);
  let gotoAddr = $state('');

  const BYTES_PER_ROW = 16;
  const NO_VIRTUALIZATION_THRESHOLD = 1500;

  function getLineCount(text: string): number {
    let count = 1;
    for (let i = 0; i < text.length; i++) {
      if (text[i] === '\n') count++;
    }
    return count;
  }

  function createHexTheme() {
    const theme = getTheme();
    const c = theme.colors;
    return EditorView.theme({
      '&': {
        height: '100%',
        fontSize: '13px',
        backgroundColor: c.editorBg,
        color: c.editorFg,
      },
      '.cm-scroller': {
        overflow: 'auto',
        fontFamily: 'var(--font-mono)',
        willChange: 'transform',
        transform: 'translateZ(0)',
        backfaceVisibility: 'hidden',
        overscrollBehavior: 'contain',
      },
      '.cm-content': {
        padding: '8px 0',
        lineHeight: '1.4',
        contain: 'layout style',
      },
      '.cm-line': {
        padding: '0 8px',
        contain: 'layout style',
        lineHeight: '1.4',
      },
      '.cm-gutters': {
        backgroundColor: c.editorGutter,
        border: 'none',
      },
      '.cm-lineNumbers .cm-gutterElement': { color: c.editorGutterFg, minWidth: '3em' },
      '.cm-activeLine': { backgroundColor: c.editorLineHighlight },
      '.cm-activeLineGutter': { backgroundColor: c.editorLineHighlight },
      '.hex-addr': { color: c.editorGutterFg },
      '.hex-byte': { color: c.number },
      '.hex-ascii': { color: c.string },
    }, { dark: true });
  }

  const addrMark = Decoration.mark({ class: 'hex-addr' });
  const byteMark = Decoration.mark({ class: 'hex-byte' });
  const asciiMark = Decoration.mark({ class: 'hex-ascii' });

  function buildAllDecorations(state: EditorState): RangeSet<Decoration> {
    const builder = new RangeSetBuilder<Decoration>();
    const doc = state.doc;
    for (let i = 1; i <= doc.lines; i++) {
      const line = doc.line(i);
      const len = line.length;
      if (len >= 8) {
        builder.add(line.from, line.from + 8, addrMark);
      }
      if (len >= 58) {
        builder.add(line.from + 10, line.from + 33, byteMark);
        builder.add(line.from + 35, line.from + 58, byteMark);
      }
      const text = line.text;
      const pipeIdx = text.lastIndexOf('|');
      if (pipeIdx > 0 && text.endsWith('|')) {
        builder.add(line.from + pipeIdx, line.to, asciiMark);
      }
    }
    return builder.finish();
  }

  const hexDecorations = StateField.define<RangeSet<Decoration>>({
    create(state) { return buildAllDecorations(state); },
    update(decos, tr) { return tr.docChanged ? buildAllDecorations(tr.state) : decos; },
    provide: f => EditorView.decorations.from(f),
  });

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
      const result = await GetMemoryData(modulePath, memIndex, offset, 1000000);
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

  // Extract only hex byte values from hex dump text
  // Format: "XXXXXXXX  HH HH HH HH HH HH HH HH  HH HH HH HH HH HH HH HH  |ASCII|"
  // Hex bytes are at positions 10-32 and 35-57
  function extractHexBytes(text: string): string {
    const lines = text.split('\n');
    const allBytes: string[] = [];

    for (const line of lines) {
      if (line.length < 10) continue;

      // Extract hex sections
      const hex1 = line.substring(10, 33);
      const hex2 = line.length >= 35 ? line.substring(35, 58) : '';

      // Parse individual bytes (filter valid 2-char hex values)
      const bytes = (hex1 + ' ' + hex2)
        .split(/\s+/)
        .filter(b => /^[0-9a-f]{2}$/i.test(b));

      allBytes.push(...bytes);
    }

    return allBytes.join('');
  }

  function copyHexBytes() {
    if (!view) return false;
    const selection = view.state.selection.main;
    if (selection.from === selection.to) return false;
    const selectedText = view.state.sliceDoc(selection.from, selection.to);
    const hexBytes = extractHexBytes(selectedText);
    if (hexBytes) {
      navigator.clipboard.writeText(hexBytes);
    }
    return true;
  }

  function handleContextMenu(e: MouseEvent) {
    e.preventDefault();
    const selection = view?.state.selection.main;
    const hasSelection = selection && selection.from !== selection.to;
    const selectedText = hasSelection ? view?.state.sliceDoc(selection.from, selection.to) : null;

    const items: MenuItem[] = [];
    if (selectedText) {
      const hexBytes = extractHexBytes(selectedText);
      if (hexBytes) {
        items.push({ label: 'Copy hex', action: () => navigator.clipboard.writeText(hexBytes) });
      }
      items.push({ label: 'Copy raw', action: () => navigator.clipboard.writeText(selectedText) });
    }
    items.push({ label: 'Copy all hex', action: () => navigator.clipboard.writeText(extractHexBytes(content)) });
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

    const lineCount = getLineCount(content);
    const disableVirtualization = lineCount < NO_VIRTUALIZATION_THRESHOLD;

    const extensions = [
      Prec.high(keymap.of([
        { key: 'Mod-a', run: selectAll },
        { key: 'Mod-c', run: () => copyHexBytes() },
        ...defaultKeymap,
        ...searchKeymap,
      ])),
      lineNumbers(),
      drawSelection(),
      createHexTheme(),
      hexDecorations,
      EditorState.readOnly.of(true),
      search(),
    ];

    if (disableVirtualization) {
      extensions.push(
        EditorView.scrollMargins.of(() => ({ top: 1000000, bottom: 1000000 }))
      );
    }

    view = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions,
      }),
      parent: container,
    });
  });
</script>

<div class="h-full flex flex-col">
  <div class="flex items-center gap-2 p-2 text-xs" style="background: var(--panel-bg); border-bottom: 1px solid var(--panel-border); color: var(--sidebar-fg);">
    <span style="opacity: 0.7;">Go to:</span>
    <input
      type="text"
      bind:value={gotoAddr}
      onkeydown={(e) => e.key === 'Enter' && goToAddress()}
      placeholder="hex address"
      class="w-24 px-2 py-1 rounded font-mono"
      style="background: var(--input-bg); border: 1px solid var(--input-border); color: var(--editor-fg);"
    />
    <button onclick={goToAddress} class="px-2 py-1 rounded" style="background: var(--button-bg); color: var(--sidebar-fg);">Go</button>
  </div>
  <div bind:this={container} class="flex-1 overflow-hidden" oncontextmenu={handleContextMenu}>
    {#if loading}
      <div class="p-4" style="color: var(--sidebar-fg); opacity: 0.5;">Loading...</div>
    {/if}
  </div>
</div>

{#if contextMenu}
  <ContextMenu items={contextMenu.items} x={contextMenu.x} y={contextMenu.y} onClose={() => contextMenu = null} />
{/if}
