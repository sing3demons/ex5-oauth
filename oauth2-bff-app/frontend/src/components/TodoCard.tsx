import { useState } from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Todo } from '../types/todo';

interface TodoCardProps {
  todo: Todo;
  isDragging?: boolean;
  onDelete?: (id: string) => void;
  onUpdate?: (id: string, data: any) => void;
}

export default function TodoCard({ todo, isDragging, onDelete, onUpdate }: TodoCardProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [title, setTitle] = useState(todo.title);
  const [description, setDescription] = useState(todo.description || '');

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging: isSortableDragging
  } = useSortable({ id: todo.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isSortableDragging ? 0.5 : 1
  };

  const handleSave = () => {
    if (onUpdate && title.trim()) {
      onUpdate(todo.id, {
        title: title.trim(),
        description: description.trim()
      });
      setIsEditing(false);
    }
  };

  const handleCancel = () => {
    setTitle(todo.title);
    setDescription(todo.description || '');
    setIsEditing(false);
  };

  const priorityColors = {
    low: '#4caf50',
    medium: '#ff9800',
    high: '#f44336'
  };

  if (isEditing) {
    return (
      <div ref={setNodeRef} style={{ ...styles.card, ...style }}>
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          style={styles.input}
          placeholder="Task title"
          autoFocus
        />
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          style={styles.textarea}
          placeholder="Description (optional)"
          rows={3}
        />
        <div style={styles.editActions}>
          <button onClick={handleSave} style={styles.saveButton}>
            Save
          </button>
          <button onClick={handleCancel} style={styles.cancelButton}>
            Cancel
          </button>
        </div>
      </div>
    );
  }

  return (
    <div
      ref={setNodeRef}
      style={{
        ...styles.card,
        ...style,
        ...(isDragging ? styles.cardDragging : {})
      }}
      {...attributes}
      {...listeners}
    >
      <div style={styles.cardHeader}>
        <div
          style={{
            ...styles.priorityBadge,
            background: priorityColors[todo.priority]
          }}
        >
          {todo.priority}
        </div>
        <div style={styles.actions}>
          <button
            onClick={(e) => {
              e.stopPropagation();
              setIsEditing(true);
            }}
            style={styles.actionButton}
            title="Edit"
          >
            ✏️
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onDelete?.(todo.id);
            }}
            style={styles.actionButton}
            title="Delete"
          >
            🗑️
          </button>
        </div>
      </div>

      <h4 style={styles.cardTitle}>{todo.title}</h4>

      {todo.description && (
        <p style={styles.cardDescription}>{todo.description}</p>
      )}

      <div style={styles.cardFooter}>
        <span style={styles.date}>
          {new Date(todo.createdAt).toLocaleDateString()}
        </span>
      </div>
    </div>
  );
}

const styles = {
  card: {
    background: 'white',
    borderRadius: '8px',
    padding: '16px',
    marginBottom: '12px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
    cursor: 'grab',
    transition: 'all 0.2s',
    border: '2px solid transparent'
  },
  cardDragging: {
    boxShadow: '0 8px 16px rgba(0,0,0,0.2)',
    transform: 'rotate(2deg)',
    cursor: 'grabbing'
  },
  cardHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '12px'
  },
  priorityBadge: {
    fontSize: '11px',
    fontWeight: 'bold',
    color: 'white',
    padding: '4px 8px',
    borderRadius: '4px',
    textTransform: 'uppercase' as const
  },
  actions: {
    display: 'flex',
    gap: '8px'
  },
  actionButton: {
    background: 'none',
    border: 'none',
    cursor: 'pointer',
    fontSize: '16px',
    padding: '4px',
    opacity: 0.6,
    transition: 'opacity 0.2s'
  },
  cardTitle: {
    fontSize: '16px',
    fontWeight: 'bold',
    margin: '0 0 8px 0',
    color: '#333',
    wordBreak: 'break-word' as const
  },
  cardDescription: {
    fontSize: '14px',
    color: '#666',
    margin: '0 0 12px 0',
    lineHeight: '1.5',
    wordBreak: 'break-word' as const
  },
  cardFooter: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingTop: '12px',
    borderTop: '1px solid #f0f0f0'
  },
  date: {
    fontSize: '12px',
    color: '#999'
  },
  input: {
    width: '100%',
    padding: '12px',
    fontSize: '16px',
    border: '2px solid #667eea',
    borderRadius: '8px',
    marginBottom: '12px',
    outline: 'none'
  },
  textarea: {
    width: '100%',
    padding: '12px',
    fontSize: '14px',
    border: '2px solid #e0e0e0',
    borderRadius: '8px',
    marginBottom: '12px',
    outline: 'none',
    fontFamily: 'inherit',
    resize: 'vertical' as const
  },
  editActions: {
    display: 'flex',
    gap: '8px',
    justifyContent: 'flex-end'
  },
  saveButton: {
    padding: '8px 16px',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
    fontWeight: 'bold'
  },
  cancelButton: {
    padding: '8px 16px',
    background: '#e0e0e0',
    color: '#333',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
    fontWeight: 'bold'
  }
};
