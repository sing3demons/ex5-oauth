import express, { Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { requireAuth } from '../middleware/auth';
import { decodeJWT } from '../utils/oidc';
import { Todo, CreateTodoRequest, UpdateTodoRequest } from '../types/todo';
import { getDB } from '../db/mongodb';

const router = express.Router();

// Helper to get user ID from token
function getUserIdFromToken(authHeader: string): string | null {
  try {
    const token = authHeader.replace('Bearer ', '');
    const claims = decodeJWT(token);
    return claims.sub || claims.user_id || null;
  } catch {
    return null;
  }
}

/**
 * GET /api/todos
 * Get all todos for current user
 */
router.get('/', requireAuth, async (req: Request, res: Response) => {
  const userId = getUserIdFromToken(req.headers.authorization!);
  
  if (!userId) {
    return res.status(401).json({ error: 'Invalid token' });
  }
  
  try {
    const db = getDB();
    const userTodos = await db
      .collection('todos')
      .find({ userId })
      .sort({ createdAt: -1 })
      .toArray();
    
    res.json(userTodos);
  } catch (error) {
    console.error('Get todos error:', error);
    res.status(500).json({ error: 'Failed to fetch todos' });
  }
});

/**
 * GET /api/todos/:id
 * Get specific todo
 */
router.get('/:id', requireAuth, async (req: Request, res: Response) => {
  const userId = getUserIdFromToken(req.headers.authorization!);
  
  try {
    const db = getDB();
    const todo = await db.collection('todos').findOne({ id: req.params.id });
    
    if (!todo) {
      return res.status(404).json({ error: 'Todo not found' });
    }
    
    if (todo.userId !== userId) {
      return res.status(403).json({ error: 'Forbidden' });
    }
    
    res.json(todo);
  } catch (error) {
    console.error('Get todo error:', error);
    res.status(500).json({ error: 'Failed to fetch todo' });
  }
});

/**
 * POST /api/todos
 * Create new todo
 */
router.post('/', requireAuth, async (req: Request, res: Response) => {
  const userId = getUserIdFromToken(req.headers.authorization!);
  
  if (!userId) {
    return res.status(401).json({ error: 'Invalid token' });
  }
  
  const body: CreateTodoRequest = req.body;
  
  if (!body.title || body.title.trim() === '') {
    return res.status(400).json({ error: 'Title is required' });
  }
  
  const todo: Todo = {
    id: uuidv4(),
    userId,
    title: body.title.trim(),
    description: body.description?.trim(),
    status: 'todo',
    priority: body.priority || 'medium',
    createdAt: new Date(),
    updatedAt: new Date()
  };
  
  try {
    const db = getDB();
    await db.collection('todos').insertOne(todo);
    res.status(201).json(todo);
  } catch (error) {
    console.error('Create todo error:', error);
    res.status(500).json({ error: 'Failed to create todo' });
  }
});

/**
 * PUT /api/todos/:id
 * Update todo
 */
router.put('/:id', requireAuth, async (req: Request, res: Response) => {
  const userId = getUserIdFromToken(req.headers.authorization!);
  
  try {
    const db = getDB();
    const todo = await db.collection('todos').findOne({ id: req.params.id });
    
    if (!todo) {
      return res.status(404).json({ error: 'Todo not found' });
    }
    
    if (todo.userId !== userId) {
      return res.status(403).json({ error: 'Forbidden' });
    }
    
    const body: UpdateTodoRequest = req.body;
    const updates: any = { updatedAt: new Date() };
    
    if (body.title !== undefined) {
      if (body.title.trim() === '') {
        return res.status(400).json({ error: 'Title cannot be empty' });
      }
      updates.title = body.title.trim();
    }
    
    if (body.description !== undefined) {
      updates.description = body.description.trim();
    }
    
    if (body.status !== undefined) {
      updates.status = body.status;
    }
    
    if (body.priority !== undefined) {
      updates.priority = body.priority;
    }
    
    await db.collection('todos').updateOne(
      { id: req.params.id },
      { $set: updates }
    );
    
    const updated = await db.collection('todos').findOne({ id: req.params.id });
    res.json(updated);
  } catch (error) {
    console.error('Update todo error:', error);
    res.status(500).json({ error: 'Failed to update todo' });
  }
});

/**
 * DELETE /api/todos/:id
 * Delete todo
 */
router.delete('/:id', requireAuth, async (req: Request, res: Response) => {
  const userId = getUserIdFromToken(req.headers.authorization!);
  
  try {
    const db = getDB();
    const todo = await db.collection('todos').findOne({ id: req.params.id });
    
    if (!todo) {
      return res.status(404).json({ error: 'Todo not found' });
    }
    
    if (todo.userId !== userId) {
      return res.status(403).json({ error: 'Forbidden' });
    }
    
    await db.collection('todos').deleteOne({ id: req.params.id });
    res.status(204).send();
  } catch (error) {
    console.error('Delete todo error:', error);
    res.status(500).json({ error: 'Failed to delete todo' });
  }
});

/**
 * PATCH /api/todos/:id/status
 * Update todo status (for drag & drop)
 */
router.patch('/:id/status', requireAuth, async (req: Request, res: Response) => {
  const userId = getUserIdFromToken(req.headers.authorization!);
  
  try {
    const db = getDB();
    const todo = await db.collection('todos').findOne({ id: req.params.id });
    
    if (!todo) {
      return res.status(404).json({ error: 'Todo not found' });
    }
    
    if (todo.userId !== userId) {
      return res.status(403).json({ error: 'Forbidden' });
    }
    
    const { status } = req.body;
    
    if (!['todo', 'in_progress', 'done'].includes(status)) {
      return res.status(400).json({ error: 'Invalid status' });
    }
    
    await db.collection('todos').updateOne(
      { id: req.params.id },
      { $set: { status, updatedAt: new Date() } }
    );
    
    const updated = await db.collection('todos').findOne({ id: req.params.id });
    res.json(updated);
  } catch (error) {
    console.error('Update status error:', error);
    res.status(500).json({ error: 'Failed to update status' });
  }
});

export default router;
