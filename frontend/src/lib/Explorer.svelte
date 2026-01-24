<script lang="ts">
  import type { ModuleInfo, FunctionInfo, MemoryInfo, TableInfo, GlobalInfo, ExportInfo, Bookmark } from './types';
  import ContextMenu, { type MenuItem } from './ContextMenu.svelte';

  interface LoadedModule {
    path: string;
    name: string;
    info: ModuleInfo;
  }

  let {
    modules,
    activeModuleIndex,
    selected,
    loading,
    loadingName,
    bookmarks,
    onSelectModule,
    onCloseModule,
    onSelectFunction,
    onSelectMemory,
    onSelectTable,
    onSelectGlobal,
    onSelectExport,
    onOpenFile,
    onToggleBookmark,
    onSelectBookmark,
    onRemoveBookmark,
    isBookmarked,
    getErrorCount,
  }: {
    modules: LoadedModule[];
    activeModuleIndex: number;
    selected: string;
    loading: boolean;
    loadingName: string;
    bookmarks: Bookmark[];
    onSelectModule: (index: number) => void;
    onCloseModule: (index: number) => void;
    onSelectFunction: (fn: FunctionInfo) => void;
    onSelectMemory: (mem: MemoryInfo) => void;
    onSelectTable: (tbl: TableInfo) => void;
    onSelectGlobal: (glob: GlobalInfo) => void;
    onSelectExport: (exp: ExportInfo) => void;
    onOpenFile: () => void;
    onToggleBookmark: (fn: FunctionInfo) => void;
    onSelectBookmark: (bookmark: Bookmark) => void;
    onRemoveBookmark: (id: string) => void;
    isBookmarked: (modulePath: string, funcIndex: number) => boolean;
    getErrorCount: (modulePath: string, funcIndex: number) => number;
  } = $props();

  let expanded: Record<string, boolean> = $state({});
  let expandedGroups: Record<string, boolean> = $state({});
  let contextMenu: { x: number; y: number; items: MenuItem[] } | null = $state(null);
  let searchQuery = $state('');

  function matchesSearch(name: string): boolean {
    if (!searchQuery) return true;
    return name.toLowerCase().includes(searchQuery.toLowerCase());
  }

  function hasMatchingFunctions(fns: FunctionInfo[]): boolean {
    if (!searchQuery) return true;
    return fns.some(fn => matchesSearch(fn.name));
  }

  function getFilteredFunctions(fns: FunctionInfo[]): FunctionInfo[] {
    if (!searchQuery) return fns;
    return fns.filter(fn => matchesSearch(fn.name));
  }

  function showFnContextMenu(e: MouseEvent, fn: FunctionInfo, modulePath: string) {
    e.preventDefault();
    const bookmarked = isBookmarked(modulePath, fn.index);
    contextMenu = {
      x: e.clientX,
      y: e.clientY,
      items: [
        { label: bookmarked ? 'Remove bookmark' : 'Bookmark', action: () => onToggleBookmark(fn) },
        { label: 'Copy name', action: () => navigator.clipboard.writeText(fn.name) },
        { label: 'Copy index', action: () => navigator.clipboard.writeText(String(fn.index)) },
      ]
    };
  }

  function showBookmarkContextMenu(e: MouseEvent, bookmark: Bookmark) {
    e.preventDefault();
    contextMenu = {
      x: e.clientX,
      y: e.clientY,
      items: [
        { label: 'Remove bookmark', action: () => onRemoveBookmark(bookmark.id) },
        { label: 'Copy name', action: () => navigator.clipboard.writeText(bookmark.funcName) },
      ]
    };
  }

  function toggle(key: string) { expanded[key] = !expanded[key]; }
  function toggleGroup(key: string) { expandedGroups[key] = !expandedGroups[key]; }

  function getPrefix(name: string): string {
    const dot = name.indexOf('.');
    return dot > 0 ? name.slice(0, dot) : '_ungrouped';
  }

  function getGroupedFunctions(moduleInfo: ModuleInfo, imported: boolean) {
    if (!moduleInfo?.functions) return [];
    const groups = new Map<string, FunctionInfo[]>();
    for (const fn of moduleInfo.functions) {
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

  function countFunctions(moduleInfo: ModuleInfo, imported: boolean): number {
    return moduleInfo?.functions?.filter(f => f.imported === imported).length ?? 0;
  }

  let expandedBookmarks = $state(true);

  let activeBookmarks = $derived(
    bookmarks.filter(b => modules.some(m => m.path === b.modulePath))
  );
</script>

<aside class="h-full overflow-auto text-sm flex flex-col" style="background: var(--sidebar-bg); border-right: 1px solid var(--sidebar-border); color: var(--sidebar-fg);">
  <div class="p-2 text-xs uppercase tracking-wide flex items-center justify-between" style="border-bottom: 1px solid var(--sidebar-border); color: var(--sidebar-fg); opacity: 0.7;">
    <span>Explorer</span>
    <button class="px-1 opacity-60 hover:opacity-100" onclick={onOpenFile} title="Open file">+</button>
  </div>
  {#if modules.length > 0}
    <div class="px-2 py-1.5" style="border-bottom: 1px solid var(--sidebar-border);">
      <div class="relative">
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Filter functions..."
          class="w-full pl-2 pr-7 py-1 text-xs rounded outline-none"
          style="background: var(--input-bg); border: 1px solid var(--input-border); color: var(--editor-fg);"
        />
        {#if searchQuery}
          <button
            class="absolute right-1.5 top-1/2 -translate-y-1/2 text-xs opacity-50 hover:opacity-100"
            onclick={() => searchQuery = ''}
          >×</button>
        {/if}
      </div>
    </div>
  {/if}
  <div class="p-1 flex-1 overflow-auto">
    {#if activeBookmarks.length > 0}
      <div class="mb-1">
        <button
          class="flex items-center gap-1 w-full px-2 py-1 hover:bg-gray-800 rounded text-left"
          onclick={() => expandedBookmarks = !expandedBookmarks}
        >
          <span class="text-gray-500 w-3">{expandedBookmarks ? '▼' : '▶'}</span>
          <span style="color: var(--color-warning);">★</span>
          <span>Bookmarks ({activeBookmarks.length})</span>
        </button>
        {#if expandedBookmarks}
          <div class="ml-4">
            {#each activeBookmarks as bookmark}
              {@const mod = modules.find(m => m.path === bookmark.modulePath)}
              <button
                class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs"
                onclick={() => onSelectBookmark(bookmark)}
                oncontextmenu={(e) => showBookmarkContextMenu(e, bookmark)}
              >
                <span class="w-3" style="color: var(--icon-function);">f</span>
                <span class="truncate text-gray-300">{bookmark.funcName}</span>
                {#if modules.length > 1 && mod}
                  <span class="text-gray-600 text-[10px] truncate">({mod.name})</span>
                {/if}
              </button>
            {/each}
          </div>
        {/if}
      </div>
    {/if}

    {#if loading}
      <div class="px-2 py-1 text-gray-400">Loading {loadingName}...</div>
    {/if}

    {#if modules.length === 0 && !loading}
      <div class="px-2 py-4 text-gray-500 text-center">
        <p class="mb-2">No module loaded</p>
        <button class="text-blue-400 hover:text-blue-300" onclick={onOpenFile}>Open file</button>
        <p class="mt-2 text-xs">or drop a .wasm file</p>
      </div>
    {/if}

    {#each modules as mod, modIndex}
      {@const isActive = modIndex === activeModuleIndex}
      {@const modKey = `mod-${modIndex}`}
      {@const isExpanded = expanded[modKey] ?? isActive}
      {@const importedFunctions = getGroupedFunctions(mod.info, true)}
      {@const definedFunctions = getGroupedFunctions(mod.info, false)}

      <div class="group">
        <button
          class="flex items-center gap-1 w-full px-2 py-1 hover:bg-gray-800 rounded text-left {isActive ? 'bg-gray-800' : ''}"
          onclick={() => { onSelectModule(modIndex); toggle(modKey); }}
        >
          <span class="text-gray-500 w-3">{isExpanded ? '▼' : '▶'}</span>
          <span class="truncate flex-1">{mod.name}</span>
          <span
            class="text-gray-600 hover:text-gray-300 opacity-0 group-hover:opacity-100 px-1"
            onclick={(e) => { e.stopPropagation(); onCloseModule(modIndex); }}
          >×</span>
        </button>

        {#if isExpanded}
          <div class="ml-4">
            <button class="flex items-center gap-1 w-full px-2 py-1 hover:bg-gray-800 rounded text-left" onclick={() => toggle(`${modKey}-functions`)}>
              <span class="text-gray-500 w-3">{expanded[`${modKey}-functions`] ? '▼' : '▶'}</span>
              <span>Functions ({countFunctions(mod.info, false)})</span>
            </button>
            {#if expanded[`${modKey}-functions`] || searchQuery}
              <div class="ml-4">
                {#each definedFunctions as [group, fns]}
                  {@const groupKey = `${modKey}-grp-${group}`}
                  {@const filteredFns = getFilteredFunctions(fns)}
                  {#if filteredFns.length > 0}
                    <button
                      class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs"
                      onclick={() => toggleGroup(groupKey)}
                    >
                      <span class="text-gray-500 w-3">{expandedGroups[groupKey] || searchQuery ? '▼' : '▶'}</span>
                      <span class="truncate" style="color: var(--icon-group);">{group}</span>
                      <span class="text-gray-500">({filteredFns.length}{searchQuery && filteredFns.length !== fns.length ? `/${fns.length}` : ''})</span>
                    </button>
                    {#if expandedGroups[groupKey] || searchQuery}
                      <div class="ml-4">
                        {#each filteredFns as fn}
                          {@const fnBookmarked = isBookmarked(mod.path, fn.index)}
                          {@const errCount = getErrorCount(mod.path, fn.index)}
                          <button
                            class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs {isActive && selected === `func-${fn.index}` ? 'bg-blue-600/30' : ''}"
                            onclick={() => { onSelectModule(modIndex); onSelectFunction(fn); }}
                            oncontextmenu={(e) => showFnContextMenu(e, fn, mod.path)}
                          >
                            <span class="w-3" style="color: var(--icon-function);">{fnBookmarked ? '★' : 'f'}</span>
                            <span class="truncate text-gray-300 flex-1">{fn.name.slice(group.length + 1) || fn.name}</span>
                            {#if errCount > 0}
                              <span class="px-1 rounded text-[10px]" style="background: var(--color-error); color: white;" title="{errCount} error{errCount > 1 ? 's' : ''}">{errCount}</span>
                            {/if}
                          </button>
                        {/each}
                      </div>
                    {/if}
                  {/if}
                {/each}
              </div>
            {/if}

            <button class="flex items-center gap-1 w-full px-2 py-1 hover:bg-gray-800 rounded text-left" onclick={() => toggle(`${modKey}-imports`)}>
              <span class="text-gray-500 w-3">{expanded[`${modKey}-imports`] ? '▼' : '▶'}</span>
              <span>Imports ({countFunctions(mod.info, true)})</span>
            </button>
            {#if expanded[`${modKey}-imports`] || searchQuery}
              <div class="ml-4">
                {#each importedFunctions as [group, fns]}
                  {@const groupKey = `${modKey}-imp-${group}`}
                  {@const filteredFns = getFilteredFunctions(fns)}
                  {#if filteredFns.length > 0}
                    <button
                      class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs"
                      onclick={() => toggleGroup(groupKey)}
                    >
                      <span class="text-gray-500 w-3">{expandedGroups[groupKey] || searchQuery ? '▼' : '▶'}</span>
                      <span class="truncate" style="color: var(--icon-group);">{group}</span>
                      <span class="text-gray-500">({filteredFns.length}{searchQuery && filteredFns.length !== fns.length ? `/${fns.length}` : ''})</span>
                    </button>
                    {#if expandedGroups[groupKey] || searchQuery}
                      <div class="ml-4">
                        {#each filteredFns as fn}
                          {@const fnBookmarked = isBookmarked(mod.path, fn.index)}
                          <button
                            class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs {isActive && selected === `func-${fn.index}` ? 'bg-blue-600/30' : ''}"
                            onclick={() => { onSelectModule(modIndex); onSelectFunction(fn); }}
                            oncontextmenu={(e) => showFnContextMenu(e, fn, mod.path)}
                          >
                            <span class="w-3" style="color: var(--icon-import);">{fnBookmarked ? '★' : 'f'}</span>
                            <span class="truncate text-gray-300">{fn.name.slice(group.length + 1) || fn.name}</span>
                          </button>
                        {/each}
                      </div>
                    {/if}
                  {/if}
                {/each}
              </div>
            {/if}

            {#if mod.info.memories?.length}
              <button class="flex items-center gap-1 w-full px-2 py-1 hover:bg-gray-800 rounded text-left" onclick={() => toggle(`${modKey}-memories`)}>
                <span class="text-gray-500 w-3">{expanded[`${modKey}-memories`] ? '▼' : '▶'}</span>
                <span>Memories ({mod.info.memories.length})</span>
              </button>
              {#if expanded[`${modKey}-memories`]}
                <div class="ml-4">
                  {#each mod.info.memories as mem}
                    <button
                      class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs {isActive && selected === `mem-${mem.index}` ? 'bg-blue-600/30' : ''}"
                      onclick={() => { onSelectModule(modIndex); onSelectMemory(mem); }}
                    >
                      <span class="w-3" style="color: var(--icon-memory);">m</span>
                      <span class="text-gray-300">memory {mem.index}</span>
                    </button>
                  {/each}
                </div>
              {/if}
            {/if}

            {#if mod.info.tables?.length}
              <button class="flex items-center gap-1 w-full px-2 py-1 hover:bg-gray-800 rounded text-left" onclick={() => toggle(`${modKey}-tables`)}>
                <span class="text-gray-500 w-3">{expanded[`${modKey}-tables`] ? '▼' : '▶'}</span>
                <span>Tables ({mod.info.tables.length})</span>
              </button>
              {#if expanded[`${modKey}-tables`]}
                <div class="ml-4">
                  {#each mod.info.tables as tbl}
                    <button
                      class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs {isActive && selected === `tbl-${tbl.index}` ? 'bg-blue-600/30' : ''}"
                      onclick={() => { onSelectModule(modIndex); onSelectTable(tbl); }}
                    >
                      <span class="w-3" style="color: var(--icon-table);">t</span>
                      <span class="text-gray-300">table {tbl.index}</span>
                    </button>
                  {/each}
                </div>
              {/if}
            {/if}

            {#if mod.info.globals?.length}
              <button class="flex items-center gap-1 w-full px-2 py-1 hover:bg-gray-800 rounded text-left" onclick={() => toggle(`${modKey}-globals`)}>
                <span class="text-gray-500 w-3">{expanded[`${modKey}-globals`] ? '▼' : '▶'}</span>
                <span>Globals ({mod.info.globals.length})</span>
              </button>
              {#if expanded[`${modKey}-globals`]}
                <div class="ml-4">
                  {#each mod.info.globals as glob}
                    <button
                      class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs {isActive && selected === `glob-${glob.index}` ? 'bg-blue-600/30' : ''}"
                      onclick={() => { onSelectModule(modIndex); onSelectGlobal(glob); }}
                    >
                      <span class="w-3" style="color: var(--icon-global);">g</span>
                      <span class="text-gray-300">global {glob.index} ({glob.type})</span>
                    </button>
                  {/each}
                </div>
              {/if}
            {/if}

            {#if mod.info.exports?.length}
              <button class="flex items-center gap-1 w-full px-2 py-1 hover:bg-gray-800 rounded text-left" onclick={() => toggle(`${modKey}-exports`)}>
                <span class="text-gray-500 w-3">{expanded[`${modKey}-exports`] ? '▼' : '▶'}</span>
                <span>Exports ({mod.info.exports.length})</span>
              </button>
              {#if expanded[`${modKey}-exports`]}
                <div class="ml-4">
                  {#each mod.info.exports as exp}
                    <button
                      class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs"
                      onclick={() => { onSelectModule(modIndex); onSelectExport(exp); }}
                    >
                      <span class="w-3" style="color: var(--icon-export);">e</span>
                      <span class="truncate text-gray-300">{exp.name}</span>
                      <span class="text-gray-500">({exp.kind})</span>
                    </button>
                  {/each}
                </div>
              {/if}
            {/if}
          </div>
        {/if}
      </div>
    {/each}
  </div>
</aside>

{#if contextMenu}
  <ContextMenu items={contextMenu.items} x={contextMenu.x} y={contextMenu.y} onClose={() => contextMenu = null} />
{/if}
