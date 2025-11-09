import api from './api';
import { Todo, TodoStatus, TodoPriority } from '../types/todo';

export const todoService = {
  async getAll(): Promise<Todo[]> {
    const response = await api.get('/api/todos');
    return response.data;
  },

  async create(data: {
    title: string;
    description?: string;
    priority?: TodoPriority;
  }): Promise<Todo> {
    const response = await api.post('/api/todos', data);
    return response.data;
  },

  async update(
    id: string,
    data: {
      title?: string;
      description?: string;
      status?: TodoStatus;
      priority?: TodoPriority;
    }
  ): Promise<Todo> {
    const response = await api.put(`/api/todos/${id}`, data);
    return response.data;
  },

  async updateStatus(id: string, status: TodoStatus): Promise<Todo> {
    const response = await api.patch(`/api/todos/${id}/status`, { status });
    return response.data;
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/api/todos/${id}`);
  },
};
