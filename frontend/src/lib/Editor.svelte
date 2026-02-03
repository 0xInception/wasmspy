<script lang="ts">
  import { onDestroy } from 'svelte';
  import { EditorState, Prec, Compartment, type Extension } from '@codemirror/state';
  import { EditorView, lineNumbers, keymap, drawSelection, Decoration } from '@codemirror/view';
  import { defaultKeymap, selectAll } from '@codemirror/commands';
  import { search, searchKeymap } from '@codemirror/search';
  import { wat } from '../lang-wat';
  import { pseudo } from '../lang-pseudo';
  import { getTheme, getEditorExtensions } from './themes';
  import ContextMenu, { type MenuItem } from './ContextMenu.svelte';
  import { createStaticHighlighter } from './editor/tokenizers';

  const DEFAULT_VIRTUALIZATION_THRESHOLD = 3000;
  const DEFAULT_FONT_SIZE = 13;

  function getCurrentFuncIndex(): number | null {
    // Parse function index from first line: ";; Function 123: name" or "func name(...)"
    const newlineIdx = content.indexOf('\n');
    const firstLine = newlineIdx >= 0 ? content.slice(0, newlineIdx) : content;
    const match = firstLine.match(/;;\s*Function\s+(\d+):/);
    if (match) return parseInt(match[1], 10);
    // For decompile mode: "func funcName("
    const decompMatch = firstLine.match(/^func\s+(\S+)\s*\(/);
    if (decompMatch && functions) {
      for (const fn of functions) {
        if (fn.name === decompMatch[1]) return fn.index;
      }
    }
    return null;
  }

  let { content, mode = 'disasm', onGotoAddress, onGotoFunction, onShowXRefs, onRenameFunction, onAddComment, functions, functionsByName, nicknames, lineMappings, onLineClick, onSelectionChange, highlightLines, onShowDecompile, onShowDisassembly, virtualizationThreshold = DEFAULT_VIRTUALIZATION_THRESHOLD, fontSize = DEFAULT_FONT_SIZE, onFocus, syncScroll = true }: {
    content: string;
    mode?: 'disasm' | 'wat' | 'decompile';
    onGotoAddress?: (addr: number) => void;
    onGotoFunction?: (index: number) => void;
    onShowXRefs?: (index: number) => void;
    onRenameFunction?: (index: number) => void;
    onAddComment?: (offset: number, comment: string) => void;
    functions?: { index: number; name: string }[];
    functionsByName?: Map<string, { index: number; name: string }>;
    nicknames?: Map<number, string>;
    lineMappings?: Map<number, number[]>;
    onLineClick?: (lineNumber: number) => void;
    onSelectionChange?: (startLine: number, endLine: number) => void;
    highlightLines?: number[] | null;
    onShowDecompile?: () => void;
    onShowDisassembly?: () => void;
    virtualizationThreshold?: number;
    fontSize?: number;
    onFocus?: () => void;
    syncScroll?: boolean;
  } = $props();

  let container: HTMLDivElement;
  let view: EditorView | null = null;
  let lastContent: string = '';
  let contextMenu: { x: number; y: number; items: MenuItem[] } | null = $state(null);
  let tooltip: { x: number; y: number; text: string } | null = $state(null);
  let hoverTimeout: ReturnType<typeof setTimeout> | null = null;
  let commentInput: { offset: number; x: number; y: number; value: string } | null = $state(null);
  let commentInputEl: HTMLInputElement | null = null;

  function getLineCount(text: string): number {
    let count = 1;
    for (let i = 0; i < text.length; i++) {
      if (text[i] === '\n') count++;
    }
    return count;
  }

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

  function getFunctionAtPosition(x: number, y: number): number | null {
    if (!view) return null;
    const pos = view.posAtCoords({ x, y });
    if (pos === null) return null;
    const line = view.state.doc.lineAt(pos);
    const col = pos - line.from;
    const text = line.text;

    if (mode === 'disasm' || mode === 'wat') {
      // Look for "call N" pattern in disassembly/WAT
      const callRegex = /\bcall\s+(\d+)\b/g;
      let match;
      while ((match = callRegex.exec(text)) !== null) {
        const numStart = match.index + match[0].indexOf(match[1]);
        const numEnd = numStart + match[1].length;
        if (col >= numStart && col <= numEnd) {
          return parseInt(match[1], 10);
        }
      }
    } else if (mode === 'decompile' && functions) {
      // Look for function names followed by ( in decompile mode
      const fnCallRegex = /\b([a-zA-Z_][\w.]*)\s*\(/g;
      let match;
      while ((match = fnCallRegex.exec(text)) !== null) {
        const nameStart = match.index;
        const nameEnd = match.index + match[1].length;
        if (col >= nameStart && col <= nameEnd) {
          const fnName = match[1];
          const exactMatch = functionsByName?.get(fnName);
          if (exactMatch) return exactMatch.index;
          for (const f of functions) {
            if (f.name.endsWith('.' + fnName)) return f.index;
          }
        }
      }
    }
    return null;
  }

  function getFunctionAtCursor(): number | null {
    if (!view) return null;
    const sel = view.state.selection.main;
    const line = view.state.doc.lineAt(sel.head);
    const col = sel.head - line.from;
    const text = line.text;

    if (mode === 'disasm' || mode === 'wat') {
      const callRegex = /\bcall\s+(\d+)\b/g;
      let match;
      while ((match = callRegex.exec(text)) !== null) {
        const numStart = match.index + match[0].indexOf(match[1]);
        const numEnd = numStart + match[1].length;
        if (col >= match.index && col <= numEnd) {
          return parseInt(match[1], 10);
        }
      }
    } else if (mode === 'decompile' && functions) {
      const fnCallRegex = /\b([a-zA-Z_][\w.]*)\s*\(/g;
      let match;
      while ((match = fnCallRegex.exec(text)) !== null) {
        const nameStart = match.index;
        const parenPos = match.index + match[0].length - 1;
        if (col >= nameStart && col <= parenPos) {
          const fnName = match[1];
          const exactMatch = functionsByName?.get(fnName);
          if (exactMatch) return exactMatch.index;
          for (const f of functions) {
            if (f.name.endsWith('.' + fnName)) return f.index;
          }
        }
      }
    }
    return null;
  }

  function handleContextMenu(e: MouseEvent) {
    e.preventDefault();
    const num = getNumberAtPosition(e.clientX, e.clientY);
    const fnIndex = getFunctionAtPosition(e.clientX, e.clientY);
    const selection = view?.state.selection.main;
    const hasSelection = selection && selection.from !== selection.to;
    const selectedText = hasSelection ? view?.state.sliceDoc(selection.from, selection.to) : null;

    const items: MenuItem[] = [];

    if (selectedText) {
      items.push({ label: 'Copy', action: () => navigator.clipboard.writeText(selectedText) });
    }

    if (fnIndex !== null && onGotoFunction) {
      const fnName = functions?.find(f => f.index === fnIndex)?.name;
      items.push({ label: `Go to definition${fnName ? ` (${fnName})` : ''}`, action: () => onGotoFunction(fnIndex) });
    }

    if (num !== null && onGotoAddress) {
      items.push({ label: `Go to 0x${num.toString(16)}`, action: () => onGotoAddress(num) });
    }

    const currentFuncIdx = getCurrentFuncIndex();
    if (currentFuncIdx !== null && onShowXRefs) {
      items.push({ label: 'View references', action: () => onShowXRefs(currentFuncIdx) });
    }

    if (currentFuncIdx !== null && onRenameFunction) {
      items.push({ label: 'Add nickname', action: () => onRenameFunction(currentFuncIdx) });
    }

    if (onAddComment) {
      const offset = getOffsetAtCursor();
      if (offset !== null) {
        items.push({ label: 'Add comment', action: () => showCommentInput(offset) });
      }
    }

    if (onShowDecompile) {
      items.push({ label: 'Show Decompile', action: onShowDecompile });
    }
    if (onShowDisassembly) {
      items.push({ label: 'Show Disassembly', action: onShowDisassembly });
    }

    items.push({ label: 'Copy all', action: () => navigator.clipboard.writeText(content) });

    if (items.length > 0) {
      contextMenu = { x: e.clientX, y: e.clientY, items };
    }
  }

  function handleClick(e: MouseEvent) {
    if (!view) return;

    if (e.ctrlKey || e.metaKey) {
      if (onGotoFunction) {
        const fnIndex = getFunctionAtPosition(e.clientX, e.clientY);
        if (fnIndex !== null) {
          onGotoFunction(fnIndex);
          e.preventDefault();
          return;
        }
      }
      if (onGotoAddress) {
        const num = getNumberAtPosition(e.clientX, e.clientY);
        if (num !== null) {
          onGotoAddress(num);
          e.preventDefault();
          return;
        }
      }
    }

    setTimeout(() => notifySelectionChange(), 0);
  }

  function notifySelectionChange() {
    if (!view) return;
    const sel = view.state.selection.main;
    const startLine = view.state.doc.lineAt(sel.from).number;
    const endLine = view.state.doc.lineAt(sel.to).number;
    if (onSelectionChange) {
      onSelectionChange(startLine, endLine);
    } else if (onLineClick && startLine === endLine) {
      onLineClick(startLine);
    }
  }

  function getOffsetAtCursor(): number | null {
    if (!view) return null;
    const sel = view.state.selection.main;
    const line = view.state.doc.lineAt(sel.head);
    const lineNum = line.number;
    const text = line.text;
    if (mode === 'disasm') {
      const match = text.match(/^([0-9a-fA-F]+):/);
      if (match) return parseInt(match[1], 16);
    } else if (mode === 'decompile' && lineMappings) {
      const offsets = lineMappings.get(lineNum);
      if (offsets && offsets.length > 0) return offsets[0];
    }
    return null;
  }

  function showCommentInput(offset: number) {
    if (!view) return;
    const sel = view.state.selection.main;
    const coords = view.coordsAtPos(sel.head);
    if (coords) {
      commentInput = { offset, x: coords.left, y: coords.top, value: '' };
    }
  }

  function submitComment() {
    if (commentInput && commentInput.value.trim() && onAddComment) {
      onAddComment(commentInput.offset, commentInput.value.trim());
    }
    commentInput = null;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'F12' && onGotoFunction) {
      const fnIndex = getFunctionAtCursor();
      if (fnIndex !== null) {
        onGotoFunction(fnIndex);
        e.preventDefault();
      }
    }
    if (e.key === ';' && onAddComment) {
      const offset = getOffsetAtCursor();
      if (offset !== null) {
        e.preventDefault();
        showCommentInput(offset);
      }
    }
    if (e.key === 'ArrowUp' || e.key === 'ArrowDown' || e.key === 'PageUp' || e.key === 'PageDown' || e.key === 'Home' || e.key === 'End') {
      setTimeout(() => notifySelectionChange(), 0);
    }
  }

  function handleMouseMove(e: MouseEvent) {
    if (!nicknames || nicknames.size === 0) return;
    if (hoverTimeout) clearTimeout(hoverTimeout);
    tooltip = null;
    hoverTimeout = setTimeout(() => {
      const fnIndex = getFunctionAtPosition(e.clientX, e.clientY);
      if (fnIndex !== null) {
        const nickname = nicknames.get(fnIndex);
        if (nickname) {
          tooltip = { x: e.clientX, y: e.clientY - 30, text: nickname };
        }
      }
    }, 200);
  }

  function handleMouseLeave() {
    if (hoverTimeout) {
      clearTimeout(hoverTimeout);
      hoverTimeout = null;
    }
    tooltip = null;
  }

  function handleFocusOut(e: FocusEvent) {
    if (!view) return;
    const related = e.relatedTarget as Node | null;
    if (related && container?.contains(related)) return;
    view.dispatch({
      selection: { anchor: view.state.selection.main.head },
    });
  }

  onDestroy(() => {
    if (hoverTimeout) clearTimeout(hoverTimeout);
    if (view) {
      view.destroy();
      view = null;
    }
  });

  const highlightLineDeco = Decoration.line({ class: 'cm-highlight-line' });
  const highlightCompartment = new Compartment();

  const highlightTheme = EditorView.theme({
    '.cm-highlight-line': { backgroundColor: 'rgba(255, 200, 0, 0.15)' },
  });

  $effect(() => {
    if (!container) return;

    if (view && view.dom.parentElement === container) {
      if (lastContent !== content) {
        view.dispatch({
          changes: { from: 0, to: lastContent.length, insert: content }
        });
        lastContent = content;
      }
      return;
    }

    if (view) {
      view.destroy();
      view = null;
    }

    const theme = getTheme();
    const lineCount = getLineCount(content);
    const useStaticHighlighting = lineCount < virtualizationThreshold;

    const extensions = [
      Prec.high(keymap.of([
        { key: 'Mod-a', run: selectAll },
        ...defaultKeymap,
        ...searchKeymap,
      ])),
      lineNumbers(),
      drawSelection(),
      ...getEditorExtensions(theme, fontSize),
      EditorState.readOnly.of(true),
      search(),
      highlightTheme,
      highlightCompartment.of([]),
    ];

    if (useStaticHighlighting) {
      extensions.push(createStaticHighlighter(mode, theme.colors));
    } else {
      extensions.push(mode === 'decompile' ? pseudo : wat);
    }

    view = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions,
      }),
      parent: container,
    });
    lastContent = content;

    return () => {
      if (view) {
        view.destroy();
        view = null;
      }
    };
  });

  let lastHighlightLines: number[] | null = null;

  $effect(() => {
    if (!view) return;

    const same = lastHighlightLines === highlightLines ||
      (lastHighlightLines && highlightLines &&
       lastHighlightLines.length === highlightLines.length &&
       lastHighlightLines.every((v, i) => v === highlightLines![i]));
    if (same) return;
    lastHighlightLines = highlightLines ? [...highlightLines] : null;

    if (highlightLines && highlightLines.length > 0) {
      const decorations: any[] = [];
      let firstLineFrom: number | null = null;
      for (const lineNum of highlightLines) {
        if (lineNum > 0 && lineNum <= view.state.doc.lines) {
          const line = view.state.doc.line(lineNum);
          decorations.push(highlightLineDeco.range(line.from));
          if (firstLineFrom === null) firstLineFrom = line.from;
        }
      }
      if (decorations.length > 0) {
        view.dispatch({
          effects: highlightCompartment.reconfigure(
            EditorView.decorations.of(Decoration.set(decorations.sort((a, b) => a.from - b.from)))
          ),
        });
        if (firstLineFrom !== null && syncScroll) {
          view.dispatch({ effects: EditorView.scrollIntoView(firstLineFrom, { y: 'center' }) });
        }
      } else {
        view.dispatch({ effects: highlightCompartment.reconfigure([]) });
      }
    } else {
      view.dispatch({ effects: highlightCompartment.reconfigure([]) });
    }
  });

  $effect(() => {
    if (commentInput && commentInputEl) {
      commentInputEl.focus();
    }
  });
