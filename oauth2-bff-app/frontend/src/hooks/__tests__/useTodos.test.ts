import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createElement, ReactNode } from 'react';
import {
  useTodos,
  useCreateTodo,
  useUpdateTodo,
  useUpdateTodoStatus,
  useDeleteTodo,
} from '../useTodos';
import { todoService } from '../../services/todoService';
import { Todo } from '../../types/todo';

// Mock the todoService
vi.mock('../../services/todoService', () => ({
  todoService: {
    getAll: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    updateStatus: vi.fn(),
    delete: vi.fn(),
  },
}));

const mockTodos: Todo[] = [
  {
    id: '1',
    userId: 'user1',
    title: 'Test Todo 1',
    description: 'Description 1',
    status: 'todo',
    priority: 'medium',
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
  },
  {
    id: '2',
    userId: 'user1',
    title: 'Test Todo 2',
    status: 'in_progress',
    priority: 'high',
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
  },
];

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useTodos', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch todos successfully', async () => {
    vi.mocked(todoService.getAll).mockResolvedValue(mockTodos);

    const { result } = renderHook(() => useTodos(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(mockTodos);
    expect(todoService.getAll).toHaveBeenCalledTimes(1);
  });

  it('should handle fetch error', async () => {
    const error = new Error('Failed to fetch');
    vi.mocked(todoService.getAll).mockRejectedValue(error);

    const { result } = renderHook(() => useTodos(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error).toBe(error);
  });
});

describe('useCreateTodo', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should create todo with optimistic update', async () => {
    const newTodo: Todo = {
      id: '3',
      userId: 'user1',
      title: 'New Todo',
      description: 'New Description',
      status: 'todo',
      priority: 'low',
      createdAt: '2025-01-02T00:00:00Z',
      updatedAt: '2025-01-02T00:00:00Z',
    };

    vi.mocked(todoService.getAll).mockResolvedValue(mockTodos);
    vi.mocked(todoService.create).mockResolvedValue(newTodo);

    const wrapper = createWrapper();
    const { result: todosResult } = renderHook(() => useTodos(), { wrapper });
    const { result: createResult } = renderHook(() => useCreateTodo(), { wrapper });

    await waitFor(() => expect(todosResult.current.isSuccess).toBe(true));

    createResult.current.mutate({
      title: 'New Todo',
      description: 'New Description',
      priority: 'low',
    });

    await waitFor(() => expect(createResult.current.isSuccess).toBe(true));

    expect(todoService.create).toHaveBeenCalledWith({
      title: 'New Todo',
      description: 'New Description',
      priority: 'low',
    });
  });
});

describe('useUpdateTodo', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should update todo with optimistic update', async () => {
    const updatedTodo: Todo = {
      ...mockTodos[0],
      title: 'Updated Title',
    };

    vi.mocked(todoService.getAll).mockResolvedValue(mockTodos);
    vi.mocked(todoService.update).mockResolvedValue(updatedTodo);

    const wrapper = createWrapper();
    const { result: todosResult } = renderHook(() => useTodos(), { wrapper });
    const { result: updateResult } = renderHook(() => useUpdateTodo(), { wrapper });

    await waitFor(() => expect(todosResult.current.isSuccess).toBe(true));

    updateResult.current.mutate({
      id: '1',
      data: { title: 'Updated Title' },
    });

    await waitFor(() => expect(updateResult.current.isSuccess).toBe(true));

    expect(todoService.update).toHaveBeenCalledWith('1', { title: 'Updated Title' });
  });
});

describe('useUpdateTodoStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should update todo status with optimistic update', async () => {
    const updatedTodo: Todo = {
      ...mockTodos[0],
      status: 'done',
    };

    vi.mocked(todoService.getAll).mockResolvedValue(mockTodos);
    vi.mocked(todoService.updateStatus).mockResolvedValue(updatedTodo);

    const wrapper = createWrapper();
    const { result: todosResult } = renderHook(() => useTodos(), { wrapper });
    const { result: updateStatusResult } = renderHook(() => useUpdateTodoStatus(), { wrapper });

    await waitFor(() => expect(todosResult.current.isSuccess).toBe(true));

    updateStatusResult.current.mutate({
      id: '1',
      status: 'done',
    });

    await waitFor(() => expect(updateStatusResult.current.isSuccess).toBe(true));

    expect(todoService.updateStatus).toHaveBeenCalledWith('1', 'done');
  });
});

describe('useDeleteTodo', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should delete todo with optimistic update', async () => {
    vi.mocked(todoService.getAll).mockResolvedValue(mockTodos);
    vi.mocked(todoService.delete).mockResolvedValue();

    const wrapper = createWrapper();
    const { result: todosResult } = renderHook(() => useTodos(), { wrapper });
    const { result: deleteResult } = renderHook(() => useDeleteTodo(), { wrapper });

    await waitFor(() => expect(todosResult.current.isSuccess).toBe(true));

    deleteResult.current.mutate('1');

    await waitFor(() => expect(deleteResult.current.isSuccess).toBe(true));

    expect(todoService.delete).toHaveBeenCalledWith('1');
  });
});
