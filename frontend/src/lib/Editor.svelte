<script lang="ts">
  import { onDestroy } from 'svelte';
  import { EditorState, Prec, RangeSetBuilder, StateField, StateEffect, Compartment, type Extension, type RangeSet } from '@codemirror/state';
  import { EditorView, lineNumbers, keymap, drawSelection, Decoration } from '@codemirror/view';
  import { defaultKeymap, selectAll } from '@codemirror/commands';
  import { search, searchKeymap } from '@codemirror/search';
  import { wat } from '../lang-wat';
  import { pseudo } from '../lang-pseudo';
  import { getTheme, getEditorExtensions } from './themes';
  import ContextMenu, { type MenuItem } from './ContextMenu.svelte';

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

  let { content, mode = 'disasm', onGotoAddress, onGotoFunction, onShowXRefs, onRenameFunction, onAddComment, functions, functionsByName, nicknames, lineMappings, onLineClick, onSelectionChange, highlightLines, onShowDecompile, onShowDisassembly, virtualizationThreshold = DEFAULT_VIRTUALIZATION_THRESHOLD, fontSize = DEFAULT_FONT_SIZE }: {
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

  const watKeywords = new Set(['module', 'func', 'param', 'result', 'local', 'global', 'table', 'memory', 'export', 'import', 'type', 'data', 'elem', 'start', 'offset', 'mut']);
  const watTypes = new Set(['i32', 'i64', 'f32', 'f64', 'funcref', 'externref', 'v128']);
  const watInstructions = new Set(['unreachable', 'nop', 'block', 'loop', 'if', 'else', 'end', 'br', 'br_if', 'br_table', 'return', 'call', 'call_indirect', 'drop', 'select', 'local.get', 'local.set', 'local.tee', 'global.get', 'global.set', 'i32.load', 'i64.load', 'f32.load', 'f64.load', 'i32.store', 'i64.store', 'f32.store', 'f64.store', 'memory.size', 'memory.grow', 'i32.const', 'i64.const', 'f32.const', 'f64.const', 'i32.add', 'i32.sub', 'i32.mul', 'i32.div_s', 'i32.div_u', 'i32.and', 'i32.or', 'i32.xor', 'i32.shl', 'i32.shr_s', 'i32.shr_u', 'i32.eq', 'i32.ne', 'i32.lt_s', 'i32.lt_u', 'i32.gt_s', 'i32.gt_u', 'i32.le_s', 'i32.le_u', 'i32.ge_s', 'i32.ge_u', 'i32.eqz', 'i64.add', 'i64.sub', 'i64.mul', 'i64.eq', 'i64.ne', 'i64.eqz', 'i32.wrap_i64', 'i64.extend_i32_s', 'i64.extend_i32_u', 'f32.add', 'f32.sub', 'f32.mul', 'f32.div', 'f64.add', 'f64.sub', 'f64.mul', 'f64.div']);
  const pseudoKeywords = new Set(['func', 'if', 'else', 'switch', 'case', 'break', 'return', 'loop', 'block', 'br', 'br_if', 'unreachable', 'nop', 'default', 'mem']);
  const pseudoTypes = new Set(['i32', 'i64', 'f32', 'f64', 'v128']);

  type TokenType = 'keyword' | 'type' | 'instruction' | 'function' | 'variable' | 'variableSpecial' | 'string' | 'number' | 'comment' | 'label' | 'paren';

  interface Token { from: number; to: number; type: TokenType; }

  function tokenizeWat(text: string, lineFrom: number): Token[] {
    const tokens: Token[] = [];
    let i = 0;
    while (i < text.length) {
      if (/\s/.test(text[i])) { i++; continue; }
      if (text.slice(i, i + 2) === ';;') { tokens.push({ from: lineFrom + i, to: lineFrom + text.length, type: 'comment' }); break; }
      if (text[i] === '"') {
        const start = i++;
        while (i < text.length && text[i] !== '"') { if (text[i] === '\\') i++; i++; }
        tokens.push({ from: lineFrom + start, to: lineFrom + ++i, type: 'string' });
        continue;
      }
      if (text[i] === '$') {
        const start = i++;
        while (i < text.length && /[\w!#$%&'*+\-./:;<=>?@\\^_`|~]/.test(text[i])) i++;
        tokens.push({ from: lineFrom + start, to: lineFrom + i, type: 'variable' });
        continue;
      }
      if (/[-\d]/.test(text[i]) && (text[i] !== '-' || /\d/.test(text[i + 1] || ''))) {
        const start = i;
        if (text[i] === '-') i++;
        if (text.slice(i, i + 2) === '0x') { i += 2; while (i < text.length && /[0-9a-fA-F_]/.test(text[i])) i++; }
        else { while (i < text.length && /[\d_eE.+-]/.test(text[i])) i++; }
        if (i > start) { tokens.push({ from: lineFrom + start, to: lineFrom + i, type: 'number' }); continue; }
      }
      if (text[i] === '(' || text[i] === ')') { tokens.push({ from: lineFrom + i, to: lineFrom + i + 1, type: 'paren' }); i++; continue; }
      if (/[a-zA-Z_]/.test(text[i])) {
        const start = i;
        while (i < text.length && /[\w.]/.test(text[i])) i++;
        const word = text.slice(start, i);
        let type: TokenType = 'variable';
        if (watKeywords.has(word)) type = 'keyword';
        else if (watTypes.has(word)) type = 'type';
        else if (watInstructions.has(word)) type = 'instruction';
        tokens.push({ from: lineFrom + start, to: lineFrom + i, type });
        continue;
      }
      i++;
    }
    return tokens;
  }

  function tokenizePseudo(text: string, lineFrom: number): Token[] {
    const tokens: Token[] = [];
    let i = 0;
    while (i < text.length) {
      if (/\s/.test(text[i])) { i++; continue; }
      if (text.slice(i, i + 2) === '//' || text[i] === ';') { tokens.push({ from: lineFrom + i, to: lineFrom + text.length, type: 'comment' }); break; }
      if (text[i] === '"') {
        const start = i++;
        while (i < text.length && text[i] !== '"') { if (text[i] === '\\') i++; i++; }
        tokens.push({ from: lineFrom + start, to: lineFrom + ++i, type: 'string' });
        continue;
      }
      if (/\bL\d/.test(text.slice(i, i + 10))) {
        const start = i; i++;
        while (i < text.length && /\d/.test(text[i])) i++;
        tokens.push({ from: lineFrom + start, to: lineFrom + i, type: 'label' });
        continue;
      }
      if (text.slice(i, i + 2) === '0x') {
        const start = i; i += 2;
        while (i < text.length && /[0-9a-fA-F]/.test(text[i])) i++;
        tokens.push({ from: lineFrom + start, to: lineFrom + i, type: 'number' });
        continue;
      }
      if (/[-\d]/.test(text[i])) {
        const start = i;
        if (text[i] === '-') i++;
        while (i < text.length && /[\d.]/.test(text[i])) i++;
        if (i > start + (text[start] === '-' ? 1 : 0)) { tokens.push({ from: lineFrom + start, to: lineFrom + i, type: 'number' }); continue; }
        i = start;
      }
      if (/[a-zA-Z_]/.test(text[i])) {
        const start = i;
        while (i < text.length && /[\w.]/.test(text[i])) i++;
        const word = text.slice(start, i);
        let type: TokenType = 'variable';
        if (pseudoKeywords.has(word)) type = 'keyword';
        else if (pseudoTypes.has(word)) type = 'type';
        else if (/^[vp]\d+$/.test(word)) type = 'variable';
        else if (/^global\d+$/.test(word)) type = 'variableSpecial';
        else if (i < text.length && /\s*\(/.test(text.slice(i, i + 5))) type = 'function';
        tokens.push({ from: lineFrom + start, to: lineFrom + i, type });
        continue;
      }
      i++;
    }
    return tokens;
  }

  function tokenizeDisasm(text: string, lineFrom: number): Token[] {
    const tokens: Token[] = [];
    let i = 0;

    // Match leading hex address like "00000000:"
    const addrMatch = text.match(/^([0-9a-fA-F]+):/);
    if (addrMatch) {
      tokens.push({ from: lineFrom, to: lineFrom + addrMatch[1].length, type: 'number' });
      i = addrMatch[0].length;
    }

    // Check for comment line
    if (text.trimStart().startsWith(';')) {
      const commentStart = text.indexOf(';');
      tokens.push({ from: lineFrom + commentStart, to: lineFrom + text.length, type: 'comment' });
      return tokens;
    }

    while (i < text.length) {
      if (/\s/.test(text[i])) { i++; continue; }

      // Numbers (including hex)
      if (/[-\d]/.test(text[i]) && (text[i] !== '-' || /\d/.test(text[i + 1] || ''))) {
        const start = i;
        if (text[i] === '-') i++;
        if (text.slice(i, i + 2) === '0x') { i += 2; while (i < text.length && /[0-9a-fA-F]/.test(text[i])) i++; }
        else { while (i < text.length && /[\d.]/.test(text[i])) i++; }
        if (i > start) { tokens.push({ from: lineFrom + start, to: lineFrom + i, type: 'number' }); continue; }
      }

      // Brackets
      if (text[i] === '[' || text[i] === ']' || text[i] === ',' ) { i++; continue; }

      // Instructions and keywords
      if (/[a-zA-Z_]/.test(text[i])) {
        const start = i;
        while (i < text.length && /[\w.]/.test(text[i])) i++;
        const word = text.slice(start, i);
        let type: TokenType = 'variable';
        if (watInstructions.has(word)) type = 'instruction';
        else if (watTypes.has(word)) type = 'type';
        else if (watKeywords.has(word)) type = 'keyword';
        tokens.push({ from: lineFrom + start, to: lineFrom + i, type });
        continue;
      }
      i++;
    }
    return tokens;
  }

  function createStaticHighlighter(mode: 'disasm' | 'wat' | 'decompile', theme: ReturnType<typeof getTheme>): Extension {
    const c = theme.colors;
    const styles: Record<TokenType, Decoration> = {
      keyword: Decoration.mark({ class: 'tok-keyword' }),
      type: Decoration.mark({ class: 'tok-type' }),
      instruction: Decoration.mark({ class: 'tok-instruction' }),
      function: Decoration.mark({ class: 'tok-function' }),
      variable: Decoration.mark({ class: 'tok-variable' }),
      variableSpecial: Decoration.mark({ class: 'tok-variableSpecial' }),
      string: Decoration.mark({ class: 'tok-string' }),
      number: Decoration.mark({ class: 'tok-number' }),
      comment: Decoration.mark({ class: 'tok-comment' }),
      label: Decoration.mark({ class: 'tok-label' }),
      paren: Decoration.mark({ class: 'tok-paren' }),
    };
    const tokenize = mode === 'decompile' ? tokenizePseudo : (mode === 'wat' ? tokenizeWat : tokenizeDisasm);

    const field = StateField.define<RangeSet<Decoration>>({
      create(state) {
        const allTokens: Token[] = [];
        for (let i = 1; i <= state.doc.lines; i++) {
          const line = state.doc.line(i);
          allTokens.push(...tokenize(line.text, line.from));
        }
        allTokens.sort((a, b) => a.from - b.from);
        const builder = new RangeSetBuilder<Decoration>();
        for (const tok of allTokens) builder.add(tok.from, tok.to, styles[tok.type]);
        return builder.finish();
      },
      update(decos, tr) {
        if (!tr.docChanged) return decos;
        const allTokens: Token[] = [];
        for (let i = 1; i <= tr.state.doc.lines; i++) {
          const line = tr.state.doc.line(i);
          allTokens.push(...tokenize(line.text, line.from));
        }
        allTokens.sort((a, b) => a.from - b.from);
        const builder = new RangeSetBuilder<Decoration>();
        for (const tok of allTokens) builder.add(tok.from, tok.to, styles[tok.type]);
        return builder.finish();
      },
      provide: f => EditorView.decorations.from(f),
    });

    const tokTheme = EditorView.theme({
      '.tok-keyword': { color: c.keyword },
      '.tok-type': { color: c.type },
      '.tok-instruction': { color: c.keyword },
      '.tok-function': { color: c.function },
      '.tok-variable': { color: c.variable },
      '.tok-variableSpecial': { color: c.variableSpecial },
      '.tok-string': { color: c.string },
      '.tok-number': { color: c.number },
      '.tok-comment': { color: c.comment, fontStyle: 'italic' },
      '.tok-label': { color: c.label },
      '.tok-paren': { color: c.punctuation },
    });

    return [field, tokTheme];
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
      extensions.push(createStaticHighlighter(mode, theme));
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
        if (firstLineFrom !== null) {
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

<div bind:this={container} class="h-full overflow-hidden" onclick={handleClick} oncontextmenu={handleContextMenu} onkeydown={handleKeydown} onmousemove={handleMouseMove} onmouseleave={handleMouseLeave} role="presentation" tabindex="-1"></div>

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
