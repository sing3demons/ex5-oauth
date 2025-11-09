export type TodoStatus = 'todo' | 'in_progress' | 'done';

export type TodoPriority = 'low' | 'medium' | 'high';

export interface Todo {
  id: string;
  userId: string;
  title: string;
  description?: string;
  status: TodoStatus;
  priority: TodoPriority;
  createdAt: Date;
  updatedAt: Date;
}
