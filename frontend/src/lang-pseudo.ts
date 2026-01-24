import { StreamLanguage } from '@codemirror/language';

const keywords = new Set([
  'func', 'if', 'else', 'switch', 'case', 'break', 'return', 'loop', 'block',
  'br', 'br_if', 'unreachable', 'nop', 'default', 'mem',
]);

const types = new Set(['i32', 'i64', 'f32', 'f64', 'v128']);

export const pseudo = StreamLanguage.define({
  token(stream) {
    if (stream.eatSpace()) return null;

    // Comments
    if (stream.match(/\/\/.*/) || stream.match(/;.*/)) return 'comment';

    // Strings
    if (stream.match(/"(?:[^"\\]|\\.)*"/)) return 'string';

    // Labels L0, L1, etc
    if (stream.match(/\bL\d+\b/)) return 'labelName';

    // Hex numbers
    if (stream.match(/\b0x[0-9a-fA-F]+\b/)) return 'number';

    // Decimal numbers
    if (stream.match(/-?\d+(\.\d+)?/)) return 'number';

    // Variables v0, v1, p0, p1
    if (stream.match(/\b[vp]\d+\b/)) return 'variableName';

    // Global variables
    if (stream.match(/\bglobal\d+\b/)) return 'variableName.special';

    // Function calls: word followed by (
    if (stream.match(/[a-zA-Z_][\w.]*(?=\s*\()/)) {
      const matched = stream.current();
      if (keywords.has(matched)) return 'keyword';
      if (types.has(matched)) return 'typeName';
      return 'function';
    }

    // Keywords, types, and other words
    const word = stream.match(/[a-zA-Z_][\w.]*/);
    if (word && typeof word !== 'boolean') {
      const w = word[0];
      if (keywords.has(w)) return 'keyword';
      if (types.has(w)) return 'typeName';
      return null;
    }

    // Operators
    if (stream.match(/->|[+\-*/%&|^~<>=!]+/)) return 'operator';

    // Brackets and punctuation
    if (stream.match(/[(){}\[\]:,;]/)) return 'punctuation';

    stream.next();
    return null;
  },
});
