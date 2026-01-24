import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface AppSettings {
  theme: 'light' | 'dark' | 'system';
  language: 'en' | 'fa' | 'ar';
  autoSave: boolean;
  autoSaveInterval: number;
  gridSnap: boolean;
  gridSize: number;
  showMinimap: boolean;
  editorZoom: number;
  notifications: boolean;
  soundEffects: boolean;
}

export interface ServerSettings {
  apiUrl: string;
  wsUrl: string;
  timeout: number;
  retryAttempts: number;
}

interface SettingsState {
  app: AppSettings;
  server: ServerSettings;

  updateAppSettings: (settings: Partial<AppSettings>) => void;
  updateServerSettings: (settings: Partial<ServerSettings>) => void;
  resetToDefaults: () => void;
}

const defaultAppSettings: AppSettings = {
  theme: 'system',
  language: 'en',
  autoSave: true,
  autoSaveInterval: 30,
  gridSnap: true,
  gridSize: 15,
  showMinimap: true,
  editorZoom: 1,
  notifications: true,
  soundEffects: false,
};

const defaultServerSettings: ServerSettings = {
  apiUrl: 'http://localhost:8080',
  wsUrl: 'ws://localhost:8080/ws',
  timeout: 30000,
  retryAttempts: 3,
};

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      app: defaultAppSettings,
      server: defaultServerSettings,

      updateAppSettings: (settings) => {
        set((state) => ({
          app: { ...state.app, ...settings },
        }));
      },

      updateServerSettings: (settings) => {
        set((state) => ({
          server: { ...state.server, ...settings },
        }));
      },

      resetToDefaults: () => {
        set({
          app: defaultAppSettings,
          server: defaultServerSettings,
        });
      },
    }),
    {
      name: 'edgeflow-settings',
    }
  )
);
