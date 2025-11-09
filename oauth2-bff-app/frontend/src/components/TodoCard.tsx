import { useState } from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Todo, TodoPriority } from '../types/todo';
import TodoForm from './TodoForm';

interface TodoCardProps {
  todo: Todo;
  isDragging?: boolean;
  onDelete?: (id: string) => void;
  onUpdate?: (id: string, data: any) => void;
}

export default function TodoCard({ todo, isDragging, onDelete, onUpdate }: TodoCardProps) {
  const [isEditing, setIsEditing] = useState(false);

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging: isSortableDragging
  } = useSortable({
    id: todo.id,
    disabled: isEditing, // Disable dragging while editing
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isSortableDragging ? 0.5 : 1,
    cursor: isEditing ? 'default' : 'grab',
  };

  const handleUpdate = (data: { title: string; description?: string; priority?: TodoPriority }) => {
    if (onUpdate) {
      onUpdate(todo.id, data);
      setIsEditing(false);
    }
  };

  const handleCancel = () => {
    setIsEditing(false);
  };

  const priorityColors = {
    low: '#4caf50',
    medium: '#ff9800',
    high: '#f44336'
  };

  if (isEditing) {
    return (
      <div 
        ref={setNodeRef} 
        style={style}
        className="bg-white rounded-lg p-3 md:p-4 mb-3 shadow-sm border-2 border-transparent cursor-default"
      >
        <TodoForm
          initialData={todo}
          onSubmit={handleUpdate}
          onCancel={handleCancel}
        />
      </div>
    );
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`bg-white rounded-lg p-3 md:p-4 mb-3 shadow-sm transition-all border-2 border-transparent touch-none ${
        isDragging || isSortableDragging 
          ? 'shadow-2xl rotate-2 scale-105 cursor-grabbing border-indigo-600 z-[1000]' 
          : 'cursor-grab hover:shadow-md'
      }`}
      {...attributes}
      {...listeners}
    >
      <div className="flex justify-between items-center mb-3">
        <div
          className="text-[10px] md:text-xs font-bold text-white px-2 py-1 rounded uppercase"
          style={{ background: priorityColors[todo.priority] }}
        >
          {todo.priority}
        </div>
        <div className="flex gap-2">
          <button
            onClick={(e) => {
              e.stopPropagation();
              setIsEditing(true);
            }}
            className="bg-none border-none cursor-pointer text-base md:text-lg p-2 opacity-60 hover:opacity-100 transition-opacity touch-manipulation min-w-[44px] min-h-[44px] md:min-w-0 md:min-h-0 flex items-center justify-center"
            title="Edit"
          >
            ✏️
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onDelete?.(todo.id);
            }}
            className="bg-none border-none cursor-pointer text-base md:text-lg p-2 opacity-60 hover:opacity-100 transition-opacity touch-manipulation min-w-[44px] min-h-[44px] md:min-w-0 md:min-h-0 flex items-center justify-center"
            title="Delete"
          >
            🗑️
          </button>
        </div>
      </div>

      <h4 className="text-sm md:text-base font-bold m-0 mb-2 text-gray-800 break-words">
        {todo.title}
      </h4>

      {todo.description && (
        <p className="text-xs md:text-sm text-gray-600 m-0 mb-3 leading-relaxed break-words">
          {todo.description}
        </p>
      )}

      <div className="flex justify-between items-center pt-3 border-t border-gray-100">
        <span className="text-[11px] md:text-xs text-gray-400">
          {new Date(todo.createdAt).toLocaleDateString()}
        </span>
      </div>
    </div>
  );
}


