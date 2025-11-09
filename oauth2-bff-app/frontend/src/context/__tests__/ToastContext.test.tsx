import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ToastProvider, useToast } from '../ToastContext';
import { act } from 'react';

// Test component that uses toast
function TestComponent() {
  const { showSuccess, showError, showInfo, showWarning, toasts } = useToast();

  return (
    <div>
      <button onClick={() => showSuccess('Success message')}>Show Success</button>
      <button onClick={() => showError('Error message')}>Show Error</button>
      <button onClick={() => showInfo('Info message')}>Show Info</button>
      <button onClick={() => showWarning('Warning message')}>Show Warning</button>
      <div data-testid="toast-count">{toasts.length}</div>
      {toasts.map((toast) => (
        <div key={toast.id} data-testid={`toast-${toast.type}`}>
          {toast.message}
        </div>
      ))}
    </div>
  );
}

describe('ToastContext', () => {
  it('provides toast functions', () => {
    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    );

    expect(screen.getByRole('button', { name: /show success/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /show error/i })).toBeInTheDocument();
  });

  it('shows success toast', async () => {
    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    );

    const button = screen.getByRole('button', { name: /show success/i });
    act(() => {
      button.click();
    });

    await waitFor(() => {
      expect(screen.getByTestId('toast-success')).toBeInTheDocument();
      expect(screen.getByText('Success message')).toBeInTheDocument();
    });
  });

  it('shows error toast', async () => {
    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    );

    const button = screen.getByRole('button', { name: /show error/i });
    act(() => {
      button.click();
    });

    await waitFor(() => {
      expect(screen.getByTestId('toast-error')).toBeInTheDocument();
      expect(screen.getByText('Error message')).toBeInTheDocument();
    });
  });

  it('shows multiple toasts', async () => {
    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    );

    act(() => {
      screen.getByRole('button', { name: /show success/i }).click();
      screen.getByRole('button', { name: /show error/i }).click();
    });

    await waitFor(() => {
      expect(screen.getByTestId('toast-count')).toHaveTextContent('2');
    });
  });

  it('auto-removes toast after duration', async () => {
    vi.useFakeTimers();

    const { rerender } = render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    );

    // Show the toast
    act(() => {
      screen.getByRole('button', { name: /show info/i }).click();
    });

    // Verify toast is shown
    expect(screen.getByTestId('toast-info')).toBeInTheDocument();

    // Fast-forward time by 5 seconds (default duration)
    act(() => {
      vi.advanceTimersByTime(5100);
    });

    // Force re-render to reflect state changes
    rerender(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>
    );

    // Toast should be removed
    expect(screen.queryByTestId('toast-info')).not.toBeInTheDocument();

    vi.useRealTimers();
  });
});
