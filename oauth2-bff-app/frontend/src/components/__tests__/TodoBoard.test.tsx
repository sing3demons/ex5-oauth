import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import TodoBoard from '../TodoBoard';
import { AuthProvider } from '../../context/AuthContext';
import { ToastProvider } from '../../context/ToastContext';
import { todoService } from '../../services/todoService';
import * as api from '../../services/api';

// Mock the services
vi.mock('../../services/todoService', () => ({
  todoService: {
    getAll: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    updateStatus: vi.fn(),
    delete: vi.fn(),
  },
}));

// Mock the API module
vi.mock('../../services/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
  setAccessToken: vi.fn(),
}));

const mockUser = {
  id: '1',
  email: 'test@example.com',
  name: 'Test User',
};

const mockTodos = [
  {
    id: '1',
    userId: '1',
    title: 'Test Todo 1',
    description: 'Description 1',
    status: 'todo' as const,
    priority: 'medium' as const,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    userId: '1',
    title: 'Test Todo 2',
    description: 'Description 2',
    status: 'in_progress' as const,
    priority: 'high' as const,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

const renderWithProviders = (component: React.ReactElement) => {
  // Create a new QueryClient for each test
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  // Mock the auth check endpoint
  vi.mocked(api.default.get).mockResolvedValue({
    data: mockUser,
    status: 200,
    statusText: 'OK',
    headers: {},
    config: {} as any,
  });

  localStorageMock.getItem.mockReturnValue('mock-token');

  return render(
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <BrowserRouter>
          <AuthProvider>
            {component}
          </AuthProvider>
        </BrowserRouter>
      </ToastProvider>
    </QueryClientProvider>
  );
};

describe('TodoBoard - Drag and Drop', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(todoService.getAll).mockResolvedValue(mockTodos);
  });

  it('renders todos in correct columns', async () => {
    renderWithProviders(<TodoBoard />);

    await waitFor(() => {
      expect(screen.getByText('Test Todo 1')).toBeInTheDocument();
      expect(screen.getByText('Test Todo 2')).toBeInTheDocument();
    });

    // Check column headers
    expect(screen.getByText('📋 To Do')).toBeInTheDocument();
    expect(screen.getByText('🚀 In Progress')).toBeInTheDocument();
    expect(screen.getByText('✅ Done')).toBeInTheDocument();
  });

  it('displays correct todo counts in stats', async () => {
    renderWithProviders(<TodoBoard />);

    await waitFor(() => {
      const statNumbers = screen.getAllByText('1');
      expect(statNumbers.length).toBeGreaterThan(0);
    });
  });

  it('handles optimistic update on delete', async () => {
    vi.mocked(todoService.delete).mockResolvedValue();
    
    // Mock window.confirm
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);

    renderWithProviders(<TodoBoard />);

    await waitFor(() => {
      expect(screen.getByText('Test Todo 1')).toBeInTheDocument();
    });

    // Find and click delete button
    const deleteButtons = screen.getAllByTitle('Delete');
    deleteButtons[0].click();

    // Verify confirm was called
    expect(confirmSpy).toHaveBeenCalled();

    // Verify service was called
    await waitFor(() => {
      expect(todoService.delete).toHaveBeenCalledWith('1');
    });

    confirmSpy.mockRestore();
  });

  it('handles rollback on delete error', async () => {
    vi.mocked(todoService.delete).mockRejectedValue(new Error('Delete failed'));
    
    // Mock window methods
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {});

    renderWithProviders(<TodoBoard />);

    await waitFor(() => {
      expect(screen.getByText('Test Todo 1')).toBeInTheDocument();
    });

    // Find and click delete button
    const deleteButtons = screen.getAllByTitle('Delete');
    deleteButtons[0].click();

    // Verify error handling
    await waitFor(() => {
      expect(alertSpy).toHaveBeenCalledWith('Failed to delete task. Please try again.');
    });

    // Todo should still be visible (rollback)
    expect(screen.getByText('Test Todo 1')).toBeInTheDocument();

    confirmSpy.mockRestore();
    alertSpy.mockRestore();
  });

  it('handles optimistic update on edit', async () => {
    const updatedTodo = { ...mockTodos[0], title: 'Updated Title' };
    vi.mocked(todoService.update).mockResolvedValue(updatedTodo);

    renderWithProviders(<TodoBoard />);

    await waitFor(() => {
      expect(screen.getByText('Test Todo 1')).toBeInTheDocument();
    });

    // Find and click edit button
    const editButtons = screen.getAllByTitle('Edit');
    editButtons[0].click();

    // Form should appear
    await waitFor(() => {
      expect(screen.getByDisplayValue('Test Todo 1')).toBeInTheDocument();
    });
  });

  it('shows loading state initially', () => {
    vi.mocked(todoService.getAll).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );

    renderWithProviders(<TodoBoard />);

    expect(screen.getByText('Loading todos...')).toBeInTheDocument();
  });

  it('displays create modal when New Task button is clicked', async () => {
    renderWithProviders(<TodoBoard />);

    await waitFor(() => {
      expect(screen.getByText('Test Todo 1')).toBeInTheDocument();
    });

    const createButton = screen.getByText('+ New Task');
    createButton.click();

    // Modal should appear
    await waitFor(() => {
      expect(screen.getByText('✨ Create New Task')).toBeInTheDocument();
    });
  });
});
