import { Request, Response, NextFunction } from 'express';
import { CreateTodoDto, UpdateTodoDto, MoveTodoDto, TodoStatus } from '../types/todo';

/**
 * Validation middleware for creating todos
 */
export function validateCreateTodo(req: Request, res: Response, next: NextFunction) {
  const body: CreateTodoDto = req.body;
  
  // Validate title
  if (!body.title) {
    return res.status(400).json({
      error: 'validation_error',
      message: 'Title is required',
    });
  }
  
  if (typeof body.title !== 'string') {
    return res.status(400).json({
      error: 'validation_error',
      message: 'Title must be a string',
    });
  }
  
  const trimmedTitle = body.title.trim();
  if (trimmedTitle.length === 0) {
    return res.status(400).json({
      error: 'validation_error',
      message: 'Title cannot be empty',
    });
  }
  
  if (trimmedTitle.length > 200) {
    return res.status(400).json({
      error: 'validation_error',
      message: 'Title cannot exceed 200 characters',
    });
  }
  
  // Validate description (optional)
  if (body.description !== undefined) {
    if (typeof body.description !== 'string') {
      return res.status(400).json({
        error: 'validation_error',
        message: 'Description must be a string',
      });
    }
    
    if (body.description.length > 1000) {
      return res.status(400).json({
        error: 'validation_error',
        message: 'Description cannot exceed 1000 characters',
      });
    }
  }
  
  // Validate status (optional)
  if (body.status !== undefined) {
    const validStatuses = Object.values(TodoStatus);
    if (!validStatuses.includes(body.status)) {
      return res.status(400).json({
        error: 'validation_error',
        message: `Status must be one of: ${validStatuses.join(', ')}`,
      });
    }
  }
  
  next();
}

/**
 * Validation middleware for updating todos
 */
export function validateUpdateTodo(req: Request, res: Response, next: NextFunction) {
  const body: UpdateTodoDto = req.body;
  
  // At least one field must be provided
  if (!body.title && !body.description && !body.status && body.position === undefined) {
    return res.status(400).json({
      error: 'validation_error',
      message: 'At least one field must be provided for update',
    });
  }
  
  // Validate title (optional)
  if (body.title !== undefined) {
    if (typeof body.title !== 'string') {
      return res.status(400).json({
        error: 'validation_error',
        message: 'Title must be a string',
      });
    }
    
    const trimmedTitle = body.title.trim();
    if (trimmedTitle.length === 0) {
      return res.status(400).json({
        error: 'validation_error',
        message: 'Title cannot be empty',
      });
    }
    
    if (trimmedTitle.length > 200) {
      return res.status(400).json({
        error: 'validation_error',
        message: 'Title cannot exceed 200 characters',
      });
    }
  }
  
  // Validate description (optional)
  if (body.description !== undefined) {
    if (typeof body.description !== 'string') {
      return res.status(400).json({
        error: 'validation_error',
        message: 'Description must be a string',
      });
    }
    
    if (body.description.length > 1000) {
      return res.status(400).json({
        error: 'validation_error',
        message: 'Description cannot exceed 1000 characters',
      });
    }
  }
  
  // Validate status (optional)
  if (body.status !== undefined) {
    const validStatuses = Object.values(TodoStatus);
    if (!validStatuses.includes(body.status)) {
      return res.status(400).json({
        error: 'validation_error',
        message: `Status must be one of: ${validStatuses.join(', ')}`,
      });
    }
  }
  
  // Validate position (optional)
  if (body.position !== undefined) {
    if (typeof body.position !== 'number' || body.position < 0) {
      return res.status(400).json({
        error: 'validation_error',
        message: 'Position must be a non-negative number',
      });
    }
  }
  
  next();
}

/**
 * Validation middleware for moving todos
 */
export function validateMoveTodo(req: Request, res: Response, next: NextFunction) {
  const body: MoveTodoDto = req.body;
  
  // Validate status (required)
  if (!body.status) {
    return res.status(400).json({
      error: 'validation_error',
      message: 'Status is required',
    });
  }
  
  const validStatuses = Object.values(TodoStatus);
  if (!validStatuses.includes(body.status)) {
    return res.status(400).json({
      error: 'validation_error',
      message: `Status must be one of: ${validStatuses.join(', ')}`,
    });
  }
  
  // Validate position (required)
  if (body.position === undefined) {
    return res.status(400).json({
      error: 'validation_error',
      message: 'Position is required',
    });
  }
  
  if (typeof body.position !== 'number' || body.position < 0) {
    return res.status(400).json({
      error: 'validation_error',
      message: 'Position must be a non-negative number',
    });
  }
  
  next();
}
