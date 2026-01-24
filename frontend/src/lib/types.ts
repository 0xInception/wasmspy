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
