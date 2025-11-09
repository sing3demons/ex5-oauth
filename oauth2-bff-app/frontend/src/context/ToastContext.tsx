import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
  id: string;
  message: string;
  type: ToastType;
  duration?: number;
  action?: {
    label: string;
    onClick: () => void;
  };
}

interface ToastContextType {
  toasts: Toast[];
  showToast: (message: string, type?: ToastType, options?: Partial<Toast>) => void;
  showSuccess: (message: string, options?: Partial<Toast>) => void;
  showError: (message: string, options?: Partial<Toast>) => void;
  showInfo: (message: string, options?: Partial<Toast>) => void;
  showWarning: (message: string, options?: Partial<Toast>) => void;
  removeToast: (id: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export function useToast(): ToastContextType {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within ToastProvider');
  }
  return context;
}

interface ToastProviderProps {
  children: ReactNode;
}

export function ToastProvider({ children }: ToastProviderProps) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  }, []);

  const showToast = useCallback(
    (message: string, type: ToastType = 'info', options?: Partial<Toast>) => {
      const id = `toast-${Date.now()}-${Math.random()}`;
      const duration = options?.duration ?? 5000;

      const toast: Toast = {
        id,
        message,
        type,
        duration,
        action: options?.action,
      };

      setToasts((prev) => [...prev, toast]);

      // Auto-remove toast after duration
      if (duration > 0) {
        setTimeout(() => {
          removeToast(id);
        }, duration);
      }
    },
    [removeToast]
  );

  const showSuccess = useCallback(
    (message: string, options?: Partial<Toast>) => {
      showToast(message, 'success', options);
    },
    [showToast]
  );

  const showError = useCallback(
    (message: string, options?: Partial<Toast>) => {
      showToast(message, 'error', { duration: 7000, ...options });
    },
    [showToast]
  );

  const showInfo = useCallback(
    (message: string, options?: Partial<Toast>) => {
      showToast(message, 'info', options);
    },
    [showToast]
  );

  const showWarning = useCallback(
    (message: string, options?: Partial<Toast>) => {
      showToast(message, 'warning', options);
    },
    [showToast]
  );

  return (
    <ToastContext.Provider
      value={{
        toasts,
        showToast,
        showSuccess,
        showError,
        showInfo,
        showWarning,
        removeToast,
      }}
    >
      {children}
    </ToastContext.Provider>
  );
}
