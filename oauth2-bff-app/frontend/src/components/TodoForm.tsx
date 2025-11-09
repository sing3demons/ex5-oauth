import { useState, useEffect } from 'react';
import { Todo, TodoPriority } from '../types/todo';

interface TodoFormProps {
  initialData?: Todo;
  onSubmit: (data: { title: string; description?: string; priority?: TodoPriority }) => void;
  onCancel: () => void;
  isSubmitting?: boolean;
}

export default function TodoForm({ initialData, onSubmit, onCancel, isSubmitting = false }: TodoFormProps) {
  const [title, setTitle] = useState(initialData?.title || '');
  const [description, setDescription] = useState(initialData?.description || '');
  const [priority, setPriority] = useState<TodoPriority>(initialData?.priority || 'medium');
  const [errors, setErrors] = useState<{ title?: string; description?: string }>({});

  useEffect(() => {
    if (initialData) {
      setTitle(initialData.title);
      setDescription(initialData.description || '');
      setPriority(initialData.priority);
    }
  }, [initialData]);

  const validateForm = (): boolean => {
    const newErrors: { title?: string; description?: string } = {};

    // Title validation: required, 1-200 chars
    if (!title.trim()) {
      newErrors.title = 'Title is required';
    } else if (title.trim().length > 200) {
      newErrors.title = 'Title must be 200 characters or less';
    }

    // Description validation: max 1000 chars
    if (description.trim().length > 1000) {
      newErrors.description = 'Description must be 1000 characters or less';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    onSubmit({
      title: title.trim(),
      description: description.trim() || undefined,
      priority
    });
  };

  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTitle(e.target.value);
    if (errors.title) {
      setErrors({ ...errors, title: undefined });
    }
  };

  const handleDescriptionChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setDescription(e.target.value);
    if (errors.description) {
      setErrors({ ...errors, description: undefined });
    }
  };

  const isEditing = !!initialData;

  return (
    <form onSubmit={handleSubmit} className="w-full">
      <div className="mb-4 md:mb-5">
        <label htmlFor="title" className="block text-sm md:text-base font-bold text-gray-800 mb-2">
          Title <span className="text-red-500">*</span>
        </label>
        <input
          id="title"
          type="text"
          value={title}
          onChange={handleTitleChange}
          className={`w-full p-3 md:p-3 text-base md:text-base border-2 rounded-lg outline-none transition-colors box-border ${
            errors.title ? 'border-red-500' : 'border-gray-200 focus:border-indigo-600'
          }`}
          placeholder="Enter task title (max 200 characters)"
          autoFocus
          maxLength={200}
        />
        {errors.title && <div className="text-xs md:text-sm text-red-500 mt-1">{errors.title}</div>}
        <div className="text-xs md:text-sm text-gray-400 text-right mt-1">
          {title.length}/200
        </div>
      </div>

      <div className="mb-4 md:mb-5">
        <label htmlFor="description" className="block text-sm md:text-base font-bold text-gray-800 mb-2">
          Description
        </label>
        <textarea
          id="description"
          value={description}
          onChange={handleDescriptionChange}
          className={`w-full p-3 md:p-3 text-sm md:text-base border-2 rounded-lg outline-none font-inherit resize-y transition-colors box-border ${
            errors.description ? 'border-red-500' : 'border-gray-200 focus:border-indigo-600'
          }`}
          placeholder="Add more details (optional, max 1000 characters)"
          rows={4}
          maxLength={1000}
        />
        {errors.description && <div className="text-xs md:text-sm text-red-500 mt-1">{errors.description}</div>}
        <div className="text-xs md:text-sm text-gray-400 text-right mt-1">
          {description.length}/1000
        </div>
      </div>

      <div className="mb-4 md:mb-5">
        <label className="block text-sm md:text-base font-bold text-gray-800 mb-2">Priority</label>
        <div className="grid grid-cols-3 gap-2 md:gap-3">
          <button
            type="button"
            onClick={() => setPriority('low')}
            className={`p-3 md:p-3 text-xs md:text-sm font-bold bg-white border-2 rounded-lg cursor-pointer transition-all touch-manipulation min-h-[44px] ${
              priority === 'low' 
                ? 'bg-indigo-50 scale-105 border-green-500' 
                : 'border-green-500 hover:scale-102'
            }`}
            aria-pressed={priority === 'low'}
          >
            🟢 Low
          </button>
          <button
            type="button"
            onClick={() => setPriority('medium')}
            className={`p-3 md:p-3 text-xs md:text-sm font-bold bg-white border-2 rounded-lg cursor-pointer transition-all touch-manipulation min-h-[44px] ${
              priority === 'medium' 
                ? 'bg-indigo-50 scale-105 border-orange-500' 
                : 'border-orange-500 hover:scale-102'
            }`}
            aria-pressed={priority === 'medium'}
          >
            🟡 Medium
          </button>
          <button
            type="button"
            onClick={() => setPriority('high')}
            className={`p-3 md:p-3 text-xs md:text-sm font-bold bg-white border-2 rounded-lg cursor-pointer transition-all touch-manipulation min-h-[44px] ${
              priority === 'high' 
                ? 'bg-indigo-50 scale-105 border-red-500' 
                : 'border-red-500 hover:scale-102'
            }`}
            aria-pressed={priority === 'high'}
          >
            🔴 High
          </button>
        </div>
      </div>

      <div className="flex flex-col sm:flex-row gap-3 justify-end mt-6">
        <button 
          type="button" 
          onClick={onCancel} 
          className="px-6 py-3 text-sm md:text-base font-bold bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors touch-manipulation min-h-[44px] order-2 sm:order-1"
          disabled={isSubmitting}
        >
          Cancel
        </button>
        <button
          type="submit"
          className={`px-6 py-3 text-sm md:text-base font-bold bg-gradient-to-br from-indigo-500 to-purple-600 text-white rounded-lg transition-transform touch-manipulation min-h-[44px] order-1 sm:order-2 flex items-center justify-center gap-2 ${
            !title.trim() || isSubmitting
              ? 'opacity-50 cursor-not-allowed' 
              : 'hover:scale-102'
          }`}
          disabled={!title.trim() || isSubmitting}
        >
          {isSubmitting ? (
            <>
              <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
              <span>Saving...</span>
            </>
          ) : (
            <span>{isEditing ? 'Update Task' : 'Create Task'}</span>
          )}
        </button>
      </div>
    </form>
  );
}


