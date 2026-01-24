<script lang="ts">
  import { LoadModuleFromPath, DisassembleFunction, DecompileFunction, GetTable, GetGlobal, OpenFileDialog } from '../wailsjs/go/main/App';
  import { EventsOn } from '../wailsjs/runtime/runtime';
  import type { ModuleInfo, FunctionInfo, MemoryInfo, TableInfo, GlobalInfo, ExportInfo } from './lib/types';
  import Explorer from './lib/Explorer.svelte';
  import Editor from './lib/Editor.svelte';
  import HexView from './lib/HexView.svelte';

  interface LoadedModule {
    path: string;
    name: string;
    info: ModuleInfo;
  }

  let modules: LoadedModule[] = $state([]);
  let activeModuleIndex: number = $state(-1);
  let output = $state('');
  let error = $state('');
  let loading = $state(false);
  let loadingName = $state('');
  let selected = $state('');
  let mode: 'disasm' | 'decompile' = $state('disasm');
  let selectedFn: FunctionInfo | null = $state(null);
  let selectedMemIndex: number | null = $state(null);
  let gotoMemAddress: number | null = $state(null);

  let activeModule = $derived(activeModuleIndex >= 0 ? modules[activeModuleIndex] : null);

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
      output = '';
      selected = '';
      selectedFn = null;
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
    output = '';
    selected = '';
    selectedFn = null;
    selectedMemIndex = null;
  }

  function closeModule(index: number) {
    modules = modules.filter((_, i) => i !== index);
    if (activeModuleIndex >= modules.length) {
      activeModuleIndex = modules.length - 1;
    }
    if (activeModuleIndex < 0) {
      output = '';
      selected = '';
      selectedFn = null;
      selectedMemIndex = null;
    }
  }

  async function selectFunction(fn: FunctionInfo) {
    if (!activeModule) return;
    selected = `func-${fn.index}`;
    selectedFn = fn;
    selectedMemIndex = null;
    try {
      output = mode === 'decompile'
        ? await DecompileFunction(activeModule.path, fn.index)
        : await DisassembleFunction(activeModule.path, fn.index);
    } catch (e: any) { error = e.message || String(e); }
  }

  async function switchMode(newMode: 'disasm' | 'decompile') {
    mode = newMode;
    if (selectedFn) await selectFunction(selectedFn);
  }

  function selectMemory(mem: MemoryInfo) {
    if (!activeModule) return;
    selected = `mem-${mem.index}`;
    selectedFn = null;
    selectedMemIndex = mem.index;
    output = '';
  }

  async function selectTable(tbl: TableInfo) {
    if (!activeModule) return;
    selected = `tbl-${tbl.index}`;
    selectedFn = null;
    selectedMemIndex = null;
    try { output = await GetTable(activeModule.path, tbl.index); }
    catch (e: any) { error = e.message || String(e); }
  }

  async function selectGlobal(glob: GlobalInfo) {
    if (!activeModule) return;
    selected = `glob-${glob.index}`;
    selectedFn = null;
    selectedMemIndex = null;
    try { output = await GetGlobal(activeModule.path, glob.index); }
    catch (e: any) { error = e.message || String(e); }
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
    selected = 'mem-0';
    selectedFn = null;
    selectedMemIndex = 0;
    gotoMemAddress = addr;
    output = '';
  }
</script>

<main class="flex h-screen select-none">
  <Explorer
    {modules}
    {activeModuleIndex}
    {selected}
    {loading}
    {loadingName}
    onSelectModule={selectModule}
    onCloseModule={closeModule}
    onSelectFunction={selectFunction}
    onSelectMemory={selectMemory}
    onSelectTable={selectTable}
    onSelectGlobal={selectGlobal}
    onSelectExport={selectExport}
    onOpenFile={openFile}
  />
  <div class="flex-1 flex flex-col">
    {#if selectedFn}
      <div class="flex gap-1 p-2 border-b border-gray-700 bg-gray-900">
        <button
          class="px-3 py-1 text-xs rounded {mode === 'disasm' ? 'bg-blue-600' : 'bg-gray-700 hover:bg-gray-600'}"
          onclick={() => switchMode('disasm')}
        >Disassembly</button>
        <button
          class="px-3 py-1 text-xs rounded {mode === 'decompile' ? 'bg-blue-600' : 'bg-gray-700 hover:bg-gray-600'}"
          onclick={() => switchMode('decompile')}
        >Decompile</button>
      </div>
    {/if}
    {#if error}<div class="p-2 bg-red-900/50 text-red-300 text-sm">{error}</div>{/if}
    {#if selectedMemIndex !== null && activeModule}
      {#key gotoMemAddress}
        <HexView modulePath={activeModule.path} memIndex={selectedMemIndex} initialAddress={gotoMemAddress} />
      {/key}
    {:else if output}
      <Editor content={output} onGotoAddress={gotoAddress} />
    {:else}
      <div class="flex-1 flex items-center justify-center text-gray-500">
        {#if modules.length === 0}
          Drop a .wasm file or open from explorer
        {:else}
          Select an item to view
        {/if}
      </div>
    {/if}
  </div>
</main>
