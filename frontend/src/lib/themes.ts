import { EditorView } from '@codemirror/view';
import { HighlightStyle, syntaxHighlighting } from '@codemirror/language';
import { tags } from '@lezer/highlight';

export interface Theme {
  name: string;
  colors: {
    // Editor
    editorBg: string;
    editorFg: string;
    editorSelection: string;
    editorLineHighlight: string;
    editorGutter: string;
    editorGutterFg: string;
    editorCursor: string;

    // Sidebar
    sidebarBg: string;
    sidebarFg: string;
    sidebarHover: string;
    sidebarActive: string;
    sidebarBorder: string;

    // Tabs
    tabBg: string;
    tabActiveBg: string;
    tabFg: string;
    tabActiveFg: string;
    tabBorder: string;

    // UI
    panelBg: string;
    panelBorder: string;
    buttonBg: string;
    buttonHover: string;
    buttonActive: string;
    inputBg: string;
    inputBorder: string;
    scrollbarThumb: string;
    scrollbarTrack: string;

    // Syntax
    keyword: string;
    function: string;
    variable: string;
    variableSpecial: string;
    string: string;
    number: string;
    comment: string;
    type: string;
    operator: string;
    punctuation: string;
    label: string;

    // Semantic
    success: string;
    warning: string;
    error: string;
    info: string;

    // Icons
    iconFunction: string;
    iconImport: string;
    iconMemory: string;
    iconTable: string;
    iconGlobal: string;
    iconExport: string;
    iconGroup: string;
  };
}

export const darcula: Theme = {
  name: 'Darcula',
  colors: {
    // Editor - JetBrains Darcula
    editorBg: '#2b2b2b',
    editorFg: '#a9b7c6',
    editorSelection: '#214283',
    editorLineHighlight: '#323232',
    editorGutter: '#313335',
    editorGutterFg: '#606366',
    editorCursor: '#bbbbbb',

    // Sidebar
    sidebarBg: '#1e1f22',
    sidebarFg: '#bbbbbb',
    sidebarHover: '#2d2f33',
    sidebarActive: '#4b6eaf',
    sidebarBorder: '#1a1a1a',

    // Tabs
    tabBg: '#1e1f22',
    tabActiveBg: '#2b2b2b',
    tabFg: '#888888',
    tabActiveFg: '#bbbbbb',
    tabBorder: '#1a1a1a',

    // UI
    panelBg: '#1e1f22',
    panelBorder: '#1a1a1a',
    buttonBg: '#4c5052',
    buttonHover: '#5c6164',
    buttonActive: '#4b6eaf',
    inputBg: '#45494a',
    inputBorder: '#646464',
    scrollbarThumb: '#424242',
    scrollbarTrack: '#1e1f22',

    // Syntax - JetBrains colors
    keyword: '#cc7832',      // Orange
    function: '#ffc66d',     // Yellow
    variable: '#a9b7c6',     // Light gray
    variableSpecial: '#9876aa', // Purple for globals
    string: '#6a8759',       // Green
    number: '#6897bb',       // Blue
    comment: '#808080',      // Gray
    type: '#cc7832',         // Orange (same as keyword)
    operator: '#a9b7c6',     // Light gray
    punctuation: '#a9b7c6',  // Light gray
    label: '#bbb529',        // Yellow for labels

    // Semantic
    success: '#6a8759',
    warning: '#bbb529',
    error: '#ff6b68',
    info: '#6897bb',

    // Icons
    iconFunction: '#ffc66d',
    iconImport: '#808080',
    iconMemory: '#6897bb',
    iconTable: '#cc7832',
    iconGlobal: '#9876aa',
    iconExport: '#6897bb',
    iconGroup: '#bbb529',
  },
};

