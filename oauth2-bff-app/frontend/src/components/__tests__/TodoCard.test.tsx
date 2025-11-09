import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DndContext } from '@dnd-kit/core';
import TodoCard from '../TodoCard';
import { Todo } from '../../types/todo';

// Wrapper for DnD context
const DndWrapper = ({ children }: { children: React.ReactNode }) => (
  <DndContext>{children}</DndContext>
);

describe('TodoCard Component', () => {
  const mockTodo: Todo = {
    id: '1',
    userId: 'user1',
    title: 'Test Todo',
    description: 'Test Description',
    status: 'todo',
    priority: 'medium',
    createdAt: '2025-11-09T10:00:00Z',
    updatedAt: '2025-11-09T10:00:00Z'
  };

  const mockDelete = vi.fn();
  const mockUpdate = vi.fn();

  it('renders todo information', () => {
    render(
      <DndWrapper>
        <TodoCard todo={mockTodo} onDelete={mockDelete} onUpdate={mockUpdate} />
      </DndWrapper>
    );
    
    expect(screen.getByText('Test Todo')).toBeInTheDocument();
    expect(screen.getByText('Test Description')).toBeInTheDocument();
    expect(screen.getByText('medium')).toBeInTheDocument();
  });

  it('displays edit and delete buttons', () => {
    render(
      <DndWrapper>
        <TodoCard todo={mockTodo} onDelete={mockDelete} onUpdate={mockUpdate} />
      </DndWrapper>
    );
    
    expect(screen.getByTitle('Edit')).toBeInTheDocument();
    expect(screen.getByTitle('Delete')).toBeInTheDocument();
  });

  it('calls onDelete when delete button is clicked', () => {
    render(
      <DndWrapper>
        <TodoCard todo={mockTodo} onDelete={mockDelete} onUpdate={mockUpdate} />
      </DndWrapper>
    );
    
    const deleteButton = screen.getByTitle('Delete');
    fireEvent.click(deleteButton);
    
    expect(mockDelete).toHaveBeenCalledWith('1');
  });

  it('shows edit form when edit button is clicked', () => {
    render(
      <DndWrapper>
        <TodoCard todo={mockTodo} onDelete={mockDelete} onUpdate={mockUpdate} />
      </DndWrapper>
    );
    
    const editButton = screen.getByTitle('Edit');
    fireEvent.click(editButton);
    
    expect(screen.getByDisplayValue('Test Todo')).toBeInTheDocument();
    expect(screen.getByText('Update Task')).toBeInTheDocument();
  });
});
