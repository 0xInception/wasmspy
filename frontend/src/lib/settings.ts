export type WindowSize = '4k' | '2k' | '1080p' | '720p' | 'custom';

export interface WindowDimensions {
  width: number;
  height: number;
}

export const windowSizePresets: Record<Exclude<WindowSize, 'custom'>, WindowDimensions> = {
  '4k': { width: 3840, height: 2160 },
  '2k': { width: 2560, height: 1440 },
  '1080p': { width: 1920, height: 1080 },
  '720p': { width: 1280, height: 720 },
};

export interface Settings {
  virtualizationThreshold: number;
  defaultShowDecompile: boolean;
  defaultShowDisassembly: boolean;
  fontSize: number;
  windowSize: WindowSize;
  syncSelection: boolean;
  syncScroll: boolean;
}

export const defaultSettings: Settings = {
  virtualizationThreshold: 3000,
  defaultShowDecompile: true,
  defaultShowDisassembly: true,
  fontSize: 13,
  windowSize: '1080p',
  syncSelection: true,
  syncScroll: true,
};

export function loadSettings(): Settings {
  try {
    const saved = localStorage.getItem('wasmspy-settings');
    if (saved) {
      return { ...defaultSettings, ...JSON.parse(saved) };
    }
  } catch (e) {
    console.error('Failed to load settings:', e);
  }
  return { ...defaultSettings };
}

export function saveSettings(settings: Settings) {
  localStorage.setItem('wasmspy-settings', JSON.stringify(settings));
}
