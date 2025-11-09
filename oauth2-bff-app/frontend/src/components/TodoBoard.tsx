import { useState, useMemo } from 'react';
import {
  DndContext,
  DragEndEvent,
  DragOverlay,
  DragStartEvent,
  PointerSensor,
  TouchSensor,
  useSensor,
  useSensors
} from '@dnd-kit/core';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import {
  useTodos,
  useCreateTodo,
  useUpdateTodo,
  useUpdateTodoStatus,
  useDeleteTodo,
} from '../hooks/useTodos';
import { TodoStatus } from '../types/todo';
import TodoColumn from './TodoColumn';
import TodoCard from './TodoCard';
import CreateTodoModal from './CreateTodoModal';
import Header from './Header';
import LoadingSkeleton from './LoadingSkeleton';

export default function TodoBoard() {
  const { logout, user } = useAuth();
  const { showSuccess, showError } = useToast();
  const [activeId, setActiveId] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);

  // React Query hooks
  const { data: todos = [], isLoading, isError, error } = useTodos();
  const createTodo = useCreateTodo();
  const updateTodo = useUpdateTodo();
  const updateTodoStatus = useUpdateTodoStatus();
  const deleteTodo = useDeleteTodo();

  const activeTodo = activeId ? todos.find(t => t.id === activeId) : null;

  // Configure sensors for both desktop and mobile
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8, // 8px movement required before drag starts
      },
    }),
    useSensor(TouchSensor, {
      activationConstraint: {
        delay: 200, // 200ms delay for touch to distinguish from scrolling
        tolerance: 5,
      },
    })
  );

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveId(null);

    if (!over) return;

    const todoId = active.id as string;
    const newStatus = over.id as TodoStatus;

    const todo = todos.find(t => t.id === todoId);
    if (!todo || todo.status === newStatus) return;

    // Use React Query mutation with optimistic update
    updateTodoStatus.mutate(
      { id: todoId, status: newStatus },
      {
        onSuccess: () => {
          showSuccess('Task moved successfully');
        },
        onError: (error) => {
          const message = error instanceof Error ? error.message : 'Failed to move task';
          showError(message, {
            action: {
              label: 'Retry',
              onClick: () => updateTodoStatus.mutate({ id: todoId, status: newStatus }),
            },
          });
        },
      }
    );
  };

  const handleCreate = async (data: { title: string; description?: string; priority?: any }) => {
    createTodo.mutate(data, {
      onSuccess: () => {
        setShowCreateModal(false);
        showSuccess('Task created successfully');
      },
      onError: (error) => {
        const message = error instanceof Error ? error.message : 'Failed to create task';
        showError(message, {
          action: {
            label: 'Retry',
            onClick: () => createTodo.mutate(data),
          },
        });
      },
    });
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this todo?')) return;

    deleteTodo.mutate(id, {
      onSuccess: () => {
        showSuccess('Task deleted successfully');
      },
      onError: (error) => {
        const message = error instanceof Error ? error.message : 'Failed to delete task';
        showError(message, {
          action: {
            label: 'Retry',
            onClick: () => deleteTodo.mutate(id),
          },
        });
      },
    });
  };

  const handleUpdate = async (id: string, data: any) => {
    updateTodo.mutate(
      { id, data },
      {
        onSuccess: () => {
          showSuccess('Task updated successfully');
        },
        onError: (error) => {
          const message = error instanceof Error ? error.message : 'Failed to update task';
          showError(message, {
            action: {
              label: 'Retry',
              onClick: () => updateTodo.mutate({ id, data }),
            },
          });
        },
      }
    );
  };

  // Group todos by status using useMemo for performance
  const todosByStatus = useMemo(() => ({
    todo: todos.filter(t => t.status === 'todo'),
    in_progress: todos.filter(t => t.status === 'in_progress'),
    done: todos.filter(t => t.status === 'done')
  }), [todos]);

  // Loading state - use skeleton for better UX
  if (isLoading) {
    return <LoadingSkeleton />;
  }

  // Error state
  if (isError) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-indigo-500 to-purple-600 text-white text-lg p-4">
        <div className="text-6xl mb-5">⚠️</div>
        <p className="text-2xl font-bold mb-2">Failed to load todos</p>
        <p className="text-base opacity-90 mb-8 text-center max-w-md">
          {error instanceof Error ? error.message : 'An unexpected error occurred'}
        </p>
        <button
          onClick={() => window.location.reload()}
          className="px-6 py-3 text-base font-bold text-indigo-600 bg-white rounded-lg hover:shadow-lg transition-all touch-manipulation"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-500 to-purple-600 p-4 md:p-6 lg:p-8">
      {/* Header */}
      {user && <Header user={user} onLogout={logout} />}

      {/* Page Title */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-6 md:mb-8 gap-4">
        <div>
          <h1 className="text-3xl md:text-4xl lg:text-5xl font-bold text-white m-0">
            📝 My Tasks
          </h1>
          <p className="text-sm md:text-base text-white/90 mt-1">
            Welcome back, {user?.name || user?.email}!
          </p>
        </div>
        <button 
          onClick={() => setShowCreateModal(true)} 
          className="px-4 py-3 md:px-6 md:py-3 text-sm md:text-base font-bold text-indigo-600 bg-white rounded-lg hover:shadow-lg transition-all hover:-translate-y-0.5 active:translate-y-0 touch-manipulation min-h-[44px]"
        >
          + New Task
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 md:gap-6 mb-6 md:mb-8">
        <div className="bg-white/95 rounded-xl p-4 md:p-6 text-center shadow-md">
          <div className="text-3xl md:text-4xl lg:text-5xl font-bold text-indigo-600 mb-1">
            {todosByStatus.todo.length}
          </div>
          <div className="text-xs md:text-sm text-gray-600 uppercase tracking-wider">
            To Do
          </div>
        </div>
        <div className="bg-white/95 rounded-xl p-4 md:p-6 text-center shadow-md">
          <div className="text-3xl md:text-4xl lg:text-5xl font-bold text-indigo-600 mb-1">
            {todosByStatus.in_progress.length}
          </div>
          <div className="text-xs md:text-sm text-gray-600 uppercase tracking-wider">
            In Progress
          </div>
        </div>
        <div className="bg-white/95 rounded-xl p-4 md:p-6 text-center shadow-md">
          <div className="text-3xl md:text-4xl lg:text-5xl font-bold text-indigo-600 mb-1">
            {todosByStatus.done.length}
          </div>
          <div className="text-xs md:text-sm text-gray-600 uppercase tracking-wider">
            Done
          </div>
        </div>
      </div>

      {/* Board - Responsive Grid Layout */}
      <DndContext
        sensors={sensors}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
      >
        {/* Mobile: Stacked (1 column), Tablet: 2 columns, Desktop: 3 columns */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6 items-start">
          <TodoColumn
            id="todo"
            title="📋 To Do"
            todos={todosByStatus.todo}
            onDelete={handleDelete}
            onUpdate={handleUpdate}
            onAddClick={() => setShowCreateModal(true)}
          />
          <TodoColumn
            id="in_progress"
            title="🚀 In Progress"
            todos={todosByStatus.in_progress}
            onDelete={handleDelete}
            onUpdate={handleUpdate}
          />
          <TodoColumn
            id="done"
            title="✅ Done"
            todos={todosByStatus.done}
            onDelete={handleDelete}
            onUpdate={handleUpdate}
          />
        </div>

        <DragOverlay>
          {activeTodo ? <TodoCard todo={activeTodo} isDragging /> : null}
        </DragOverlay>
      </DndContext>

      {/* Create Modal */}
      {showCreateModal && (
        <CreateTodoModal
          onClose={() => setShowCreateModal(false)}
          onCreate={handleCreate}
          isCreating={createTodo.isPending}
        />
      )}
    </div>
  );
}


