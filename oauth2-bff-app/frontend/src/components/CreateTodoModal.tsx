import { TodoPriority } from '../types/todo';
import TodoForm from './TodoForm';

interface CreateTodoModalProps {
  onClose: () => void;
  onCreate: (data: { title: string; description?: string; priority?: TodoPriority }) => void;
  isCreating?: boolean;
}

export default function CreateTodoModal({ onClose, onCreate, isCreating = false }: CreateTodoModalProps) {
  const handleSubmit = (data: { title: string; description?: string; priority?: TodoPriority }) => {
    onCreate(data);
  };

  return (
    <div 
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-[1000] p-4 md:p-5"
      onClick={isCreating ? undefined : onClose}
    >
      <div 
        className="bg-white rounded-2xl p-6 md:p-8 max-w-lg w-full max-h-[90vh] overflow-auto shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl md:text-2xl font-bold m-0 text-gray-800">
            ✨ Create New Task
          </h2>
          <button 
            onClick={onClose} 
            className="bg-none border-none text-2xl cursor-pointer text-gray-400 p-0 w-8 h-8 flex items-center justify-center rounded-full hover:bg-gray-100 transition-colors touch-manipulation min-w-[44px] min-h-[44px] md:min-w-[32px] md:min-h-[32px]"
            disabled={isCreating}
          >
            ✕
          </button>
        </div>

        <TodoForm onSubmit={handleSubmit} onCancel={onClose} isSubmitting={isCreating} />
      </div>
    </div>
  );
}