export const oneDarkPro: Theme = {
  name: 'One Dark Pro',
  colors: {
    // Editor
    editorBg: '#282c34',
    editorFg: '#abb2bf',
    editorSelection: '#3e4451',
    editorLineHighlight: '#2c313c',
    editorGutter: '#282c34',
    editorGutterFg: '#4b5263',
    editorCursor: '#528bff',

    // Sidebar
    sidebarBg: '#21252b',
    sidebarFg: '#9da5b4',
    sidebarHover: '#2c313a',
    sidebarActive: '#2c313a',
    sidebarBorder: '#181a1f',

    // Tabs
    tabBg: '#21252b',
    tabActiveBg: '#282c34',
    tabFg: '#6b717d',
    tabActiveFg: '#abb2bf',
    tabBorder: '#181a1f',

    // UI
    panelBg: '#21252b',
    panelBorder: '#181a1f',
    buttonBg: '#3a3f4b',
    buttonHover: '#4d78cc',
    buttonActive: '#528bff',
    inputBg: '#1b1d23',
    inputBorder: '#181a1f',
    scrollbarThumb: '#4e5666',
    scrollbarTrack: '#21252b',

    // Syntax - One Dark colors
    keyword: '#c678dd',      // Purple
    function: '#61afef',     // Blue
    variable: '#e06c75',     // Red
    variableSpecial: '#e5c07b', // Yellow for globals
    string: '#98c379',       // Green
    number: '#d19a66',       // Orange
    comment: '#5c6370',      // Gray
    type: '#e5c07b',         // Yellow
    operator: '#56b6c2',     // Cyan
    punctuation: '#abb2bf',  // Light gray
    label: '#e5c07b',        // Yellow for labels

    // Semantic
    success: '#98c379',
    warning: '#e5c07b',
    error: '#e06c75',
    info: '#61afef',

    // Icons
    iconFunction: '#61afef',
    iconImport: '#5c6370',
    iconMemory: '#56b6c2',
    iconTable: '#d19a66',
    iconGlobal: '#c678dd',
    iconExport: '#56b6c2',
    iconGroup: '#e5c07b',
  },
};

export const github: Theme = {
  name: 'GitHub Dark',
  colors: {
    // Editor
    editorBg: '#0d1117',
    editorFg: '#e6edf3',
    editorSelection: '#264f78',
    editorLineHighlight: '#161b22',
    editorGutter: '#0d1117',
    editorGutterFg: '#484f58',
    editorCursor: '#58a6ff',

    // Sidebar
    sidebarBg: '#010409',
    sidebarFg: '#c9d1d9',
    sidebarHover: '#161b22',
    sidebarActive: '#1f6feb',
    sidebarBorder: '#21262d',

    // Tabs
    tabBg: '#010409',
    tabActiveBg: '#0d1117',
    tabFg: '#8b949e',
    tabActiveFg: '#e6edf3',
    tabBorder: '#21262d',

    // UI
    panelBg: '#010409',
    panelBorder: '#21262d',
    buttonBg: '#21262d',
    buttonHover: '#30363d',
    buttonActive: '#1f6feb',
    inputBg: '#0d1117',
    inputBorder: '#30363d',
    scrollbarThumb: '#484f58',
    scrollbarTrack: '#010409',

    // Syntax
    keyword: '#ff7b72',
    function: '#d2a8ff',
    variable: '#ffa657',
    variableSpecial: '#ff7b72',
    string: '#a5d6ff',
    number: '#79c0ff',
    comment: '#8b949e',
    type: '#ff7b72',
    operator: '#79c0ff',
    punctuation: '#c9d1d9',
    label: '#d29922',

    // Semantic
    success: '#3fb950',
    warning: '#d29922',
    error: '#f85149',
    info: '#58a6ff',

    // Icons
    iconFunction: '#d2a8ff',
    iconImport: '#8b949e',
    iconMemory: '#79c0ff',
    iconTable: '#ffa657',
    iconGlobal: '#ff7b72',
    iconExport: '#79c0ff',
    iconGroup: '#d29922',
  },
};

export const themes: Theme[] = [darcula, oneDarkPro, github];
export const defaultTheme = darcula;

let currentTheme: Theme = defaultTheme;

export function getTheme(): Theme {
  return currentTheme;
}

export function setTheme(theme: Theme) {
  currentTheme = theme;
  applyTheme(theme);
}