</script>

<div bind:this={container} class="h-full overflow-hidden" onclick={handleClick} oncontextmenu={handleContextMenu} onkeydown={handleKeydown} onmousemove={handleMouseMove} onmouseleave={handleMouseLeave} onfocusin={() => onFocus?.()} onfocusout={handleFocusOut} role="presentation" tabindex="-1"></div>

{#if contextMenu}
  <ContextMenu items={contextMenu.items} x={contextMenu.x} y={contextMenu.y} onClose={() => contextMenu = null} />
{/if}

{#if tooltip}
  <div class="fixed px-2 py-1 text-xs rounded shadow-lg pointer-events-none z-50" style="left: {tooltip.x}px; top: {tooltip.y}px; background: var(--panel-bg); color: var(--syntax-function); border: 1px solid var(--panel-border);">
    {tooltip.text}
  </div>
{/if}

{#if commentInput}
  <div class="fixed z-50" style="left: {commentInput.x}px; top: {commentInput.y}px;">
    <input
      bind:this={commentInputEl}
      type="text"
      class="px-2 py-1 text-xs rounded outline-none"
      style="background: var(--editor-bg); color: var(--editor-fg); border: 1px solid var(--button-active); min-width: 200px;"
      placeholder="; comment"
      bind:value={commentInput.value}
      onkeydown={(e) => {
        if (e.key === 'Enter') submitComment();
        if (e.key === 'Escape') commentInput = null;
        e.stopPropagation();
      }}
    />
  </div>
{/if}
