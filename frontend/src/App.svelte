<script lang="ts">
  import { LoadModuleFromPath, DisassembleFunction, DecompileFunction, OpenFileDialog, GetXRefs, GetModuleErrors } from '../wailsjs/go/main/App';
  import { EventsOn } from '../wailsjs/runtime/runtime';
  import type { ModuleInfo, FunctionInfo, MemoryInfo, TableInfo, GlobalInfo, ExportInfo, Bookmark, XRefInfo, ModuleErrorsInfo } from './lib/types';
  import Explorer from './lib/Explorer.svelte';
  import Editor from './lib/Editor.svelte';
  import HexView from './lib/HexView.svelte';
  import SplitPane from './lib/SplitPane.svelte';
  import TabBar, { type Tab } from './lib/TabBar.svelte';
  import CommandPalette from './lib/CommandPalette.svelte';
  import { applyTheme, defaultTheme } from './lib/themes';

  applyTheme(defaultTheme);

  interface LoadedModule {
    path: string;
    name: string;
    info: ModuleInfo;
  }

  interface OpenTab extends Tab {
    type: 'function' | 'memory';
    modulePath: string;
    index: number;
    content: string;
    mode: 'disasm' | 'decompile';
  }

  let modules: LoadedModule[] = $state([]);
  let activeModuleIndex: number = $state(-1);
  let error = $state('');
  let loading = $state(false);
  let loadingName = $state('');
  let selected = $state('');
  let mode: 'disasm' | 'decompile' = $state('decompile');
  let tabs: OpenTab[] = $state([]);
  let activeTabId: string | null = $state(null);
  let gotoMemAddress: number | null = $state(null);
  let showCommandPalette = $state(false);
  let bookmarks: Bookmark[] = $state(loadBookmarks());
  let xrefs: XRefInfo | null = $state(null);
  let xrefsFuncName: string = $state('');
  let functionErrors: Map<string, Map<number, number>> = $state(new Map());

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

  function addBookmark(fn: FunctionInfo) {
    if (!activeModule) return;
    const id = `${activeModule.path}:${fn.index}`;
    if (bookmarks.some(b => b.id === id)) return;
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
    return bookmarks.some(b => b.modulePath === modulePath && b.funcIndex === funcIndex);
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
      if (tabs.length > 1 && activeTabId) {
        const idx = tabs.findIndex(t => t.id === activeTabId);
        const nextIdx = e.shiftKey
          ? (idx - 1 + tabs.length) % tabs.length
          : (idx + 1) % tabs.length;
        selectTab(tabs[nextIdx].id);
      }
    }
    if ((e.ctrlKey || e.metaKey) && e.key === 'b') {
      e.preventDefault();
      if (activeTab?.type === 'function' && activeModule) {
        const fn = activeModule.info.functions?.find(f => f.index === activeTab.index);
        if (fn) toggleBookmark(fn);
      }
    }
  }

  let activeModule = $derived(activeModuleIndex >= 0 ? modules[activeModuleIndex] : null);
  let activeTab = $derived(tabs.find(t => t.id === activeTabId) || null);

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
      modules = [...modules, { path, name: loadingName, info }];
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

  async function selectFunction(fn: FunctionInfo) {
    if (!activeModule) return;
    selected = `func-${fn.index}`;
    const tabId = `${activeModule.path}:func:${fn.index}`;
    const existing = tabs.find(t => t.id === tabId);
    if (existing) {
      activeTabId = tabId;
      return;
    }
    try {
      const content = mode === 'decompile'
        ? await DecompileFunction(activeModule.path, fn.index)
        : await DisassembleFunction(activeModule.path, fn.index);
      const newTab: OpenTab = {
        id: tabId,
        title: fn.name,
        icon: 'f',
        type: 'function',
        modulePath: activeModule.path,
        index: fn.index,
        content,
        mode,
      };
      tabs = [...tabs, newTab];
      activeTabId = tabId;
    } catch (e: any) { error = e.message || String(e); }
  }

  async function switchMode(newMode: 'disasm' | 'decompile') {
    mode = newMode;
    if (activeTab && activeTab.type === 'function') {
      try {
        const content = newMode === 'decompile'
          ? await DecompileFunction(activeTab.modulePath, activeTab.index)
          : await DisassembleFunction(activeTab.modulePath, activeTab.index);
        tabs = tabs.map(t => t.id === activeTabId ? { ...t, content, mode: newMode } : t);
      } catch (e: any) { error = e.message || String(e); }
    }
  }

  function selectTab(id: string) {
    activeTabId = id;
    const tab = tabs.find(t => t.id === id);
    if (tab) {
      selected = tab.type === 'function' ? `func-${tab.index}` : `mem-${tab.index}`;
      mode = tab.mode;
    }
  }

  function closeTab(id: string) {
    const idx = tabs.findIndex(t => t.id === id);
    tabs = tabs.filter(t => t.id !== id);
    if (activeTabId === id) {
      activeTabId = tabs[Math.min(idx, tabs.length - 1)]?.id || null;
    }
  }

  function selectMemory(mem: MemoryInfo) {
    if (!activeModule) return;
    selected = `mem-${mem.index}`;
    const tabId = `${activeModule.path}:mem:${mem.index}`;
    const existing = tabs.find(t => t.id === tabId);
    if (existing) {
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
      content: '',
      mode: 'disasm',
    };
    tabs = [...tabs, newTab];
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
      const fn = activeModule?.info.functions?.find(f => f.index === exp.index);
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
    const fn = modules[modIdx].info.functions?.find(f => f.index === bookmark.funcIndex);
    if (fn) await selectFunction(fn);
  }

  async function gotoFunction(index: number) {
    if (!activeModule) return;
    const fn = activeModule.info.functions?.find(f => f.index === index);
    if (fn) await selectFunction(fn);
  }

  async function showXRefs(funcIndex: number) {
    if (!activeModule) return;
    const fn = activeModule.info.functions?.find(f => f.index === funcIndex);
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
</script>

<svelte:window onkeydown={handleGlobalKeydown} />

<main class="h-screen select-none" style="background: var(--editor-bg);">
  <SplitPane leftWidth={288} minLeft={200} maxLeft={500} storageKey="explorerWidth">
    {#snippet left()}
      <Explorer
        {modules}
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
        {#if tabs.length > 0}
          <TabBar {tabs} activeId={activeTabId} onSelect={selectTab} onClose={closeTab} />
        {/if}
        {#if activeTab?.type === 'function'}
          <div class="flex gap-1 p-2" style="background: var(--panel-bg); border-bottom: 1px solid var(--panel-border);">
            <button
              class="px-3 py-1 text-xs rounded"
              style="background: {activeTab.mode === 'disasm' ? 'var(--button-active)' : 'var(--button-bg)'}; color: var(--sidebar-fg);"
              onclick={() => switchMode('disasm')}
            >Disassembly</button>
            <button
              class="px-3 py-1 text-xs rounded"
              style="background: {activeTab.mode === 'decompile' ? 'var(--button-active)' : 'var(--button-bg)'}; color: var(--sidebar-fg);"
              onclick={() => switchMode('decompile')}
            >Decompile</button>
          </div>
        {/if}
        {#if error}<div class="p-2 text-sm" style="background: color-mix(in srgb, var(--color-error) 20%, transparent); color: var(--color-error);">{error}</div>{/if}
        {#if activeTab?.type === 'memory'}
          {#key gotoMemAddress}
            <HexView modulePath={activeTab.modulePath} memIndex={activeTab.index} initialAddress={gotoMemAddress} />
          {/key}
        {:else if activeTab?.content}
          <div class="flex-1 flex overflow-hidden">
            <div class="flex-1 overflow-hidden">
              <Editor content={activeTab.content} mode={activeTab.mode} onGotoAddress={gotoAddress} onGotoFunction={gotoFunction} onShowXRefs={showXRefs} functions={activeModule?.info.functions} />
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