export function applyTheme(theme: Theme) {
  const root = document.documentElement;
  const c = theme.colors;

  // Editor
  root.style.setProperty('--editor-bg', c.editorBg);
  root.style.setProperty('--editor-fg', c.editorFg);
  root.style.setProperty('--editor-selection', c.editorSelection);
  root.style.setProperty('--editor-line-highlight', c.editorLineHighlight);
  root.style.setProperty('--editor-gutter', c.editorGutter);
  root.style.setProperty('--editor-gutter-fg', c.editorGutterFg);
  root.style.setProperty('--editor-cursor', c.editorCursor);

  // Sidebar
  root.style.setProperty('--sidebar-bg', c.sidebarBg);
  root.style.setProperty('--sidebar-fg', c.sidebarFg);
  root.style.setProperty('--sidebar-hover', c.sidebarHover);
  root.style.setProperty('--sidebar-active', c.sidebarActive);
  root.style.setProperty('--sidebar-border', c.sidebarBorder);

  // Tabs
  root.style.setProperty('--tab-bg', c.tabBg);
  root.style.setProperty('--tab-active-bg', c.tabActiveBg);
  root.style.setProperty('--tab-fg', c.tabFg);
  root.style.setProperty('--tab-active-fg', c.tabActiveFg);
  root.style.setProperty('--tab-border', c.tabBorder);

  // UI
  root.style.setProperty('--panel-bg', c.panelBg);
  root.style.setProperty('--panel-border', c.panelBorder);
  root.style.setProperty('--button-bg', c.buttonBg);
  root.style.setProperty('--button-hover', c.buttonHover);
  root.style.setProperty('--button-active', c.buttonActive);
  root.style.setProperty('--input-bg', c.inputBg);
  root.style.setProperty('--input-border', c.inputBorder);
  root.style.setProperty('--scrollbar-thumb', c.scrollbarThumb);
  root.style.setProperty('--scrollbar-track', c.scrollbarTrack);

  // Syntax
  root.style.setProperty('--syntax-keyword', c.keyword);
  root.style.setProperty('--syntax-function', c.function);
  root.style.setProperty('--syntax-variable', c.variable);
  root.style.setProperty('--syntax-string', c.string);
  root.style.setProperty('--syntax-number', c.number);
  root.style.setProperty('--syntax-comment', c.comment);
  root.style.setProperty('--syntax-type', c.type);
  root.style.setProperty('--syntax-operator', c.operator);
  root.style.setProperty('--syntax-punctuation', c.punctuation);

  // Semantic
  root.style.setProperty('--color-success', c.success);
  root.style.setProperty('--color-warning', c.warning);
  root.style.setProperty('--color-error', c.error);
  root.style.setProperty('--color-info', c.info);

  // Icons
  root.style.setProperty('--icon-function', c.iconFunction);
  root.style.setProperty('--icon-import', c.iconImport);
  root.style.setProperty('--icon-memory', c.iconMemory);
  root.style.setProperty('--icon-table', c.iconTable);
  root.style.setProperty('--icon-global', c.iconGlobal);
  root.style.setProperty('--icon-export', c.iconExport);
  root.style.setProperty('--icon-group', c.iconGroup);
}

export function createEditorTheme(theme: Theme, fontSize: number = 13) {
  const c = theme.colors;
  return EditorView.theme({
    '&': {
      height: '100%',
      fontSize: `${fontSize}px`,
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
      caretColor: c.editorCursor,
      lineHeight: '1.4',
      contain: 'layout style',
    },
    '.cm-line': {
      contain: 'layout style',
      lineHeight: '1.4',
    },
    '.cm-cursor, .cm-dropCursor': {
      borderLeftColor: c.editorCursor,
    },
    '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection': {
      backgroundColor: c.editorSelection,
    },
    '.cm-activeLine': {
      backgroundColor: c.editorLineHighlight,
    },
    '.cm-gutters': {
      backgroundColor: c.editorGutter,
      color: c.editorGutterFg,
      border: 'none',
    },
    '.cm-activeLineGutter': {
      backgroundColor: c.editorLineHighlight,
    },
    '.cm-lineNumbers .cm-gutterElement': {
      padding: '0 8px 0 16px',
    },
    '.cm-searchMatch': {
      backgroundColor: '#515c6a',
      outline: '1px solid #515c6a',
    },
    '.cm-searchMatch.cm-searchMatch-selected': {
      backgroundColor: '#2e4134',
    },
  }, { dark: true });
}

export function createHighlightStyle(theme: Theme) {
  const c = theme.colors;
  return HighlightStyle.define([
    { tag: tags.keyword, color: c.keyword },
    { tag: tags.operatorKeyword, color: c.keyword },
    { tag: tags.controlKeyword, color: c.keyword },
    { tag: tags.definitionKeyword, color: c.keyword },
    { tag: tags.function(tags.variableName), color: c.function },
    { tag: tags.function(tags.propertyName), color: c.function },
    { tag: tags.variableName, color: c.variable },
    { tag: tags.special(tags.variableName), color: c.variableSpecial },
    { tag: tags.propertyName, color: c.variable },
    { tag: tags.string, color: c.string },
    { tag: tags.number, color: c.number },
    { tag: tags.comment, color: c.comment, fontStyle: 'italic' },
    { tag: tags.typeName, color: c.type },
    { tag: tags.className, color: c.type },
    { tag: tags.labelName, color: c.label },
    { tag: tags.operator, color: c.operator },
    { tag: tags.paren, color: c.punctuation },
    { tag: tags.bracket, color: c.punctuation },
    { tag: tags.brace, color: c.punctuation },
    { tag: tags.punctuation, color: c.punctuation },
  ]);
}

export function getEditorExtensions(theme: Theme, fontSize: number = 13) {
  return [
    createEditorTheme(theme, fontSize),
    syntaxHighlighting(createHighlightStyle(theme)),
  ];
}
