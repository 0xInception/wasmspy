<script lang="ts">
  import { LoadModuleFromPath, DisassembleFunction, DecompileFunctionWithMappings, GetFunctionWAT, OpenFileDialog, GetXRefs, GetModuleErrors } from '../wailsjs/go/main/App';
  import { EventsOn } from '../wailsjs/runtime/runtime';
  import type { ModuleInfo, FunctionInfo, MemoryInfo, TableInfo, GlobalInfo, ExportInfo, Bookmark, XRefInfo, ModuleErrorsInfo } from './lib/types';
  import Explorer from './lib/Explorer.svelte';
  import EditorSplit from './lib/EditorSplit.svelte';
  import HexView from './lib/HexView.svelte';
  import SplitPane from './lib/SplitPane.svelte';
  import TabBar, { type Tab } from './lib/TabBar.svelte';
  import CommandPalette from './lib/CommandPalette.svelte';
  import { applyTheme, defaultTheme } from './lib/themes';

  applyTheme(defaultTheme);

  type GroupedFunctions = [string, FunctionInfo[]][];

  interface LoadedModule {
    path: string;
    name: string;
    info: ModuleInfo;
    functionsById: Map<number, FunctionInfo>;
    functionsByName: Map<string, FunctionInfo>;
    groupedFunctions: GroupedFunctions;
    groupedImports: GroupedFunctions;
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
  let bookmarks: Bookmark[] = $state(loadBookmarks());
  let xrefs: XRefInfo | null = $state(null);
  let xrefsFuncName: string = $state('');
  let functionErrors: Map<string, Map<number, number>> = $state(new Map());
  let leftHighlightLines: number[] | null = $state(null);
  let rightHighlightLines: number[] | null = $state(null);

  function getFunctionErrorCount(modulePath: string, funcIndex: number): number {
    return functionErrors.get(modulePath)?.get(funcIndex) ?? 0;
  }

  function loadBookmarks(): Bookmark[] {
    try {
      const stored = localStorage.getItem('wasmspy-bookmarks');
      return stored ? JSON.parse(stored) : [];
    } catch { return []; }
  }

  function saveBookmarks() {
    localStorage.setItem('wasmspy-bookmarks', JSON.stringify(bookmarks));
  }

  let bookmarkIds = $derived(new Set(bookmarks.map(b => b.id)));

  function addBookmark(fn: FunctionInfo) {
    if (!activeModule) return;
    const id = `${activeModule.path}:${fn.index}`;
    if (bookmarkIds.has(id)) return;
    bookmarks = [...bookmarks, {
      id,
      modulePath: activeModule.path,
      funcIndex: fn.index,
      funcName: fn.name,
    }];
    saveBookmarks();
  }

  function removeBookmark(id: string) {
    bookmarks = bookmarks.filter(b => b.id !== id);
    saveBookmarks();
  }

  function isBookmarked(modulePath: string, funcIndex: number): boolean {
    return bookmarkIds.has(`${modulePath}:${funcIndex}`);
  }

  function toggleBookmark(fn: FunctionInfo) {
    if (!activeModule) return;
    const id = `${activeModule.path}:${fn.index}`;
    if (isBookmarked(activeModule.path, fn.index)) {
      removeBookmark(id);
    } else {
      addBookmark(fn);
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
      const info = await LoadModuleFromPath(path) as ModuleInfo;
      const { byId, byName } = buildFunctionMaps(info.functions);
      const groupedFunctions = buildGroupedFunctions(info.functions, false);
      const groupedImports = buildGroupedFunctions(info.functions, true);
      modules = [...modules, { path, name: loadingName, info, functionsById: byId, functionsByName: byName, groupedFunctions, groupedImports }];
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

  async function selectFunction(fn: FunctionInfo) {
    if (!activeModule) return;
    selected = `func-${fn.index}`;
    const tabId = `${activeModule.path}:func:${fn.index}`;
    if (tabsById.has(tabId)) {
      activeTabId = tabId;
      return;
    }
    try {
      const [decompileResult, disasmContent] = await Promise.all([
        DecompileFunctionWithMappings(activeModule.path, fn.index),
        DisassembleFunction(activeModule.path, fn.index, false),
      ]);
      const disasmMappings = parseDisasmOffsets(disasmContent);
      const decompileMappings = decompileResult.mappings?.length
        ? buildDecompileMappings(decompileResult.mappings)
        : null;
      const newTab: OpenTab = {
        id: tabId,
        title: fn.name,
        icon: 'f',
        type: 'function',
        modulePath: activeModule.path,
        index: fn.index,
        decompileContent: decompileResult.code,
        decompileMappings,
        decompileLineCount: decompileResult.code ? countLines(decompileResult.code) : 0,
        disasmContent,
        disasmMappings,
        disasmLineCount: countLines(disasmContent),
        showLeft: true,
        showRight: true,
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
      />
    {/snippet}
    {#snippet right()}
      <div class="flex flex-col h-full">
        {#if tabOrder.length > 0}
          <TabBar {tabs} activeId={activeTabId} onSelect={selectTab} onClose={closeTab} />
        {/if}
        {#if error}<div class="p-2 text-sm" style="background: color-mix(in srgb, var(--color-error) 20%, transparent); color: var(--color-error);">{error}</div>{/if}
        {#if activeTab?.type === 'memory'}
          {#key gotoMemAddress}
            <HexView modulePath={activeTab.modulePath} memIndex={activeTab.index} initialAddress={gotoMemAddress} />
          {/key}
        {:else if activeTab?.type === 'function'}
          <div class="flex-1 flex overflow-hidden">
            <div class="flex-1 overflow-hidden">
              <EditorSplit
                leftContent={activeTab.decompileContent}
                rightContent={activeTab.disasmContent}
                showLeft={activeTab.showLeft}
                showRight={activeTab.showRight}
                showOffsets={activeTab.showOffsets}
                disasmIndent={activeTab.disasmIndent}
                onCloseLeft={() => { if (activeTabId && activeTab) { tabsById = new Map(tabsById).set(activeTabId, { ...activeTab, showLeft: false }); } }}
                onCloseRight={() => { if (activeTabId && activeTab) { tabsById = new Map(tabsById).set(activeTabId, { ...activeTab, showRight: false }); } }}
                onShowLeft={() => { if (activeTabId && activeTab) { tabsById = new Map(tabsById).set(activeTabId, { ...activeTab, showLeft: true }); } }}
                onShowRight={() => { if (activeTabId && activeTab) { tabsById = new Map(tabsById).set(activeTabId, { ...activeTab, showRight: true }); } }}
                onToggleOffsets={() => { if (activeTabId && activeTab) { tabsById = new Map(tabsById).set(activeTabId, { ...activeTab, showOffsets: !activeTab.showOffsets }); } }}
                onToggleDisasmIndent={handleToggleDisasmIndent}
                onGotoAddress={gotoAddress}
                onGotoFunction={gotoFunction}
                onShowXRefs={showXRefs}
                functions={activeModule?.info.functions ?? undefined}
                functionsByName={activeModule?.functionsByName}
                onLeftSelectionChange={handleLeftSelectionChange}
                onRightSelectionChange={handleRightSelectionChange}
                {leftHighlightLines}
                {rightHighlightLines}
              />
            </div>
            {#if xrefs}
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
        {:else}
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
