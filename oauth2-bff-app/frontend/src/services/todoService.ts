import axios from 'axios';
import { Todo, TodoStatus, TodoPriority } from '../types/todo';

const BFF_URL = import.meta.env.VITE_BFF_URL || 'http://localhost:3001';

export const todoService = {
  async getAll(accessToken: string): Promise<Todo[]> {
    const response = await axios.get(`${BFF_URL}/api/todos`, {
      headers: { Authorization: `Bearer ${accessToken}` }
    });
    return response.data;
  },

  async create(
    accessToken: string,
    data: { title: string; description?: string; priority?: TodoPriority }
  ): Promise<Todo> {
    const response = await axios.post(`${BFF_URL}/api/todos`, data, {
      headers: { Authorization: `Bearer ${accessToken}` }
    });
    return response.data;
  },

  async update(
    accessToken: string,
    id: string,
    data: { title?: string; description?: string; status?: TodoStatus; priority?: TodoPriority }
  ): Promise<Todo> {
    const response = await axios.put(`${BFF_URL}/api/todos/${id}`, data, {
      headers: { Authorization: `Bearer ${accessToken}` }
    });
    return response.data;
  },

  async updateStatus(accessToken: string, id: string, status: TodoStatus): Promise<Todo> {
    const response = await axios.patch(
      `${BFF_URL}/api/todos/${id}/status`,
      { status },
      { headers: { Authorization: `Bearer ${accessToken}` } }
    );
    return response.data;
  },

  async delete(accessToken: string, id: string): Promise<void> {
    await axios.delete(`${BFF_URL}/api/todos/${id}`, {
      headers: { Authorization: `Bearer ${accessToken}` }
    });
  }
};
