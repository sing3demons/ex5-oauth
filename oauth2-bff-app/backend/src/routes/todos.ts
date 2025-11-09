import express, { Request, Response, NextFunction } from 'express';
import { requireAuth } from '../middleware/auth';
import { validateCreateTodo, validateUpdateTodo, validateMoveTodo } from '../middleware/validation';
import { csrfProtection } from '../middleware/csrf';
import { CreateTodoDto, UpdateTodoDto, MoveTodoDto } from '../types/todo';
import { Todo } from '../models/Todo';
import mongoose from 'mongoose';

const router = express.Router();

/**
 * GET /api/todos
 * Get all todos for current user
 */
router.get('/', requireAuth, async (req: Request, res: Response, next: NextFunction) => {
  const userId = req.user?.id;
  
  if (!userId) {
    return res.status(401).json({ 
      error: 'unauthorized',
      message: 'Invalid token' 
    });
  }
  
  try {
    const todos = await Todo.find({ userId })
      .sort({ status: 1, position: 1 })
      .lean();
    
    res.json(todos);
  } catch (error) {
    next(error);
  }
});

/**
 * POST /api/todos
 * Create new todo
 */
router.post('/', csrfProtection, requireAuth, validateCreateTodo, async (req: Request, res: Response, next: NextFunction) => {
  const userId = req.user?.id;
  
  if (!userId) {
    return res.status(401).json({ 
      error: 'unauthorized',
      message: 'Invalid token' 
    });
  }
  
  try {
    const body: CreateTodoDto = req.body;
    
    // Validation is handled by Mongoose schema
    const todo = new Todo({
      userId,
      title: body.title,
      description: body.description,
      status: body.status || 'todo',
    });
    
    await todo.save();
    
    res.status(201).json(todo.toJSON());
  } catch (error) {
    if (error instanceof mongoose.Error.ValidationError) {
      return res.status(400).json({
        error: 'validation_error',
        message: error.message,
        details: error.errors,
      });
    }
    next(error);
  }
});

/**
 * PATCH /api/todos/:id
 * Update todo
 */
router.patch('/:id', csrfProtection, requireAuth, validateUpdateTodo, async (req: Request, res: Response, next: NextFunction) => {
  const userId = req.user?.id;
  const { id } = req.params;
  
  if (!userId) {
    return res.status(401).json({ 
      error: 'unauthorized',
      message: 'Invalid token' 
    });
  }
  
  // Validate MongoDB ObjectId
  if (!mongoose.Types.ObjectId.isValid(id)) {
    return res.status(400).json({
      error: 'validation_error',
      message: 'Invalid todo ID format',
    });
  }
  
  try {
    const todo = await Todo.findById(id);
    
    if (!todo) {
      return res.status(404).json({ 
        error: 'not_found',
        message: 'Todo not found' 
      });
    }
    
    if (todo.userId !== userId) {
      return res.status(403).json({ 
        error: 'forbidden',
        message: 'You do not have permission to update this todo' 
      });
    }
    
    const body: UpdateTodoDto = req.body;
    
    // Update only provided fields
    if (body.title !== undefined) {
      todo.title = body.title;
    }
    if (body.description !== undefined) {
      todo.description = body.description;
    }
    if (body.status !== undefined) {
      todo.status = body.status;
    }
    if (body.position !== undefined) {
      todo.position = body.position;
    }
    
    await todo.save();
    
    res.json(todo.toJSON());
  } catch (error) {
    if (error instanceof mongoose.Error.ValidationError) {
      return res.status(400).json({
        error: 'validation_error',
        message: error.message,
        details: error.errors,
      });
    }
    next(error);
  }
});

/**
 * DELETE /api/todos/:id
 * Delete todo
 */
router.delete('/:id', csrfProtection, requireAuth, async (req: Request, res: Response, next: NextFunction) => {
  const userId = req.user?.id;
  const { id } = req.params;
  
  if (!userId) {
    return res.status(401).json({ 
      error: 'unauthorized',
      message: 'Invalid token' 
    });
  }
  
  // Validate MongoDB ObjectId
  if (!mongoose.Types.ObjectId.isValid(id)) {
    return res.status(400).json({
      error: 'validation_error',
      message: 'Invalid todo ID format',
    });
  }
  
  try {
    const todo = await Todo.findById(id);
    
    if (!todo) {
      return res.status(404).json({ 
        error: 'not_found',
        message: 'Todo not found' 
      });
    }
    
    if (todo.userId !== userId) {
      return res.status(403).json({ 
        error: 'forbidden',
        message: 'You do not have permission to delete this todo' 
      });
    }
    
    await todo.deleteOne();
    
    res.status(204).send();
  } catch (error) {
    next(error);
  }
});

/**
 * PATCH /api/todos/:id/move
 * Move todo to different status/position (for drag & drop)
 */
router.patch('/:id/move', csrfProtection, requireAuth, validateMoveTodo, async (req: Request, res: Response, next: NextFunction) => {
  const userId = req.user?.id;
  const { id } = req.params;
  
  if (!userId) {
    return res.status(401).json({ 
      error: 'unauthorized',
      message: 'Invalid token' 
    });
  }
  
  // Validate MongoDB ObjectId
  if (!mongoose.Types.ObjectId.isValid(id)) {
    return res.status(400).json({
      error: 'validation_error',
      message: 'Invalid todo ID format',
    });
  }
  
  try {
    const todo = await Todo.findById(id);
    
    if (!todo) {
      return res.status(404).json({ 
        error: 'not_found',
        message: 'Todo not found' 
      });
    }
    
    if (todo.userId !== userId) {
      return res.status(403).json({ 
        error: 'forbidden',
        message: 'You do not have permission to move this todo' 
      });
    }
    
    const body: MoveTodoDto = req.body;
    
    // Update todo status and position (validation done by middleware)
    todo.status = body.status;
    todo.position = body.position;
    
    await todo.save();
    
    res.json(todo.toJSON());
  } catch (error) {
    if (error instanceof mongoose.Error.ValidationError) {
      return res.status(400).json({
        error: 'validation_error',
        message: error.message,
        details: error.errors,
      });
    }
    next(error);
  }
});

export default router;
