import { useEffect, useCallback } from 'react';

export type KeyBinding = {
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  meta?: boolean;
  action: () => void;
  description: string;
};

class HotkeyManager {
  private bindings: Map<string, KeyBinding> = new Map();

  register(binding: KeyBinding) {
    const key = this.getKey(binding);
    this.bindings.set(key, binding);
  }

  unregister(binding: KeyBinding) {
    const key = this.getKey(binding);
    this.bindings.delete(key);
  }

  private getKey(binding: KeyBinding): string {
    const parts: string[] = [];
    if (binding.ctrl) parts.push('ctrl');
    if (binding.shift) parts.push('shift');
    if (binding.alt) parts.push('alt');
    if (binding.meta) parts.push('meta');
    parts.push(binding.key.toLowerCase());
    return parts.join('+');
  }

  handleKeyDown(event: KeyboardEvent): boolean {
    const parts: string[] = [];
    if (event.ctrlKey) parts.push('ctrl');
    if (event.shiftKey) parts.push('shift');
    if (event.altKey) parts.push('alt');
    if (event.metaKey) parts.push('meta');
    parts.push(event.key.toLowerCase());

    const key = parts.join('+');
    const binding = this.bindings.get(key);

    if (binding) {
      event.preventDefault();
      event.stopPropagation();
      binding.action();
      return true;
    }

    return false;
  }

  getAllBindings(): KeyBinding[] {
    return Array.from(this.bindings.values());
  }
}

export const hotkeyManager = new HotkeyManager();

export function useHotkey(binding: KeyBinding) {
  useEffect(() => {
    hotkeyManager.register(binding);
    return () => {
      hotkeyManager.unregister(binding);
    };
  }, [binding]);
}

export function useGlobalHotkeys() {
  const handleKeyDown = useCallback((event: KeyboardEvent) => {
    hotkeyManager.handleKeyDown(event);
  }, []);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [handleKeyDown]);
}

export const defaultHotkeys: KeyBinding[] = [
  {
    key: 'z',
    ctrl: true,
    action: () => {},
    description: 'Undo',
  },
  {
    key: 'y',
    ctrl: true,
    action: () => {},
    description: 'Redo',
  },
  {
    key: 'c',
    ctrl: true,
    action: () => {},
    description: 'Copy',
  },
  {
    key: 'v',
    ctrl: true,
    action: () => {},
    description: 'Paste',
  },
  {
    key: 'x',
    ctrl: true,
    action: () => {},
    description: 'Cut',
  },
  {
    key: 'a',
    ctrl: true,
    action: () => {},
    description: 'Select All',
  },
  {
    key: 'Delete',
    action: () => {},
    description: 'Delete Selected',
  },
  {
    key: 'd',
    ctrl: true,
    action: () => {},
    description: 'Duplicate',
  },
  {
    key: 's',
    ctrl: true,
    action: () => {},
    description: 'Save',
  },
  {
    key: 'f',
    ctrl: true,
    action: () => {},
    description: 'Search',
  },
];
