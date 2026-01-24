import { useEffect, useRef, useCallback } from 'react';
import { debounce } from 'lodash';

export interface AutoSaveOptions {
  interval?: number;
  onSave: () => Promise<void>;
  enabled?: boolean;
  debounceTime?: number;
}

export function useAutoSave({
  interval = 30000,
  onSave,
  enabled = true,
  debounceTime = 2000,
}: AutoSaveOptions) {
  const saveTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const lastSaveRef = useRef<number>(Date.now());
  const isSavingRef = useRef(false);

  const debouncedSave = useCallback(
    debounce(async () => {
      if (isSavingRef.current || !enabled) {
        return;
      }

      try {
        isSavingRef.current = true;
        await onSave();
        lastSaveRef.current = Date.now();
      } catch (error) {
        console.error('Auto-save failed:', error);
      } finally {
        isSavingRef.current = false;
      }
    }, debounceTime),
    [onSave, enabled, debounceTime]
  );

  const saveNow = useCallback(async () => {
    if (isSavingRef.current) {
      return;
    }

    try {
      isSavingRef.current = true;
      await onSave();
      lastSaveRef.current = Date.now();
    } catch (error) {
      console.error('Manual save failed:', error);
      throw error;
    } finally {
      isSavingRef.current = false;
    }
  }, [onSave]);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    saveTimeoutRef.current = setInterval(() => {
      const timeSinceLastSave = Date.now() - lastSaveRef.current;
      if (timeSinceLastSave >= interval && !isSavingRef.current) {
        debouncedSave();
      }
    }, 5000);

    return () => {
      if (saveTimeoutRef.current) {
        clearInterval(saveTimeoutRef.current);
      }
      debouncedSave.cancel();
    };
  }, [enabled, interval, debouncedSave]);

  useEffect(() => {
    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      if (isSavingRef.current) {
        event.preventDefault();
        event.returnValue = '';
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, []);

  return {
    saveNow,
    isSaving: isSavingRef.current,
    lastSave: lastSaveRef.current,
    triggerSave: debouncedSave,
  };
}
