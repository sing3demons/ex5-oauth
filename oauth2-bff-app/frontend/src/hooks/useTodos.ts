import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { todoService } from '../services/todoService';
import { Todo, TodoStatus, TodoPriority } from '../types/todo';

// Query keys
export const todoKeys = {
  all: ['todos'] as const,
  lists: () => [...todoKeys.all, 'list'] as const,
  list: (filters: string) => [...todoKeys.lists(), { filters }] as const,
  details: () => [...todoKeys.all, 'detail'] as const,
  detail: (id: string) => [...todoKeys.details(), id] as const,
};

// Hook to fetch all todos
export function useTodos() {
  return useQuery({
    queryKey: todoKeys.all,
    queryFn: todoService.getAll,
  });
}

// Hook to create a new todo
export function useCreateTodo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      title: string;
      description?: string;
      priority?: TodoPriority;
    }) => todoService.create(data),
    onMutate: async (newTodo) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: todoKeys.all });

      // Snapshot the previous value
      const previousTodos = queryClient.getQueryData<Todo[]>(todoKeys.all);

      // Optimistically update to the new value
      queryClient.setQueryData<Todo[]>(todoKeys.all, (old = []) => [
        {
          id: `temp-${Date.now()}`,
          userId: 'temp',
          title: newTodo.title,
          description: newTodo.description,
          status: 'todo' as TodoStatus,
          priority: newTodo.priority || 'medium',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        },
        ...old,
      ]);

      // Return context with the previous value
      return { previousTodos };
    },
    onError: (_err, _newTodo, context) => {
      // Rollback to the previous value on error
      if (context?.previousTodos) {
        queryClient.setQueryData(todoKeys.all, context.previousTodos);
      }
    },
    onSuccess: () => {
      // Refetch to get the actual data from server
      queryClient.invalidateQueries({ queryKey: todoKeys.all });
    },
  });
}

// Hook to update a todo
export function useUpdateTodo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: {
        title?: string;
        description?: string;
        status?: TodoStatus;
        priority?: TodoPriority;
      };
    }) => todoService.update(id, data),
    onMutate: async ({ id, data }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: todoKeys.all });

      // Snapshot the previous value
      const previousTodos = queryClient.getQueryData<Todo[]>(todoKeys.all);

      // Optimistically update to the new value
      queryClient.setQueryData<Todo[]>(todoKeys.all, (old = []) =>
        old.map((todo) =>
          todo.id === id
            ? { ...todo, ...data, updatedAt: new Date().toISOString() }
            : todo
        )
      );

      // Return context with the previous value
      return { previousTodos };
    },
    onError: (_err, _variables, context) => {
      // Rollback to the previous value on error
      if (context?.previousTodos) {
        queryClient.setQueryData(todoKeys.all, context.previousTodos);
      }
    },
    onSuccess: () => {
      // Refetch to get the actual data from server
      queryClient.invalidateQueries({ queryKey: todoKeys.all });
    },
  });
}

// Hook to update todo status (for drag and drop)
export function useUpdateTodoStatus() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: TodoStatus }) =>
      todoService.updateStatus(id, status),
    onMutate: async ({ id, status }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: todoKeys.all });

      // Snapshot the previous value
      const previousTodos = queryClient.getQueryData<Todo[]>(todoKeys.all);

      // Optimistically update to the new value
      queryClient.setQueryData<Todo[]>(todoKeys.all, (old = []) =>
        old.map((todo) =>
          todo.id === id
            ? { ...todo, status, updatedAt: new Date().toISOString() }
            : todo
        )
      );

      // Return context with the previous value
      return { previousTodos };
    },
    onError: (_err, _variables, context) => {
      // Rollback to the previous value on error
      if (context?.previousTodos) {
        queryClient.setQueryData(todoKeys.all, context.previousTodos);
      }
    },
    onSuccess: () => {
      // Refetch to get the actual data from server
      queryClient.invalidateQueries({ queryKey: todoKeys.all });
    },
  });
}

// Hook to delete a todo
export function useDeleteTodo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => todoService.delete(id),
    onMutate: async (id) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: todoKeys.all });

      // Snapshot the previous value
      const previousTodos = queryClient.getQueryData<Todo[]>(todoKeys.all);

      // Optimistically update to the new value
      queryClient.setQueryData<Todo[]>(todoKeys.all, (old = []) =>
        old.filter((todo) => todo.id !== id)
      );

      // Return context with the previous value
      return { previousTodos };
    },
    onError: (_err, _id, context) => {
      // Rollback to the previous value on error
      if (context?.previousTodos) {
        queryClient.setQueryData(todoKeys.all, context.previousTodos);
      }
    },
    onSuccess: () => {
      // Refetch to get the actual data from server
      queryClient.invalidateQueries({ queryKey: todoKeys.all });
    },
  });
}
