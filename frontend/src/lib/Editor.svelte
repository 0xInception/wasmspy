<script lang="ts">
  import { EditorState } from '@codemirror/state';
  import { EditorView, lineNumbers, keymap } from '@codemirror/view';
  import { syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { search, searchKeymap } from '@codemirror/search';
  import { wat } from '../lang-wat';
  import ContextMenu, { type MenuItem } from './ContextMenu.svelte';

  let { content, onGotoAddress }: { content: string; onGotoAddress?: (addr: number) => void } = $props();

  let container: HTMLDivElement;
  let view: EditorView | null = null;
  let contextMenu: { x: number; y: number; items: MenuItem[] } | null = $state(null);

  function getNumberAtPosition(x: number, y: number): number | null {
    if (!view) return null;
    const pos = view.posAtCoords({ x, y });
    if (pos === null) return null;
    const line = view.state.doc.lineAt(pos);
    const col = pos - line.from;
    const regex = /\b(\d+)\b/g;
    let match;
    while ((match = regex.exec(line.text)) !== null) {
      if (col >= match.index && col <= match.index + match[0].length) {
        const num = parseInt(match[1], 10);
        if (num > 255) return num;
        break;
      }
    }
    return null;
  }

  function handleContextMenu(e: MouseEvent) {
    e.preventDefault();
    const num = getNumberAtPosition(e.clientX, e.clientY);
    const selection = view?.state.selection.main;
    const hasSelection = selection && selection.from !== selection.to;
    const selectedText = hasSelection ? view?.state.sliceDoc(selection.from, selection.to) : null;

    const items: MenuItem[] = [];

    if (selectedText) {
      items.push({ label: 'Copy', action: () => navigator.clipboard.writeText(selectedText) });
    }

    if (num !== null && onGotoAddress) {
      items.push({ label: `Go to 0x${num.toString(16)}`, action: () => onGotoAddress(num) });
    }

    items.push({ label: 'Copy all', action: () => navigator.clipboard.writeText(content) });

    if (items.length > 0) {
      contextMenu = { x: e.clientX, y: e.clientY, items };
    }
  }

  function handleClick(e: MouseEvent) {
    if (!view || !onGotoAddress || (!e.ctrlKey && !e.metaKey)) return;
    const num = getNumberAtPosition(e.clientX, e.clientY);
    if (num !== null) {
      onGotoAddress(num);
      e.preventDefault();
    }
  }

  $effect(() => {
    if (!container) return;
    if (view) view.destroy();
    view = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions: [
          lineNumbers(),
          oneDark,
          wat,
          syntaxHighlighting(defaultHighlightStyle),
          EditorState.readOnly.of(true),
          search(),
          keymap.of(searchKeymap),
          EditorView.theme({
            '&': { height: '100%', fontSize: '13px' },
            '.cm-scroller': { overflow: 'auto' },
            '.cm-content': { fontFamily: 'ui-monospace, monospace' },
          }),
        ],
      }),
      parent: container,
    });
  });
</script>

<div bind:this={container} class="h-full overflow-hidden" onclick={handleClick} oncontextmenu={handleContextMenu} role="presentation"></div>

{#if contextMenu}
  <ContextMenu items={contextMenu.items} x={contextMenu.x} y={contextMenu.y} onClose={() => contextMenu = null} />
{/if}
