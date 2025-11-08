import { useState, useEffect } from 'react';
import { DndContext, DragEndEvent, DragOverlay, DragStartEvent } from '@dnd-kit/core';
import { useAuth } from '../context/AuthContext';
import { todoService } from '../services/todoService';
import { Todo, TodoStatus } from '../types/todo';
import TodoColumn from './TodoColumn';
import TodoCard from './TodoCard';
import CreateTodoModal from './CreateTodoModal';

export default function TodoBoard() {
  const { accessToken, logout, user } = useAuth();
  const [todos, setTodos] = useState<Todo[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);

  const activeTodo = activeId ? todos.find(t => t.id === activeId) : null;

  useEffect(() => {
    loadTodos();
  }, []);

  const loadTodos = async () => {
    if (!accessToken) return;
    
    try {
      const data = await todoService.getAll(accessToken);
      setTodos(data);
    } catch (error) {
      console.error('Failed to load todos:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveId(null);

    if (!over || !accessToken) return;

    const todoId = active.id as string;
    const newStatus = over.id as TodoStatus;

    const todo = todos.find(t => t.id === todoId);
    if (!todo || todo.status === newStatus) return;

    // Optimistic update
    setTodos(prev =>
      prev.map(t => (t.id === todoId ? { ...t, status: newStatus } : t))
    );

    try {
      await todoService.updateStatus(accessToken, todoId, newStatus);
    } catch (error) {
      console.error('Failed to update todo:', error);
      // Revert on error
      loadTodos();
    }
  };

  const handleCreate = async (data: { title: string; description?: string; priority?: any }) => {
    if (!accessToken) return;

    try {
      const newTodo = await todoService.create(accessToken, data);
      setTodos(prev => [newTodo, ...prev]);
      setShowCreateModal(false);
    } catch (error) {
      console.error('Failed to create todo:', error);
    }
  };

  const handleDelete = async (id: string) => {
    if (!accessToken || !confirm('Are you sure you want to delete this todo?')) return;

    try {
      await todoService.delete(accessToken, id);
      setTodos(prev => prev.filter(t => t.id !== id));
    } catch (error) {
      console.error('Failed to delete todo:', error);
    }
  };

  const handleUpdate = async (id: string, data: any) => {
    if (!accessToken) return;

    try {
      const updated = await todoService.update(accessToken, id, data);
      setTodos(prev => prev.map(t => (t.id === id ? updated : t)));
    } catch (error) {
      console.error('Failed to update todo:', error);
    }
  };

  const todosByStatus = {
    todo: todos.filter(t => t.status === 'todo'),
    in_progress: todos.filter(t => t.status === 'in_progress'),
    done: todos.filter(t => t.status === 'done')
  };

  if (loading) {
    return (
      <div style={styles.loading}>
        <div style={styles.spinner}></div>
        <p>Loading todos...</p>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      {/* Header */}
      <div style={styles.header}>
        <div>
          <h1 style={styles.title}>📝 My Tasks</h1>
          <p style={styles.subtitle}>
            Welcome back, {user?.name || user?.email}!
          </p>
        </div>
        <div style={styles.headerActions}>
          <button onClick={() => setShowCreateModal(true)} style={styles.createButton}>
            + New Task
          </button>
          <button onClick={logout} style={styles.logoutButton}>
            Logout
          </button>
        </div>
      </div>

      {/* Stats */}
      <div style={styles.stats}>
        <div style={styles.statCard}>
          <div style={styles.statNumber}>{todosByStatus.todo.length}</div>
          <div style={styles.statLabel}>To Do</div>
        </div>
        <div style={styles.statCard}>
          <div style={styles.statNumber}>{todosByStatus.in_progress.length}</div>
          <div style={styles.statLabel}>In Progress</div>
        </div>
        <div style={styles.statCard}>
          <div style={styles.statNumber}>{todosByStatus.done.length}</div>
          <div style={styles.statLabel}>Done</div>
        </div>
      </div>

      {/* Board */}
      <DndContext onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
        <div style={styles.board}>
          <TodoColumn
            id="todo"
            title="📋 To Do"
            todos={todosByStatus.todo}
            onDelete={handleDelete}
            onUpdate={handleUpdate}
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
        />
      )}
    </div>
  );
}

const styles = {
  container: {
    minHeight: '100vh',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    padding: '20px'
  },
  loading: {
    minHeight: '100vh',
    display: 'flex',
    flexDirection: 'column' as const,
    alignItems: 'center',
    justifyContent: 'center',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    color: 'white',
    fontSize: '18px'
  },
  spinner: {
    width: '50px',
    height: '50px',
    border: '5px solid rgba(255,255,255,0.3)',
    borderTop: '5px solid white',
    borderRadius: '50%',
    animation: 'spin 1s linear infinite',
    marginBottom: '20px'
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '30px',
    flexWrap: 'wrap' as const,
    gap: '20px'
  },
  title: {
    fontSize: '36px',
    fontWeight: 'bold',
    color: 'white',
    margin: 0
  },
  subtitle: {
    fontSize: '16px',
    color: 'rgba(255,255,255,0.9)',
    margin: '5px 0 0 0'
  },
  headerActions: {
    display: 'flex',
    gap: '10px'
  },
  createButton: {
    padding: '12px 24px',
    fontSize: '16px',
    fontWeight: 'bold',
    color: '#667eea',
    background: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    transition: 'transform 0.2s'
  },
  logoutButton: {
    padding: '12px 24px',
    fontSize: '16px',
    fontWeight: 'bold',
    color: 'white',
    background: 'rgba(255,255,255,0.2)',
    border: '2px solid white',
    borderRadius: '8px',
    cursor: 'pointer',
    transition: 'transform 0.2s'
  },
  stats: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
    gap: '20px',
    marginBottom: '30px'
  },
  statCard: {
    background: 'rgba(255,255,255,0.95)',
    borderRadius: '12px',
    padding: '20px',
    textAlign: 'center' as const,
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)'
  },
  statNumber: {
    fontSize: '36px',
    fontWeight: 'bold',
    color: '#667eea',
    marginBottom: '5px'
  },
  statLabel: {
    fontSize: '14px',
    color: '#666',
    textTransform: 'uppercase' as const,
    letterSpacing: '1px'
  },
  board: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
    gap: '20px',
    alignItems: 'start'
  }
};
