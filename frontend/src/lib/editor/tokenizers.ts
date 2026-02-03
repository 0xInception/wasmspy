import { StateField, RangeSetBuilder, type Extension, type RangeSet } from '@codemirror/state';
import { EditorView, Decoration } from '@codemirror/view';

export type TokenType = 'keyword' | 'type' | 'instruction' | 'function' | 'variable' | 'variableSpecial' | 'string' | 'number' | 'comment' | 'label' | 'paren';

export interface Token {
  from: number;
  to: number;
  type: TokenType;
}

export interface ThemeColors {
  keyword: string;
  type: string;
  function: string;
  variable: string;
  variableSpecial: string;
  string: string;
  number: string;
  comment: string;
  label: string;
  punctuation: string;
}

const watKeywords = new Set(['module', 'func', 'param', 'result', 'local', 'global', 'table', 'memory', 'export', 'import', 'type', 'data', 'elem', 'start', 'offset', 'mut']);
const watTypes = new Set(['i32', 'i64', 'f32', 'f64', 'funcref', 'externref', 'v128']);
const watInstructions = new Set(['unreachable', 'nop', 'block', 'loop', 'if', 'else', 'end', 'br', 'br_if', 'br_table', 'return', 'call', 'call_indirect', 'drop', 'select', 'local.get', 'local.set', 'local.tee', 'global.get', 'global.set', 'i32.load', 'i64.load', 'f32.load', 'f64.load', 'i32.store', 'i64.store', 'f32.store', 'f64.store', 'memory.size', 'memory.grow', 'i32.const', 'i64.const', 'f32.const', 'f64.const', 'i32.add', 'i32.sub', 'i32.mul', 'i32.div_s', 'i32.div_u', 'i32.and', 'i32.or', 'i32.xor', 'i32.shl', 'i32.shr_s', 'i32.shr_u', 'i32.eq', 'i32.ne', 'i32.lt_s', 'i32.lt_u', 'i32.gt_s', 'i32.gt_u', 'i32.le_s', 'i32.le_u', 'i32.ge_s', 'i32.ge_u', 'i32.eqz', 'i64.add', 'i64.sub', 'i64.mul', 'i64.eq', 'i64.ne', 'i64.eqz', 'i32.wrap_i64', 'i64.extend_i32_s', 'i64.extend_i32_u', 'f32.add', 'f32.sub', 'f32.mul', 'f32.div', 'f64.add', 'f64.sub', 'f64.mul', 'f64.div']);
const pseudoKeywords = new Set(['func', 'if', 'else', 'switch', 'case', 'break', 'return', 'loop', 'block', 'br', 'br_if', 'unreachable', 'nop', 'default', 'mem']);
const pseudoTypes = new Set(['i32', 'i64', 'f32', 'f64', 'v128']);

export function tokenizeWat(text: string, lineFrom: number): Token[] {
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

export function tokenizePseudo(text: string, lineFrom: number): Token[] {
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

export function tokenizeDisasm(text: string, lineFrom: number): Token[] {
  const tokens: Token[] = [];
  let i = 0;

  const addrMatch = text.match(/^([0-9a-fA-F]+):/);
  if (addrMatch) {
    tokens.push({ from: lineFrom, to: lineFrom + addrMatch[1].length, type: 'number' });
    i = addrMatch[0].length;
  }

  if (text.trimStart().startsWith(';')) {
    const commentStart = text.indexOf(';');
    tokens.push({ from: lineFrom + commentStart, to: lineFrom + text.length, type: 'comment' });
    return tokens;
  }

  while (i < text.length) {
    if (/\s/.test(text[i])) { i++; continue; }

    if (/[-\d]/.test(text[i]) && (text[i] !== '-' || /\d/.test(text[i + 1] || ''))) {
      const start = i;
      if (text[i] === '-') i++;
      if (text.slice(i, i + 2) === '0x') { i += 2; while (i < text.length && /[0-9a-fA-F]/.test(text[i])) i++; }
      else { while (i < text.length && /[\d.]/.test(text[i])) i++; }
      if (i > start) { tokens.push({ from: lineFrom + start, to: lineFrom + i, type: 'number' }); continue; }
    }

    if (text[i] === '[' || text[i] === ']' || text[i] === ',' ) { i++; continue; }

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

export function createStaticHighlighter(mode: 'disasm' | 'wat' | 'decompile', colors: ThemeColors): Extension {
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
    '.tok-keyword': { color: colors.keyword },
    '.tok-type': { color: colors.type },
    '.tok-instruction': { color: colors.keyword },
    '.tok-function': { color: colors.function },
    '.tok-variable': { color: colors.variable },
    '.tok-variableSpecial': { color: colors.variableSpecial },
    '.tok-string': { color: colors.string },
    '.tok-number': { color: colors.number },
    '.tok-comment': { color: colors.comment, fontStyle: 'italic' },
    '.tok-label': { color: colors.label },
    '.tok-paren': { color: colors.punctuation },
  });

  return [field, tokTheme];
}
