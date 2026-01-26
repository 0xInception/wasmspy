<script lang="ts">
  import { onMount } from 'svelte';
  import { LoadModuleFromPath, DisassembleFunction, DecompileFunctionWithMappings, GetFunctionWAT, OpenFileDialog, GetXRefs, GetModuleErrors, GetAnnotations, SetFunctionName, SetOffsetComment, SetBookmarks, SaveAnnotations, ClearAnnotations, ExportAnnotationsToFile, ImportAnnotationsFromFile, SetWindowSize } from '../wailsjs/go/main/App';
  import { Quit } from '../wailsjs/runtime/runtime';
  import { EventsOn } from '../wailsjs/runtime/runtime';
  import type { ModuleInfo, FunctionInfo, MemoryInfo, TableInfo, GlobalInfo, ExportInfo, Bookmark, XRefInfo, ModuleErrorsInfo, Annotations } from './lib/types';
  import Explorer from './lib/Explorer.svelte';
  import EditorSplit from './lib/EditorSplit.svelte';
  import HexView from './lib/HexView.svelte';
  import SplitPane from './lib/SplitPane.svelte';
  import TabBar, { type Tab } from './lib/TabBar.svelte';
  import CommandPalette from './lib/CommandPalette.svelte';
  import RenameModal from './lib/RenameModal.svelte';
  import { applyTheme, defaultTheme } from './lib/themes';
  import SettingsModal from './lib/SettingsModal.svelte';
  import { loadSettings, saveSettings, windowSizePresets, type Settings } from './lib/settings';

  applyTheme(defaultTheme);

  onMount(() => {
    const initialSettings = loadSettings();
    if (initialSettings.windowSize && initialSettings.windowSize !== 'custom') {
      const preset = windowSizePresets[initialSettings.windowSize];
      if (preset) {
        SetWindowSize(preset.width, preset.height);
      }
    }
  });

  type GroupedFunctions = [string, FunctionInfo[]][];

  interface LoadedModule {
    path: string;
    name: string;
    info: ModuleInfo;
    functionsById: Map<number, FunctionInfo>;
    functionsByName: Map<string, FunctionInfo>;
    groupedFunctions: GroupedFunctions;
    groupedImports: GroupedFunctions;
    annotations: Annotations;
  }

  function buildFunctionMaps(functions: FunctionInfo[] | null | undefined): { byId: Map<number, FunctionInfo>; byName: Map<string, FunctionInfo> } {
    const byId = new Map<number, FunctionInfo>();
    const byName = new Map<string, FunctionInfo>();
    if (functions) {
      for (const fn of functions) {
        byId.set(fn.index, fn);
        byName.set(fn.name, fn);
      }
    }
    return { byId, byName };
  }

  function getPrefix(name: string): string {
    const dot = name.indexOf('.');
    return dot > 0 ? name.slice(0, dot) : '_ungrouped';
  }

  function buildGroupedFunctions(functions: FunctionInfo[] | null | undefined, imported: boolean): GroupedFunctions {
    if (!functions) return [];
    const groups = new Map<string, FunctionInfo[]>();
    for (const fn of functions) {
      if (fn.imported !== imported) continue;
      const prefix = getPrefix(fn.name);
      if (!groups.has(prefix)) groups.set(prefix, []);
      groups.get(prefix)!.push(fn);
    }
    for (const fns of groups.values()) {
      fns.sort((a, b) => a.name.localeCompare(b.name));
    }
    return Array.from(groups.entries()).sort((a, b) => a[0].localeCompare(b[0]));
  }

  interface LineMapping {
    line: number;
    offsets: number[];
  }

  interface DecompileMappingsIndexed {
    byLine: Map<number, number[]>;
    byOffset: Map<number, number[]>;
  }

  interface OpenTab extends Tab {
    type: 'function' | 'memory';
    modulePath: string;
    index: number;
    decompileContent: string | null;
    decompileMappings: DecompileMappingsIndexed | null;
    decompileLineCount: number;
    disasmContent: string | null;
    disasmMappings: DisasmMappings | null;
    disasmLineCount: number;
    showLeft: boolean;
    showRight: boolean;
    showOffsets: boolean;
    disasmIndent: boolean;
  }

  let modules: LoadedModule[] = $state([]);
  let activeModuleIndex: number = $state(-1);
  let error = $state('');
  let loading = $state(false);
  let loadingName = $state('');
  let selected = $state('');
  let tabsById: Map<string, OpenTab> = $state(new Map());
  let tabOrder: string[] = $state([]);
  let activeTabId: string | null = $state(null);
  let gotoMemAddress: number | null = $state(null);
  let showCommandPalette = $state(false);
  let xrefs: XRefInfo | null = $state(null);
  let xrefsFuncName: string = $state('');
  let functionErrors: Map<string, Map<number, number>> = $state(new Map());
  let leftHighlightLines: number[] | null = $state(null);
  let rightHighlightLines: number[] | null = $state(null);
  let renameTarget: FunctionInfo | null = $state(null);
  let dirtyModules: Set<string> = $state(new Set());
  let showQuitConfirm = $state(false);
  let showSettings = $state(false);
  let settings: Settings = $state(loadSettings());

  interface CachedFunction {
    decompileCode: string;
    decompileMappings: DecompileMappingsIndexed | null;
    disasmContent: string;
    disasmMappings: DisasmMappings;
  }
  let functionCache: Map<string, CachedFunction> = $state(new Map());

  function getFunctionErrorCount(modulePath: string, funcIndex: number): number {
    return functionErrors.get(modulePath)?.get(funcIndex) ?? 0;
  }

  let bookmarks = $derived.by(() => {
    if (!activeModule?.annotations?.bookmarks) return [];
    return activeModule.annotations.bookmarks.map(idx => {
      const fn = activeModule.functionsById.get(idx);
      return {
        id: `${activeModule.path}:${idx}`,
        modulePath: activeModule.path,
        funcIndex: idx,
        funcName: fn?.name || `func_${idx}`,
      };
    });
  });

  let bookmarkIds = $derived(new Set(bookmarks.map(b => b.id)));

  async function addBookmark(fn: FunctionInfo) {
    if (!activeModule) return;
    const current = activeModule.annotations?.bookmarks || [];
    if (current.includes(fn.index)) return;
    const newBookmarks = [...current, fn.index];
    await SetBookmarks(activeModule.path, newBookmarks);
    const annotations = await GetAnnotations(activeModule.path) as Annotations;
    modules = modules.map((m, i) => i === activeModuleIndex ? { ...m, annotations } : m);
    markDirty(activeModule.path);
  }

  async function removeBookmark(id: string) {
    if (!activeModule) return;
    const parts = id.split(':');
    const idx = parseInt(parts[parts.length - 1]);
    const current = activeModule.annotations?.bookmarks || [];
    const newBookmarks = current.filter(b => b !== idx);
    await SetBookmarks(activeModule.path, newBookmarks);
    const annotations = await GetAnnotations(activeModule.path) as Annotations;
    modules = modules.map((m, i) => i === activeModuleIndex ? { ...m, annotations } : m);
    markDirty(activeModule.path);
  }

  function isBookmarked(modulePath: string, funcIndex: number): boolean {
    return bookmarkIds.has(`${modulePath}:${funcIndex}`);
  }

  async function toggleBookmark(fn: FunctionInfo) {
    if (!activeModule) return;
    if (isBookmarked(activeModule.path, fn.index)) {
      await removeBookmark(`${activeModule.path}:${fn.index}`);
    } else {
      await addBookmark(fn);
    }
  }

  function renameFunction(fn: FunctionInfo) {
    renameTarget = fn;
  }

  async function handleRenameConfirm(newName: string) {
    if (!activeModule || !renameTarget) return;
    try {
      await SetFunctionName(activeModule.path, renameTarget.index, newName);
      const annotations = await GetAnnotations(activeModule.path) as Annotations;
      modules = modules.map((m, i) => i === activeModuleIndex ? { ...m, annotations } : m);
      markDirty(activeModule.path);
    } catch (e: any) {
      error = e.message || String(e);
    }
    renameTarget = null;
  }

  function markDirty(path: string) {
    dirtyModules = new Set(dirtyModules).add(path);
  }

  function markClean(path: string) {
    const newSet = new Set(dirtyModules);
    newSet.delete(path);
    dirtyModules = newSet;
  }

  async function handleSave() {
    if (!activeModule) return;
    try {
      await SaveAnnotations(activeModule.path);
      markClean(activeModule.path);
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  async function handleImport() {
    if (!activeModule) return;
    try {
      const importedPath = await ImportAnnotationsFromFile(activeModule.path);
      if (importedPath) {
        const annotations = await GetAnnotations(activeModule.path) as Annotations;
        modules = modules.map((m, i) => i === activeModuleIndex ? { ...m, annotations } : m);
        markDirty(activeModule.path);
      }
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  async function handleExport() {
    if (!activeModule) return;
    try {
      await ExportAnnotationsToFile(activeModule.path);
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  async function handleClear() {
    if (!activeModule) return;
    try {
      const annotations = await ClearAnnotations(activeModule.path) as Annotations;
      modules = modules.map((m, i) => i === activeModuleIndex ? { ...m, annotations } : m);
      markDirty(activeModule.path);
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  function handleQuitRequest() {
    if (dirtyModules.size > 0) {
      showQuitConfirm = true;
    } else {
      Quit();
    }
  }

  function handleSaveSettings(newSettings: Settings) {
    settings = newSettings;
    saveSettings(newSettings);
    if (newSettings.windowSize && newSettings.windowSize !== 'custom') {
      const preset = windowSizePresets[newSettings.windowSize];
      if (preset) {
        SetWindowSize(preset.width, preset.height);
      }
    }
  }

  async function handleQuitWithSave() {
    for (const path of dirtyModules) {
      try {
        await SaveAnnotations(path);
      } catch (e) {
        console.error('Failed to save', path, e);
      }
    }
    Quit();
  }

  function handleQuitWithoutSave() {
    Quit();
  }

  function handleRenameCancel() {
    renameTarget = null;
  }

  async function handleAddComment(offset: number, comment: string, isDecompile: boolean) {
    if (!activeModule || !activeTab || activeTab.type !== 'function') return;
    try {
      await SetOffsetComment(activeModule.path, offset, comment, isDecompile);
      if (!isDecompile) {
        const disasmContent = await DisassembleFunction(activeModule.path, activeTab.index, activeTab.disasmIndent);
        const disasmMappings = parseDisasmOffsets(disasmContent);
        tabsById = new Map(tabsById).set(activeTab.id, {
          ...activeTab,
          disasmContent,
          disasmMappings,
          disasmLineCount: countLines(disasmContent),
        });
        const cacheKey = `${activeModule.path}:${activeTab.index}`;
        const cached = functionCache.get(cacheKey);
        if (cached) {
          functionCache = new Map(functionCache).set(cacheKey, { ...cached, disasmContent, disasmMappings });
        }
      }
      const annotations = await GetAnnotations(activeModule.path) as Annotations;
      modules = modules.map((m, i) => i === activeModuleIndex ? { ...m, annotations } : m);
      markDirty(activeModule.path);
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  function handleGlobalKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === 'p') {
      e.preventDefault();
      if (activeModule?.info.functions?.length) {
        showCommandPalette = true;
      }
    }
    if ((e.ctrlKey || e.metaKey) && e.key === 'w') {
      e.preventDefault();
      if (activeTabId) closeTab(activeTabId);
    }
    if ((e.ctrlKey || e.metaKey) && e.key === 'Tab') {
      e.preventDefault();
      if (tabOrder.length > 1 && activeTabId) {
        const idx = tabOrder.indexOf(activeTabId);
        const nextIdx = e.shiftKey
          ? (idx - 1 + tabOrder.length) % tabOrder.length
          : (idx + 1) % tabOrder.length;
        selectTab(tabOrder[nextIdx]);
      }
    }
    if ((e.ctrlKey || e.metaKey) && e.key === 'b') {
      e.preventDefault();
      if (activeTab?.type === 'function' && activeModule) {
        const fn = activeModule.functionsById.get(activeTab.index);
        if (fn) toggleBookmark(fn);
      }
    }
  }

  let activeModule = $derived(activeModuleIndex >= 0 ? modules[activeModuleIndex] : null);
  let modulesByPath = $derived(new Map(modules.map(m => [m.path, m])));
  let nicknames = $derived.by(() => {
    const map = new Map<number, string>();
    if (activeModule?.annotations?.functions) {
      for (const [idx, ann] of Object.entries(activeModule.annotations.functions)) {
        if (ann.name) map.set(parseInt(idx), ann.name);
      }
    }
    return map;
  });

  let disasmComments = $derived.by(() => {
    const map = new Map<number, string>();
    if (activeModule?.annotations?.comments) {
      for (const [offset, comment] of Object.entries(activeModule.annotations.comments)) {
        map.set(parseInt(offset), comment);
      }
    }
    return map;
  });

  let decompileComments = $derived.by(() => {
    const map = new Map<number, string>();
    if (activeModule?.annotations?.decompileComments) {
      for (const [offset, comment] of Object.entries(activeModule.annotations.decompileComments)) {
        map.set(parseInt(offset), comment);
      }
    }
    return map;
  });
  let tabs = $derived(tabOrder.map(id => tabsById.get(id)!));
  let activeTab = $derived(activeTabId ? tabsById.get(activeTabId) ?? null : null);

  async function loadFromPath(path: string) {
    const existing = modules.findIndex(m => m.path === path);
    if (existing >= 0) {
      activeModuleIndex = existing;
      return;
    }
    error = '';
    loading = true;
    loadingName = path.split('/').pop() || path;
    try {
      const [info, annotations] = await Promise.all([
        LoadModuleFromPath(path) as Promise<ModuleInfo>,
        GetAnnotations(path) as Promise<Annotations>,
      ]);
      const { byId, byName } = buildFunctionMaps(info.functions);
      const groupedFunctions = buildGroupedFunctions(info.functions, false);
      const groupedImports = buildGroupedFunctions(info.functions, true);
      modules = [...modules, { path, name: loadingName, info, functionsById: byId, functionsByName: byName, groupedFunctions, groupedImports, annotations }];
      activeModuleIndex = modules.length - 1;
      selected = '';

      try {
        const errors = await GetModuleErrors(path) as ModuleErrorsInfo;
        if (errors.functions) {
          const errMap = new Map<number, number>();
          for (const fn of errors.functions) {
            errMap.set(fn.funcIndex, fn.errors.length);
          }
          functionErrors = new Map(functionErrors).set(path, errMap);
        }
        if (errors.totalErrors > 0) {
          console.group(`Decompiler errors for ${loadingName}`);
          console.log(`Total errors: ${errors.totalErrors}`);
          console.log('Unsupported opcodes:', errors.uniqueErrors);
          console.log(`Affected functions: ${errors.functions?.length || 0}`);
          if (errors.functions) {
            for (const fn of errors.functions.slice(0, 10)) {
              console.log(`  ${fn.funcName}: ${fn.errors.length} errors`);
            }
            if (errors.functions.length > 10) {
              console.log(`  ... and ${errors.functions.length - 10} more functions`);
            }
          }
          console.groupEnd();
        }
      } catch (e) {
        console.warn('Failed to get module errors:', e);
      }
    } catch (e: any) {
      error = e.message || String(e);
    } finally {
      loading = false;
      loadingName = '';
    }
  }

  EventsOn('filedrop', (path: string) => loadFromPath(path));
  EventsOn('menu:open', () => openFile());
  EventsOn('menu:copy', () => document.execCommand('copy'));
  EventsOn('menu:selectall', () => document.execCommand('selectAll'));
  EventsOn('menu:save', () => handleSave());
  EventsOn('menu:import', () => handleImport());
  EventsOn('menu:export', () => handleExport());
  EventsOn('menu:clear', () => handleClear());
  EventsOn('menu:quit', () => handleQuitRequest());
  EventsOn('menu:preferences', () => showSettings = true);

  async function openFile() {
    try {
      const path = await OpenFileDialog();
      if (path) loadFromPath(path);
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  function selectModule(index: number) {
    activeModuleIndex = index;
    selected = '';
  }

  function closeModule(index: number) {
    modules = modules.filter((_, i) => i !== index);
    if (activeModuleIndex >= modules.length) {
      activeModuleIndex = modules.length - 1;
    }
    if (activeModuleIndex < 0) {
      selected = '';
    }
  }

  interface DisasmMappings {
    offsetToLine: Map<number, number>;
    lineToOffset: Map<number, number>;
  }

  function countLines(text: string): number {
    let count = 1;
    for (let i = 0; i < text.length; i++) {
      if (text[i] === '\n') count++;
    }
    return count;
  }

  function buildDecompileMappings(mappings: LineMapping[]): DecompileMappingsIndexed {
    const byLine = new Map<number, number[]>();
    const byOffset = new Map<number, number[]>();
    for (const m of mappings) {
      const offsets = m.offsets.map(o => Number(o));
      byLine.set(m.line, offsets);
      for (const off of offsets) {
        const existing = byOffset.get(off);
        if (existing) existing.push(m.line);
        else byOffset.set(off, [m.line]);
      }
    }
    return { byLine, byOffset };
  }

  function parseDisasmOffsets(disasmContent: string): DisasmMappings {
    const offsetToLine = new Map<number, number>();
    const lineToOffset = new Map<number, number>();
    let lineNum = 1;
    let pos = 0;
    while (pos < disasmContent.length) {
      const lineEnd = disasmContent.indexOf('\n', pos);
      const end = lineEnd === -1 ? disasmContent.length : lineEnd;
      const colonPos = disasmContent.indexOf(':', pos);
      if (colonPos !== -1 && colonPos < end) {
        let valid = true;
        for (let i = pos; i < colonPos; i++) {
          const c = disasmContent[i];
          if (!((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F'))) {
            valid = false;
            break;
          }
        }
        if (valid && colonPos > pos) {
          const offset = parseInt(disasmContent.slice(pos, colonPos), 16);
          offsetToLine.set(offset, lineNum);
          lineToOffset.set(lineNum, offset);
        }
      }
      lineNum++;
      pos = end + 1;
    }
    return { offsetToLine, lineToOffset };
  }

  async function selectFunction(fn: FunctionInfo, preview = false) {
    if (!activeModule) return;
    const newSelected = `func-${fn.index}`;
    const tabId = `${activeModule.path}:func:${fn.index}`;

    if (tabsById.has(tabId)) {
      if (activeTabId !== tabId) activeTabId = tabId;
      if (selected !== newSelected) selected = newSelected;
      return;
    }

    selected = newSelected;
    const cacheKey = `${activeModule.path}:${fn.index}`;
    let cached = functionCache.get(cacheKey);

    try {
      if (!cached) {
        const [decompileResult, disasmContent] = await Promise.all([
          DecompileFunctionWithMappings(activeModule.path, fn.index),
          DisassembleFunction(activeModule.path, fn.index, false),
        ]);
        const disasmMappings = parseDisasmOffsets(disasmContent);
        const decompileMappings = decompileResult.mappings?.length
          ? buildDecompileMappings(decompileResult.mappings)
          : null;
        cached = {
          decompileCode: decompileResult.code,
          decompileMappings,
          disasmContent,
          disasmMappings,
        };
        functionCache = new Map(functionCache).set(cacheKey, cached);
      }

      const newTab: OpenTab = {
        id: tabId,
        title: fn.name,
        icon: 'f',
        type: 'function',
        modulePath: activeModule.path,
        index: fn.index,
        decompileContent: cached.decompileCode,
        decompileMappings: cached.decompileMappings,
        decompileLineCount: cached.decompileCode ? countLines(cached.decompileCode) : 0,
        disasmContent: cached.disasmContent,
        disasmMappings: cached.disasmMappings,
        disasmLineCount: countLines(cached.disasmContent),
        showLeft: settings.defaultShowDecompile,
        showRight: settings.defaultShowDisassembly,
        showOffsets: false,
        disasmIndent: false,
      };
      tabsById = new Map(tabsById).set(tabId, newTab);
      tabOrder = [...tabOrder, tabId];
      activeTabId = tabId;
    } catch (e: any) { error = e.message || String(e); }
  }

  function selectTab(id: string) {
    activeTabId = id;
    const tab = tabsById.get(id);
    if (tab) {
      selected = tab.type === 'function' ? `func-${tab.index}` : `mem-${tab.index}`;
    }
    leftHighlightLines = null;
    rightHighlightLines = null;
  }

  function closeTab(id: string) {
    const idx = tabOrder.indexOf(id);
    const newMap = new Map(tabsById);
    newMap.delete(id);
    tabsById = newMap;
    tabOrder = tabOrder.filter(tid => tid !== id);
    if (activeTabId === id) {
      activeTabId = tabOrder[Math.min(idx, tabOrder.length - 1)] || null;
    }
  }

  function closeAllTabs() {
    tabsById = new Map();
    tabOrder = [];
    activeTabId = null;
  }

  function closeOtherTabs(id: string) {
    const tab = tabsById.get(id);
    if (!tab) return;
    tabsById = new Map([[id, tab]]);
    tabOrder = [id];
    activeTabId = id;
  }

  function closeTabsToLeft(id: string) {
    const idx = tabOrder.indexOf(id);
    if (idx <= 0) return;
    const toClose = tabOrder.slice(0, idx);
    const newMap = new Map(tabsById);
    for (const tid of toClose) newMap.delete(tid);
    tabsById = newMap;
    tabOrder = tabOrder.slice(idx);
    if (!tabOrder.includes(activeTabId || '')) {
      activeTabId = tabOrder[0] || null;
    }
  }

  function closeTabsToRight(id: string) {
    const idx = tabOrder.indexOf(id);
    if (idx >= tabOrder.length - 1) return;
    const toClose = tabOrder.slice(idx + 1);
    const newMap = new Map(tabsById);
    for (const tid of toClose) newMap.delete(tid);
    tabsById = newMap;
    tabOrder = tabOrder.slice(0, idx + 1);
    if (!tabOrder.includes(activeTabId || '')) {
      activeTabId = tabOrder[tabOrder.length - 1] || null;
    }
  }

  function selectMemory(mem: MemoryInfo) {
    if (!activeModule) return;
    selected = `mem-${mem.index}`;
    const tabId = `${activeModule.path}:mem:${mem.index}`;
    if (tabsById.has(tabId)) {
      activeTabId = tabId;
      return;
    }
    const newTab: OpenTab = {
      id: tabId,
      title: `memory ${mem.index}`,
      icon: 'm',
      type: 'memory',
      modulePath: activeModule.path,
      index: mem.index,
      decompileContent: null,
      decompileMappings: null,
      decompileLineCount: 0,
      disasmContent: null,
      disasmMappings: null,
      disasmLineCount: 0,
      showLeft: false,
      showRight: false,
      showOffsets: false,
      disasmIndent: false,
    };
    tabsById = new Map(tabsById).set(tabId, newTab);
    tabOrder = [...tabOrder, tabId];
    activeTabId = tabId;
  }

  async function selectTable(tbl: TableInfo) {
    if (!activeModule) return;
    selected = `tbl-${tbl.index}`;
  }

  async function selectGlobal(glob: GlobalInfo) {
    if (!activeModule) return;
    selected = `glob-${glob.index}`;
  }

  async function selectExport(exp: ExportInfo) {
    if (exp.kind === 'func') {
      const fn = activeModule?.functionsById.get(exp.index);
      if (fn) await selectFunction(fn);
    } else if (exp.kind === 'memory') {
      const mem = activeModule?.info.memories?.find(m => m.index === exp.index);
      if (mem) await selectMemory(mem);
    } else if (exp.kind === 'table') {
      const tbl = activeModule?.info.tables?.find(t => t.index === exp.index);
      if (tbl) await selectTable(tbl);
    } else if (exp.kind === 'global') {
      const glob = activeModule?.info.globals?.find(g => g.index === exp.index);
      if (glob) await selectGlobal(glob);
    }
  }

  function gotoAddress(addr: number) {
    if (!activeModule || !activeModule.info.memories?.length) return;
    const mem = activeModule.info.memories[0];
    selectMemory(mem);
    gotoMemAddress = addr;
  }

  async function selectBookmark(bookmark: Bookmark) {
    const modIdx = modules.findIndex(m => m.path === bookmark.modulePath);
    if (modIdx < 0) return;
    activeModuleIndex = modIdx;
    const fn = modules[modIdx].functionsById.get(bookmark.funcIndex);
    if (fn) await selectFunction(fn);
  }

  async function gotoFunction(index: number) {
    if (!activeModule) return;
    const fn = activeModule.functionsById.get(index);
    if (fn) await selectFunction(fn);
  }

  async function showXRefs(funcIndex: number) {
    if (!activeModule) return;
    const fn = activeModule.functionsById.get(funcIndex);
    xrefsFuncName = fn?.name || `func_${funcIndex}`;
    try {
      xrefs = await GetXRefs(activeModule.path, funcIndex) as XRefInfo;
    } catch (e) {
      xrefs = null;
    }
  }

  function closeXRefs() {
    xrefs = null;
  }

  function getProportionalLines(startLine: number, endLine: number, fromLineCount: number, toLineCount: number): number[] {
    const startRatio = startLine / fromLineCount;
    const endRatio = endLine / fromLineCount;
    const targetStart = Math.max(1, Math.round(startRatio * toLineCount));
    const targetEnd = Math.max(1, Math.round(endRatio * toLineCount));
    const result: number[] = [];
    for (let i = targetStart; i <= targetEnd; i++) result.push(i);
    return result;
  }

  function handleLeftSelectionChange(startLine: number, endLine: number) {
    if (!activeTab?.decompileContent || !activeTab?.disasmContent) return;

    if (activeTab.decompileMappings && activeTab.disasmMappings) {
      const allOffsets = new Set<number>();
      for (let line = startLine; line <= endLine; line++) {
        const offsets = activeTab.decompileMappings.byLine.get(line);
        if (offsets) {
          for (const off of offsets) allOffsets.add(off);
        }
      }
      if (allOffsets.size > 0) {
        const disasmLines = new Set<number>();
        for (const offset of allOffsets) {
          const disasmLine = activeTab.disasmMappings.offsetToLine.get(offset);
          if (disasmLine) disasmLines.add(disasmLine);
        }
        if (disasmLines.size > 0) {
          rightHighlightLines = Array.from(disasmLines).sort((a, b) => a - b);
          leftHighlightLines = null;
          return;
        }
      }
    }
    rightHighlightLines = getProportionalLines(startLine, endLine, activeTab.decompileLineCount, activeTab.disasmLineCount);
    leftHighlightLines = null;
  }

  function handleRightSelectionChange(startLine: number, endLine: number) {
    if (!activeTab?.decompileContent || !activeTab?.disasmContent || !activeTab?.disasmMappings) return;

    if (activeTab.decompileMappings) {
      const decompileLines = new Set<number>();
      for (let line = startLine; line <= endLine; line++) {
        const offset = activeTab.disasmMappings.lineToOffset.get(line);
        if (offset !== undefined) {
          const lines = activeTab.decompileMappings.byOffset.get(offset);
          if (lines) {
            for (const l of lines) decompileLines.add(l);
          }
        }
      }
      if (decompileLines.size > 0) {
        leftHighlightLines = Array.from(decompileLines).sort((a, b) => a - b);
        rightHighlightLines = null;
        return;
      }
    }
    leftHighlightLines = getProportionalLines(startLine, endLine, activeTab.disasmLineCount, activeTab.decompileLineCount);
    rightHighlightLines = null;
  }

  async function handleToggleDisasmIndentForTab(tab: OpenTab) {
    if (tab.type !== 'function') return;
    const newIndent = !tab.disasmIndent;
    try {
      const disasmContent = await DisassembleFunction(tab.modulePath, tab.index, newIndent);
      const disasmMappings = parseDisasmOffsets(disasmContent);
      tabsById = new Map(tabsById).set(tab.id, {
        ...tab,
        disasmContent,
        disasmMappings,
        disasmLineCount: countLines(disasmContent),
        disasmIndent: newIndent,
      });
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  async function handleAddCommentForTab(tab: OpenTab, offset: number, comment: string, isDecompile: boolean) {
    if (tab.type !== 'function') return;
    const mod = modulesByPath.get(tab.modulePath);
    if (!mod) return;
    try {
      await SetOffsetComment(tab.modulePath, offset, comment, isDecompile);
      if (!isDecompile) {
        const disasmContent = await DisassembleFunction(tab.modulePath, tab.index, tab.disasmIndent);
        const disasmMappings = parseDisasmOffsets(disasmContent);
        tabsById = new Map(tabsById).set(tab.id, {
          ...tab,
          disasmContent,
          disasmMappings,
          disasmLineCount: countLines(disasmContent),
        });
        const cacheKey = `${tab.modulePath}:${tab.index}`;
        const cached = functionCache.get(cacheKey);
        if (cached) {
          functionCache = new Map(functionCache).set(cacheKey, { ...cached, disasmContent, disasmMappings });
        }
      }
      const annotations = await GetAnnotations(tab.modulePath) as Annotations;
      const modIdx = modules.findIndex(m => m.path === tab.modulePath);
      if (modIdx >= 0) {
        modules = modules.map((m, i) => i === modIdx ? { ...m, annotations } : m);
      }
      markDirty(tab.modulePath);
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  async function handleToggleDisasmIndent() {
    if (!activeTabId || !activeTab || activeTab.type !== 'function') return;
    const newIndent = !activeTab.disasmIndent;
    try {
      const disasmContent = await DisassembleFunction(activeTab.modulePath, activeTab.index, newIndent);
      const disasmMappings = parseDisasmOffsets(disasmContent);
      tabsById = new Map(tabsById).set(activeTabId, {
        ...activeTab,
        disasmContent,
        disasmMappings,
        disasmLineCount: countLines(disasmContent),
        disasmIndent: newIndent,
      });
    } catch (e: any) {
      error = e.message || String(e);
    }
  }
</script>

<svelte:window onkeydown={handleGlobalKeydown} />

<main class="h-screen select-none" style="background: var(--editor-bg);">
  <SplitPane leftWidth={288} minLeft={200} maxLeft={500} storageKey="explorerWidth">
    {#snippet left()}
      <Explorer
        {modules}
        {modulesByPath}
        {activeModuleIndex}
        {selected}
        {loading}
        {loadingName}
        {bookmarks}
        onSelectModule={selectModule}
        onCloseModule={closeModule}
        onSelectFunction={selectFunction}
        onSelectMemory={selectMemory}
        onSelectTable={selectTable}
        onSelectGlobal={selectGlobal}
        onSelectExport={selectExport}
        onOpenFile={openFile}
        onToggleBookmark={toggleBookmark}
        onSelectBookmark={selectBookmark}
        onRemoveBookmark={removeBookmark}
        isBookmarked={(path, idx) => isBookmarked(path, idx)}
        getErrorCount={(path, idx) => getFunctionErrorCount(path, idx)}
        onRenameFunction={renameFunction}
        onOpenSettings={() => showSettings = true}
      />
    {/snippet}
    {#snippet right()}
      <div class="flex flex-col h-full">
        {#if tabOrder.length > 0}
          <TabBar {tabs} activeId={activeTabId} onSelect={selectTab} onClose={closeTab} onCloseAll={closeAllTabs} onCloseOthers={closeOtherTabs} onCloseToLeft={closeTabsToLeft} onCloseToRight={closeTabsToRight} />
        {/if}
        {#if error}<div class="p-2 text-sm" style="background: color-mix(in srgb, var(--color-error) 20%, transparent); color: var(--color-error);">{error}</div>{/if}
        {#each tabs as tab (tab.id)}
          {#if tab.type === 'memory'}
            <div class="flex-1 flex overflow-hidden" style:display={tab.id === activeTabId ? 'flex' : 'none'}>
              {#key tab.id === activeTabId ? gotoMemAddress : null}
                <HexView modulePath={tab.modulePath} memIndex={tab.index} initialAddress={tab.id === activeTabId ? gotoMemAddress : null} />
              {/key}
            </div>
          {:else if tab.type === 'function'}
            <div class="flex-1 flex overflow-hidden" style:display={tab.id === activeTabId ? 'flex' : 'none'}>
              <div class="flex-1 overflow-hidden">
                <EditorSplit
                  leftContent={tab.decompileContent}
                  rightContent={tab.disasmContent}
                  showLeft={tab.showLeft}
                  showRight={tab.showRight}
                  showOffsets={tab.showOffsets}
                  disasmIndent={tab.disasmIndent}
                  onCloseLeft={() => { tabsById = new Map(tabsById).set(tab.id, { ...tab, showLeft: false }); }}
                  onCloseRight={() => { tabsById = new Map(tabsById).set(tab.id, { ...tab, showRight: false }); }}
                  onShowLeft={() => { tabsById = new Map(tabsById).set(tab.id, { ...tab, showLeft: true }); }}
                  onShowRight={() => { tabsById = new Map(tabsById).set(tab.id, { ...tab, showRight: true }); }}
                  onToggleOffsets={() => { tabsById = new Map(tabsById).set(tab.id, { ...tab, showOffsets: !tab.showOffsets }); }}
                  onToggleDisasmIndent={() => handleToggleDisasmIndentForTab(tab)}
                  onGotoAddress={gotoAddress}
                  onGotoFunction={gotoFunction}
                  onShowXRefs={showXRefs}
                  onRenameFunction={(index) => {
                    const mod = modulesByPath.get(tab.modulePath);
                    const fn = mod?.functionsById.get(index);
                    if (fn) renameFunction(fn);
                  }}
                  onAddComment={(offset, comment, isDecompile) => handleAddCommentForTab(tab, offset, comment, isDecompile)}
                  functions={modulesByPath.get(tab.modulePath)?.info.functions ?? undefined}
                  functionsByName={modulesByPath.get(tab.modulePath)?.functionsByName}
                  nicknames={tab.id === activeTabId ? nicknames : new Map()}
                  disasmComments={tab.id === activeTabId ? disasmComments : new Map()}
                  decompileComments={tab.id === activeTabId ? decompileComments : new Map()}
                  decompileMappings={tab.decompileMappings}
                  onLeftSelectionChange={tab.id === activeTabId ? handleLeftSelectionChange : undefined}
                  onRightSelectionChange={tab.id === activeTabId ? handleRightSelectionChange : undefined}
                  leftHighlightLines={tab.id === activeTabId ? leftHighlightLines : null}
                  rightHighlightLines={tab.id === activeTabId ? rightHighlightLines : null}
                  virtualizationThreshold={settings.virtualizationThreshold}
                  fontSize={settings.fontSize}
                />
              </div>
              {#if xrefs && tab.id === activeTabId}
                <div class="w-64 flex-shrink-0 text-xs overflow-auto" style="background: var(--sidebar-bg); border-left: 1px solid var(--panel-border);">
                  <div class="flex items-center justify-between px-3 py-2" style="border-bottom: 1px solid var(--panel-border);">
                    <span style="color: var(--sidebar-fg);">References</span>
                    <button class="opacity-60 hover:opacity-100" style="color: var(--sidebar-fg);" onclick={closeXRefs}>×</button>
                  </div>
                  <div class="px-3 py-2" style="color: var(--sidebar-fg);">
                    <div class="truncate mb-3" style="color: var(--syntax-function);">{xrefsFuncName}</div>
                    {#if xrefs.callers.length > 0}
                      <div class="mb-3">
                        <div class="opacity-60 mb-1">Called by ({xrefs.callers.length})</div>
                        {#each xrefs.callers as caller}
                          <button
                            class="block hover:underline truncate text-left w-full py-0.5"
                            style="color: var(--syntax-function);"
                            onclick={() => gotoFunction(caller.index)}
                          >{caller.name}</button>
                        {/each}
                      </div>
                    {/if}
                    {#if xrefs.callees.length > 0}
                      <div>
                        <div class="opacity-60 mb-1">Calls ({xrefs.callees.length})</div>
                        {#each xrefs.callees as callee}
                          <button
                            class="block hover:underline truncate text-left w-full py-0.5"
                            style="color: var(--syntax-function);"
                            onclick={() => gotoFunction(callee.index)}
                          >{callee.name}</button>
                        {/each}
                      </div>
                    {/if}
                    {#if xrefs.callers.length === 0 && xrefs.callees.length === 0}
                      <div class="opacity-60">No references found</div>
                    {/if}
                  </div>
                </div>
              {/if}
            </div>
          {/if}
        {/each}
        {#if tabs.length === 0}
          <div class="flex-1 flex items-center justify-center" style="color: var(--sidebar-fg); opacity: 0.5;">
            {#if modules.length === 0}
              Drop a .wasm file or open from explorer
            {:else}
              Select an item to view
            {/if}
          </div>
        {/if}
      </div>
    {/snippet}
  </SplitPane>
</main>

{#if showCommandPalette && activeModule?.info.functions}
  <CommandPalette
    functions={activeModule.info.functions}
    onSelect={selectFunction}
    onClose={() => showCommandPalette = false}
  />
{/if}

{#if renameTarget}
  <RenameModal
    currentName={nicknames.get(renameTarget.index) || renameTarget.name}
    onConfirm={handleRenameConfirm}
    onCancel={handleRenameCancel}
  />
{/if}

{#if showQuitConfirm}
  <div class="fixed inset-0 flex items-center justify-center z-50" style="background: rgba(0,0,0,0.5);">
    <div class="p-4 rounded-lg shadow-xl" style="background: var(--sidebar-bg); border: 1px solid var(--panel-border); min-width: 320px;">
      <div class="text-sm mb-4" style="color: var(--sidebar-fg);">
        You have unsaved changes. Do you want to save before quitting?
      </div>
      <div class="flex justify-end gap-2">
        <button
          class="px-3 py-1.5 text-xs rounded"
          style="background: var(--panel-bg); color: var(--sidebar-fg); border: 1px solid var(--panel-border);"
          onclick={() => showQuitConfirm = false}
        >Cancel</button>
        <button
          class="px-3 py-1.5 text-xs rounded"
          style="background: var(--color-error); color: white;"
          onclick={handleQuitWithoutSave}
        >Don't Save</button>
        <button
          class="px-3 py-1.5 text-xs rounded"
          style="background: var(--button-active); color: white;"
          onclick={handleQuitWithSave}
        >Save & Quit</button>
      </div>
    </div>
  </div>
{/if}

{#if showSettings}
  <SettingsModal
    {settings}
    onSave={handleSaveSettings}
    onClose={() => showSettings = false}
  />
{/if}
