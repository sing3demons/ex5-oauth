import { useDroppable } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { Todo } from '../types/todo';
import TodoCard from './TodoCard';

interface TodoColumnProps {
  id: string;
  title: string;
  todos: Todo[];
  onDelete: (id: string) => void;
  onUpdate: (id: string, data: any) => void;
}

export default function TodoColumn({ id, title, todos, onDelete, onUpdate }: TodoColumnProps) {
  const { setNodeRef, isOver } = useDroppable({ id });

  return (
    <div style={styles.column}>
      <div style={styles.header}>
        <h3 style={styles.title}>{title}</h3>
        <span style={styles.count}>{todos.length}</span>
      </div>

      <SortableContext items={todos.map(t => t.id)} strategy={verticalListSortingStrategy}>
        <div
          ref={setNodeRef}
          style={{
            ...styles.dropZone,
            ...(isOver ? styles.dropZoneActive : {})
          }}
        >
          {todos.length === 0 ? (
            <div style={styles.empty}>
              <p>No tasks yet</p>
              <p style={styles.emptyHint}>Drag tasks here</p>
            </div>
          ) : (
            todos.map(todo => (
              <TodoCard
                key={todo.id}
                todo={todo}
                onDelete={onDelete}
                onUpdate={onUpdate}
              />
            ))
          )}
        </div>
      </SortableContext>
    </div>
  );
}

const styles = {
  column: {
    background: 'rgba(255,255,255,0.95)',
    borderRadius: '12px',
    padding: '20px',
    minHeight: '500px',
    display: 'flex',
    flexDirection: 'column' as const,
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)'
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '20px',
    paddingBottom: '15px',
    borderBottom: '2px solid #e0e0e0'
  },
  title: {
    fontSize: '18px',
    fontWeight: 'bold',
    margin: 0,
    color: '#333'
  },
  count: {
    background: '#667eea',
    color: 'white',
    borderRadius: '12px',
    padding: '4px 12px',
    fontSize: '14px',
    fontWeight: 'bold'
  },
  dropZone: {
    flex: 1,
    minHeight: '400px',
    transition: 'background 0.2s',
    borderRadius: '8px',
    padding: '8px'
  },
  dropZoneActive: {
    background: 'rgba(102, 126, 234, 0.1)',
    border: '2px dashed #667eea'
  },
  empty: {
    display: 'flex',
    flexDirection: 'column' as const,
    alignItems: 'center',
    justifyContent: 'center',
    height: '100%',
    color: '#999',
    textAlign: 'center' as const
  },
  emptyHint: {
    fontSize: '14px',
    marginTop: '10px'
  }
};
