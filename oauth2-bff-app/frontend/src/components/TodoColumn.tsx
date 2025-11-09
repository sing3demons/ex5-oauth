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
  onAddClick?: () => void;
}

export default function TodoColumn({ id, title, todos, onDelete, onUpdate, onAddClick }: TodoColumnProps) {
  const { setNodeRef, isOver } = useDroppable({
    id,
    data: {
      type: 'column',
      status: id,
    },
  });
  const showAddButton = id === 'todo' && onAddClick;

  return (
    <div className="bg-white/95 rounded-xl p-4 md:p-5 lg:p-6 min-h-[400px] md:min-h-[500px] flex flex-col shadow-md">
      <div className="flex justify-between items-center mb-4 md:mb-5 pb-3 md:pb-4 border-b-2 border-gray-200 flex-wrap gap-2">
        <div className="flex items-center gap-3">
          <h3 className="text-base md:text-lg font-bold m-0 text-gray-800">
            {title}
          </h3>
          <span className="bg-indigo-600 text-white rounded-xl px-3 py-1 text-xs md:text-sm font-bold">
            {todos.length}
          </span>
        </div>
        {showAddButton && (
          <button 
            onClick={onAddClick} 
            className="px-3 py-2 md:px-4 md:py-2 text-xs md:text-sm font-bold bg-indigo-600 text-white rounded-md hover:bg-indigo-700 transition-all touch-manipulation min-h-[44px] md:min-h-0"
            title="Add new task"
          >
            + Add
          </button>
        )}
      </div>

      <SortableContext items={todos.map(t => t.id)} strategy={verticalListSortingStrategy}>
        <div
          ref={setNodeRef}
          className={`flex-1 min-h-[350px] md:min-h-[400px] transition-all duration-200 rounded-lg p-2 border-2 ${
            isOver 
              ? 'bg-indigo-50 border-indigo-400 border-dashed shadow-inner' 
              : 'border-transparent'
          }`}
        >
          {todos.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-gray-400 text-center">
              <p className="text-base md:text-lg mb-2">No tasks yet</p>
              <p className="text-sm md:text-base">
                {showAddButton ? 'Click "+ Add" to create a task' : 'Drag tasks here'}
              </p>
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


