import { TodoStatus } from '../models/Todo';

export { TodoStatus };

export interface TodoResponse {
  _id: string;
  userId: string;
  title: string;
  description?: string;
  status: TodoStatus;
  position: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateTodoDto {
  title: string;
  description?: string;
  status?: TodoStatus;
}

export interface UpdateTodoDto {
  title?: string;
  description?: string;
  status?: TodoStatus;
  position?: number;
}

export interface MoveTodoDto {
  status: TodoStatus;
  position: number;
}
