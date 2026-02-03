export interface FunctionInfo {
  index: number;
  name: string;
  imported: boolean;
  params: number;
  results: number;
}

export interface MemoryInfo {
  index: number;
  min: number;
  max: number;
  hasMax: boolean;
}

export interface TableInfo {
  index: number;
  min: number;
  max: number;
  hasMax: boolean;
}

export interface GlobalInfo {
  index: number;
  type: string;
  mutable: boolean;
}

export interface ExportInfo {
  name: string;
  kind: string;
  index: number;
}

export interface ModuleInfo {
  functions: FunctionInfo[] | null;
  exports: ExportInfo[] | null;
  memories: MemoryInfo[] | null;
  tables: TableInfo[] | null;
  globals: GlobalInfo[] | null;
}

export interface Bookmark {
  id: string;
  modulePath: string;
  funcIndex: number;
  funcName: string;
}

export interface FunctionRef {
  index: number;
  name: string;
}

export interface XRefInfo {
  callers: FunctionRef[];
  callees: FunctionRef[];
}

export interface ErrorInfo {
  offset: number;
  opcode: string;
  message: string;
}

export interface FunctionErrorInfo {
  funcIndex: number;
  funcName: string;
  errors: ErrorInfo[];
}

export interface ModuleErrorsInfo {
  functions: FunctionErrorInfo[] | null;
  totalErrors: number;
  uniqueErrors: Record<string, number>;
}

export interface FunctionAnnotation {
  name?: string;
  comment?: string;
}

export interface Annotations {
  version: number;
  functions?: Record<string, FunctionAnnotation>;
  comments?: Record<string, string>;
  decompileComments?: Record<string, string>;
  bookmarks?: number[];
}

export type GroupedFunctions = [string, FunctionInfo[]][];

export interface LoadedModule {
  path: string;
  name: string;
  info: ModuleInfo;
  functionsById: Map<number, FunctionInfo>;
  functionsByName: Map<string, FunctionInfo>;
  groupedFunctions: GroupedFunctions;
  groupedImports: GroupedFunctions;
  annotations: Annotations;
}

export interface LineMapping {
  line: number;
  offsets: number[];
}

export interface DecompileMappingsIndexed {
  byLine: Map<number, number[]>;
  byOffset: Map<number, number[]>;
}

export interface DisasmMappings {
  offsetToLine: Map<number, number>;
  lineToOffset: Map<number, number>;
}

export interface CachedFunction {
  decompileCode: string;
  decompileMappings: DecompileMappingsIndexed | null;
  disasmContent: string;
  disasmMappings: DisasmMappings;
}

export interface OpenTab {
  id: string;
  title: string;
  icon: string;
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
