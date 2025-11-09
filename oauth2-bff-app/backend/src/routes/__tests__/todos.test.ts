import request from 'supertest';
import express from 'express';
import mongoose from 'mongoose';
import { MongoMemoryServer } from 'mongodb-memory-server';
import todoRoutes from '../todos';
import { Todo } from '../../models/Todo';
import { errorHandler } from '../../middleware/errorHandler';

// Mock the auth middleware before importing routes
jest.mock('../../middleware/auth', () => ({
  requireAuth: (req: any, res: any, next: any) => {
    // Check if Authorization header is present
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return res.status(401).json({
        error: 'unauthorized',
        message: 'No access token provided',
      });
    }
    
    // Mock user from token
    req.user = {
      id: 'test-user-123',
      email: 'test@example.com',
      name: 'Test User',
    };
    next();
  },
}));

// Create Express app for testing
const app = express();
app.use(express.json());
app.use('/api/todos', todoRoutes);
app.use(errorHandler);

let mongoServer: MongoMemoryServer;
const authHeader = { Authorization: 'Bearer test-token' };

beforeAll(async () => {
  mongoServer = await MongoMemoryServer.create();
  const mongoUri = mongoServer.getUri();
  await mongoose.connect(mongoUri);
}, 30000); // 30 second timeout for MongoDB Memory Server setup

afterAll(async () => {
  await mongoose.disconnect();
  await mongoServer.stop();
}, 30000);

beforeEach(async () => {
  await Todo.deleteMany({});
});

describe('Todo API', () => {
  describe('GET /api/todos', () => {
    it('should return empty array when no todos exist', async () => {
      const response = await request(app)
        .get('/api/todos')
        .set(authHeader);

      expect(response.status).toBe(200);
      expect(response.body).toEqual([]);
    });

    it('should return all todos for authenticated user', async () => {
      await Todo.create([
        {
          userId: 'test-user-123',
          title: 'Todo 1',
          description: 'Description 1',
          status: 'todo',
          position: 0,
        },
        {
          userId: 'test-user-123',
          title: 'Todo 2',
          status: 'in_progress',
          position: 0,
        },
      ]);

      const response = await request(app)
        .get('/api/todos')
        .set(authHeader);

      expect(response.status).toBe(200);
      expect(response.body).toHaveLength(2);
    });

    it('should only return todos for authenticated user', async () => {
      await Todo.create([
        {
          userId: 'test-user-123',
          title: 'My Todo',
          status: 'todo',
          position: 0,
        },
        {
          userId: 'other-user-456',
          title: 'Other Todo',
          status: 'todo',
          position: 0,
        },
      ]);

      const response = await request(app)
        .get('/api/todos')
        .set(authHeader);

      expect(response.status).toBe(200);
      expect(response.body).toHaveLength(1);
      expect(response.body[0].title).toBe('My Todo');
    });
  });

  describe('POST /api/todos', () => {
    it('should create a new todo', async () => {
      const newTodo = {
        title: 'New Todo',
        description: 'Test description',
      };

      const response = await request(app)
        .post('/api/todos')
        .set(authHeader)
        .send(newTodo);

      expect(response.status).toBe(201);
      expect(response.body.title).toBe('New Todo');
      expect(response.body.description).toBe('Test description');
      expect(response.body.status).toBe('todo');
      expect(response.body.userId).toBe('test-user-123');
    });

    it('should fail when title is missing', async () => {
      const response = await request(app)
        .post('/api/todos')
        .set(authHeader)
        .send({ description: 'No title' });

      expect(response.status).toBe(400);
      expect(response.body.error).toBe('validation_error');
    });

    it('should fail when title exceeds 200 characters', async () => {
      const response = await request(app)
        .post('/api/todos')
        .set(authHeader)
        .send({ title: 'a'.repeat(201) });

      expect(response.status).toBe(400);
      expect(response.body.error).toBe('validation_error');
    });
  });

  describe('PATCH /api/todos/:id', () => {
    it('should update todo title', async () => {
      const todo = await Todo.create({
        userId: 'test-user-123',
        title: 'Original Title',
        status: 'todo',
        position: 0,
      });

      const response = await request(app)
        .patch(`/api/todos/${todo._id}`)
        .set(authHeader)
        .send({ title: 'Updated Title' });

      expect(response.status).toBe(200);
      expect(response.body.title).toBe('Updated Title');
    });

    it('should fail when todo not found', async () => {
      const fakeId = new mongoose.Types.ObjectId();
      const response = await request(app)
        .patch(`/api/todos/${fakeId}`)
        .set(authHeader)
        .send({ title: 'Updated' });

      expect(response.status).toBe(404);
      expect(response.body.error).toBe('not_found');
    });

    it('should fail when user does not own todo', async () => {
      const todo = await Todo.create({
        userId: 'other-user-456',
        title: 'Other User Todo',
        status: 'todo',
        position: 0,
      });

      const response = await request(app)
        .patch(`/api/todos/${todo._id}`)
        .set(authHeader)
        .send({ title: 'Hacked' });

      expect(response.status).toBe(403);
      expect(response.body.error).toBe('forbidden');
    });
  });

  describe('DELETE /api/todos/:id', () => {
    it('should delete todo', async () => {
      const todo = await Todo.create({
        userId: 'test-user-123',
        title: 'To Delete',
        status: 'todo',
        position: 0,
      });

      const response = await request(app)
        .delete(`/api/todos/${todo._id}`)
        .set(authHeader);

      expect(response.status).toBe(204);

      const deletedTodo = await Todo.findById(todo._id);
      expect(deletedTodo).toBeNull();
    });

    it('should fail when user does not own todo', async () => {
      const todo = await Todo.create({
        userId: 'other-user-456',
        title: 'Other User Todo',
        status: 'todo',
        position: 0,
      });

      const response = await request(app)
        .delete(`/api/todos/${todo._id}`)
        .set(authHeader);

      expect(response.status).toBe(403);
      expect(response.body.error).toBe('forbidden');
    });
  });

  describe('PATCH /api/todos/:id/move', () => {
    it('should move todo to different status', async () => {
      const todo = await Todo.create({
        userId: 'test-user-123',
        title: 'Test Todo',
        status: 'todo',
        position: 0,
      });

      const response = await request(app)
        .patch(`/api/todos/${todo._id}/move`)
        .set(authHeader)
        .send({ status: 'in_progress', position: 1 });

      expect(response.status).toBe(200);
      expect(response.body.status).toBe('in_progress');
      expect(response.body.position).toBe(1);
    });

    it('should fail when status is missing', async () => {
      const todo = await Todo.create({
        userId: 'test-user-123',
        title: 'Test Todo',
        status: 'todo',
        position: 0,
      });

      const response = await request(app)
        .patch(`/api/todos/${todo._id}/move`)
        .set(authHeader)
        .send({ position: 1 });

      expect(response.status).toBe(400);
      expect(response.body.error).toBe('validation_error');
    });

    it('should fail when user does not own todo', async () => {
      const todo = await Todo.create({
        userId: 'other-user-456',
        title: 'Other User Todo',
        status: 'todo',
        position: 0,
      });

      const response = await request(app)
        .patch(`/api/todos/${todo._id}/move`)
        .set(authHeader)
        .send({ status: 'done', position: 0 });

      expect(response.status).toBe(403);
      expect(response.body.error).toBe('forbidden');
    });
  });
});
