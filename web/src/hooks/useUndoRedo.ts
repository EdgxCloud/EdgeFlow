import { useState, useCallback, useRef } from 'react';

export interface HistoryState<T> {
  past: T[];
  present: T;
  future: T[];
}

export interface UndoRedoActions {
  undo: () => void;
  redo: () => void;
  canUndo: boolean;
  canRedo: boolean;
  reset: (newPresent: any) => void;
  push: (newPresent: any) => void;
}

const MAX_HISTORY = 50;

export function useUndoRedo<T>(initialState: T): [T, UndoRedoActions] {
  const [history, setHistory] = useState<HistoryState<T>>({
    past: [],
    present: initialState,
    future: [],
  });

  const undo = useCallback(() => {
    setHistory((current) => {
      if (current.past.length === 0) {
        return current;
      }

      const previous = current.past[current.past.length - 1];
      const newPast = current.past.slice(0, current.past.length - 1);

      return {
        past: newPast,
        present: previous,
        future: [current.present, ...current.future],
      };
    });
  }, []);

  const redo = useCallback(() => {
    setHistory((current) => {
      if (current.future.length === 0) {
        return current;
      }

      const next = current.future[0];
      const newFuture = current.future.slice(1);

      return {
        past: [...current.past, current.present],
        present: next,
        future: newFuture,
      };
    });
  }, []);

  const reset = useCallback((newPresent: T) => {
    setHistory({
      past: [],
      present: newPresent,
      future: [],
    });
  }, []);

  const push = useCallback((newPresent: T) => {
    setHistory((current) => {
      let newPast = [...current.past, current.present];

      if (newPast.length > MAX_HISTORY) {
        newPast = newPast.slice(newPast.length - MAX_HISTORY);
      }

      return {
        past: newPast,
        present: newPresent,
        future: [],
      };
    });
  }, []);

  return [
    history.present,
    {
      undo,
      redo,
      canUndo: history.past.length > 0,
      canRedo: history.future.length > 0,
      reset,
      push,
    },
  ];
}

export function useFlowUndoRedo() {
  const flowStore = useFlowStore();
  const [state, actions] = useUndoRedo(flowStore.getState());

  const undoWithStore = useCallback(() => {
    actions.undo();
    flowStore.setState(state);
  }, [actions, flowStore, state]);

  const redoWithStore = useCallback(() => {
    actions.redo();
    flowStore.setState(state);
  }, [actions, flowStore, state]);

  const pushWithStore = useCallback(() => {
    actions.push(flowStore.getState());
  }, [actions, flowStore]);

  return {
    undo: undoWithStore,
    redo: redoWithStore,
    push: pushWithStore,
    canUndo: actions.canUndo,
    canRedo: actions.canRedo,
  };
}

function useFlowStore() {
  return {
    getState: () => ({}),
    setState: (_state: any) => {},
  };
}
