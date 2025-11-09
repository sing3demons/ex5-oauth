import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import TodoForm from '../TodoForm';

describe('TodoForm Component', () => {
  const mockSubmit = vi.fn();
  const mockCancel = vi.fn();

  it('renders form with empty fields for new todo', () => {
    render(<TodoForm onSubmit={mockSubmit} onCancel={mockCancel} />);
    
    expect(screen.getByPlaceholderText(/enter task title/i)).toHaveValue('');
    expect(screen.getByPlaceholderText(/add more details/i)).toHaveValue('');
    expect(screen.getByText('Create Task')).toBeInTheDocument();
  });

  it('validates title is required', async () => {
    render(<TodoForm onSubmit={mockSubmit} onCancel={mockCancel} />);
    
    const submitButton = screen.getByRole('button', { name: /create task/i });
    expect(submitButton).toBeDisabled();
  });

  it('validates title max length (200 chars)', async () => {
    render(<TodoForm onSubmit={mockSubmit} onCancel={mockCancel} />);
    
    const titleInput = screen.getByPlaceholderText(/enter task title/i);
    const longTitle = 'a'.repeat(201);
    
    fireEvent.change(titleInput, { target: { value: longTitle } });
    fireEvent.blur(titleInput);
    
    const form = titleInput.closest('form');
    fireEvent.submit(form!);
    
    await waitFor(() => {
      expect(screen.getByText(/title must be 200 characters or less/i)).toBeInTheDocument();
    });
    
    expect(mockSubmit).not.toHaveBeenCalled();
  });

  it('submits form with valid data', async () => {
    render(<TodoForm onSubmit={mockSubmit} onCancel={mockCancel} />);
    
    const titleInput = screen.getByPlaceholderText(/enter task title/i);
    const descriptionInput = screen.getByPlaceholderText(/add more details/i);
    
    fireEvent.change(titleInput, { target: { value: 'Test Task' } });
    fireEvent.change(descriptionInput, { target: { value: 'Test Description' } });
    
    const form = titleInput.closest('form');
    fireEvent.submit(form!);
    
    await waitFor(() => {
      expect(mockSubmit).toHaveBeenCalledWith({
        title: 'Test Task',
        description: 'Test Description',
        priority: 'medium'
      });
    });
  });

  it('calls onCancel when cancel button is clicked', () => {
    render(<TodoForm onSubmit={mockSubmit} onCancel={mockCancel} />);
    
    const cancelButton = screen.getByText('Cancel');
    fireEvent.click(cancelButton);
    
    expect(mockCancel).toHaveBeenCalledTimes(1);
  });

  it('allows priority selection', () => {
    render(<TodoForm onSubmit={mockSubmit} onCancel={mockCancel} />);
    
    const highPriorityButton = screen.getByText(/🔴 High/i);
    fireEvent.click(highPriorityButton);
    
    expect(highPriorityButton).toHaveAttribute('aria-pressed', 'true');
  });
});
