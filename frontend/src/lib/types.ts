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
