<script lang="ts">
  import type { ModuleInfo, FunctionInfo, MemoryInfo, TableInfo, GlobalInfo, ExportInfo } from './types';
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
    onSelectModule,
    onCloseModule,
    onSelectFunction,
    onSelectMemory,
    onSelectTable,
    onSelectGlobal,
    onSelectExport,
    onOpenFile,
  }: {
    modules: LoadedModule[];
    activeModuleIndex: number;
    selected: string;
    loading: boolean;
    loadingName: string;
    onSelectModule: (index: number) => void;
    onCloseModule: (index: number) => void;
    onSelectFunction: (fn: FunctionInfo) => void;
    onSelectMemory: (mem: MemoryInfo) => void;
    onSelectTable: (tbl: TableInfo) => void;
    onSelectGlobal: (glob: GlobalInfo) => void;
    onSelectExport: (exp: ExportInfo) => void;
    onOpenFile: () => void;
  } = $props();

  let expanded: Record<string, boolean> = $state({});
  let expandedGroups: Record<string, boolean> = $state({});
  let contextMenu: { x: number; y: number; items: MenuItem[] } | null = $state(null);

  function showFnContextMenu(e: MouseEvent, fn: FunctionInfo) {
    e.preventDefault();
    contextMenu = {
      x: e.clientX,
      y: e.clientY,
      items: [
        { label: 'Copy name', action: () => navigator.clipboard.writeText(fn.name) },
        { label: 'Copy index', action: () => navigator.clipboard.writeText(String(fn.index)) },
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
</script>

<aside class="w-72 bg-gray-900 border-r border-gray-700 overflow-auto text-sm flex flex-col">
  <div class="p-2 border-b border-gray-700 text-gray-400 text-xs uppercase tracking-wide flex items-center justify-between">
    <span>Explorer</span>
    <button class="text-gray-500 hover:text-gray-300 px-1" onclick={onOpenFile} title="Open file">+</button>
  </div>
  <div class="p-1 flex-1">
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
            {#if expanded[`${modKey}-functions`]}
              <div class="ml-4">
                {#each definedFunctions as [group, fns]}
                  {@const groupKey = `${modKey}-grp-${group}`}
                  <button
                    class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs"
                    onclick={() => toggleGroup(groupKey)}
                  >
                    <span class="text-gray-500 w-3">{expandedGroups[groupKey] ? '▼' : '▶'}</span>
                    <span class="text-yellow-400 truncate">{group}</span>
                    <span class="text-gray-500">({fns.length})</span>
                  </button>
                  {#if expandedGroups[groupKey]}
                    <div class="ml-4">
                      {#each fns as fn}
                        <button
                          class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs {isActive && selected === `func-${fn.index}` ? 'bg-blue-600/30' : ''}"
                          onclick={() => { onSelectModule(modIndex); onSelectFunction(fn); }}
                          oncontextmenu={(e) => showFnContextMenu(e, fn)}
                        >
                          <span class="w-3 text-green-400">f</span>
                          <span class="truncate text-gray-300">{fn.name.slice(group.length + 1) || fn.name}</span>
                        </button>
                      {/each}
                    </div>
                  {/if}
                {/each}
              </div>
            {/if}

            <button class="flex items-center gap-1 w-full px-2 py-1 hover:bg-gray-800 rounded text-left" onclick={() => toggle(`${modKey}-imports`)}>
              <span class="text-gray-500 w-3">{expanded[`${modKey}-imports`] ? '▼' : '▶'}</span>
              <span>Imports ({countFunctions(mod.info, true)})</span>
            </button>
            {#if expanded[`${modKey}-imports`]}
              <div class="ml-4">
                {#each importedFunctions as [group, fns]}
                  {@const groupKey = `${modKey}-imp-${group}`}
                  <button
                    class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs"
                    onclick={() => toggleGroup(groupKey)}
                  >
                    <span class="text-gray-500 w-3">{expandedGroups[groupKey] ? '▼' : '▶'}</span>
                    <span class="text-yellow-400 truncate">{group}</span>
                    <span class="text-gray-500">({fns.length})</span>
                  </button>
                  {#if expandedGroups[groupKey]}
                    <div class="ml-4">
                      {#each fns as fn}
                        <button
                          class="flex items-center gap-1 w-full px-2 py-0.5 hover:bg-gray-800 rounded text-left text-xs {isActive && selected === `func-${fn.index}` ? 'bg-blue-600/30' : ''}"
                          onclick={() => { onSelectModule(modIndex); onSelectFunction(fn); }}
                          oncontextmenu={(e) => showFnContextMenu(e, fn)}
                        >
                          <span class="w-3 text-gray-500">f</span>
                          <span class="truncate text-gray-300">{fn.name.slice(group.length + 1) || fn.name}</span>
                        </button>
                      {/each}
                    </div>
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
                      <span class="w-3 text-cyan-400">m</span>
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
                      <span class="w-3 text-orange-400">t</span>
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
                      <span class="w-3 text-purple-400">g</span>
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
                      <span class="w-3 text-cyan-400">e</span>
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
