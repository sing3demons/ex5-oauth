import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../src/app.module';
import { getAuthHeader, getMockAuthHeader, getTestUserId, cleanupTestData } from './helpers/auth.helper';

/**
 * Authenticated Todos E2E Tests
 * 
 * These tests require valid authentication tokens.
 * To enable these tests:
 * 1. Implement getTestAccessToken() in test/helpers/auth.helper.ts
 * 2. Set up test OAuth2 credentials in environment variables
 * 3. Remove .skip from the test suites
 */

describe('TodosController (Authenticated E2E)', () => {
  let app: INestApplication;
  let createdTodoId: string;

  beforeAll(async () => {
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [AppModule],
    }).compile();

    app = moduleFixture.createNestApplication();
    await app.init();
  });

  afterAll(async () => {
    // Clean up test data
    try {
      await cleanupTestData(getTestUserId());
    } catch (error) {
      // Ignore cleanup errors
    }
    await app.close();
  });

  describe.skip('CRUD Operations', () => {
    it('should create a new todo', async () => {
      const authHeader = await getAuthHeader();
      
      const response = await request(app.getHttpServer())
        .post('/api/todos')
        .set(authHeader)
        .send({
          title: 'Test Todo',
          description: 'This is a test todo',
          status: 'todo'
        })
        .expect(201);

      expect(response.body).toHaveProperty('id');
      expect(response.body).toHaveProperty('title', 'Test Todo');
      expect(response.body).toHaveProperty('description', 'This is a test todo');
      expect(response.body).toHaveProperty('status', 'todo');
      expect(response.body).toHaveProperty('userId', getTestUserId());
      expect(response.body).toHaveProperty('createdAt');
      expect(response.body).toHaveProperty('updatedAt');

      createdTodoId = response.body.id;
    });

    it('should get all todos for user', async () => {
      const authHeader = await getAuthHeader();
      
      const response = await request(app.getHttpServer())
        .get('/api/todos')
        .set(authHeader)
        .expect(200);

      expect(Array.isArray(response.body)).toBe(true);
      expect(response.body.length).toBeGreaterThan(0);
      
      const todo = response.body[0];
      expect(todo).toHaveProperty('id');
      expect(todo).toHaveProperty('title');
      expect(todo).toHaveProperty('status');
      expect(todo).toHaveProperty('userId', getTestUserId());
    });

    it('should get a specific todo by id', async () => {
      const authHeader = await getAuthHeader();
      
      const response = await request(app.getHttpServer())
        .get(`/api/todos/${createdTodoId}`)
        .set(authHeader)
        .expect(200);

      expect(response.body).toHaveProperty('id', createdTodoId);
      expect(response.body).toHaveProperty('title', 'Test Todo');
      expect(response.body).toHaveProperty('userId', getTestUserId());
    });

    it('should update a todo', async () => {
      const authHeader = await getAuthHeader();
      
      const response = await request(app.getHttpServer())
        .put(`/api/todos/${createdTodoId}`)
        .set(authHeader)
        .send({
          title: 'Updated Test Todo',
          description: 'Updated description'
        })
        .expect(200);

      expect(response.body).toHaveProperty('id', createdTodoId);
      expect(response.body).toHaveProperty('title', 'Updated Test Todo');
      expect(response.body).toHaveProperty('description', 'Updated description');
    });

    it('should update todo status', async () => {
      const authHeader = await getAuthHeader();
      
      const response = await request(app.getHttpServer())
        .patch(`/api/todos/${createdTodoId}/status`)
        .set(authHeader)
        .send({ status: 'in_progress' })
        .expect(200);

      expect(response.body).toHaveProperty('id', createdTodoId);
      expect(response.body).toHaveProperty('status', 'in_progress');
    });

    it('should delete a todo', async () => {
      const authHeader = await getAuthHeader();
      
      await request(app.getHttpServer())
        .delete(`/api/todos/${createdTodoId}`)
        .set(authHeader)
        .expect(204);

      // Verify todo is deleted
      await request(app.getHttpServer())
        .get(`/api/todos/${createdTodoId}`)
        .set(authHeader)
        .expect(404);
    });
  });

  describe.skip('Validation', () => {
    it('should reject todo without title', async () => {
      const authHeader = await getAuthHeader();
      
      await request(app.getHttpServer())
        .post('/api/todos')
        .set(authHeader)
        .send({
          description: 'No title'
        })
        .expect(400);
    });

    it('should reject todo with title too long', async () => {
      const authHeader = await getAuthHeader();
      
      await request(app.getHttpServer())
        .post('/api/todos')
        .set(authHeader)
        .send({
          title: 'a'.repeat(201),
          description: 'Test'
        })
        .expect(400);
    });

    it('should reject todo with description too long', async () => {
      const authHeader = await getAuthHeader();
      
      await request(app.getHttpServer())
        .post('/api/todos')
        .set(authHeader)
        .send({
          title: 'Test',
          description: 'a'.repeat(1001)
        })
        .expect(400);
    });

    it('should reject invalid status', async () => {
      const authHeader = await getAuthHeader();
      
      // Create a todo first
      const createResponse = await request(app.getHttpServer())
        .post('/api/todos')
        .set(authHeader)
        .send({
          title: 'Test Todo',
          description: 'Test'
        })
        .expect(201);

      const todoId = createResponse.body.id;

      // Try to update with invalid status
      await request(app.getHttpServer())
        .patch(`/api/todos/${todoId}/status`)
        .set(authHeader)
        .send({ status: 'invalid_status' })
        .expect(400);

      // Clean up
      await request(app.getHttpServer())
        .delete(`/api/todos/${todoId}`)
        .set(authHeader);
    });
  });

  describe.skip('Authorization', () => {
    it('should not allow accessing another user\'s todo', async () => {
      const authHeader = await getAuthHeader();
      
      // Try to access a todo with a different user's ID
      await request(app.getHttpServer())
        .get('/api/todos/other-user-todo-id')
        .set(authHeader)
        .expect(404); // Should return 404 (not found) or 403 (forbidden)
    });

    it('should not allow updating another user\'s todo', async () => {
      const authHeader = await getAuthHeader();
      
      await request(app.getHttpServer())
        .put('/api/todos/other-user-todo-id')
        .set(authHeader)
        .send({ title: 'Hacked' })
        .expect((res) => {
          expect([403, 404]).toContain(res.status);
        });
    });

    it('should not allow deleting another user\'s todo', async () => {
      const authHeader = await getAuthHeader();
      
      await request(app.getHttpServer())
        .delete('/api/todos/other-user-todo-id')
        .set(authHeader)
        .expect((res) => {
          expect([403, 404]).toContain(res.status);
        });
    });
  });

  describe.skip('Status Transitions', () => {
    let todoId: string;

    beforeEach(async () => {
      const authHeader = await getAuthHeader();
      const response = await request(app.getHttpServer())
        .post('/api/todos')
        .set(authHeader)
        .send({
          title: 'Status Test Todo',
          description: 'Testing status transitions'
        });
      todoId = response.body.id;
    });

    afterEach(async () => {
      const authHeader = await getAuthHeader();
      await request(app.getHttpServer())
        .delete(`/api/todos/${todoId}`)
        .set(authHeader);
    });

    it('should transition from todo to in_progress', async () => {
      const authHeader = await getAuthHeader();
      
      const response = await request(app.getHttpServer())
        .patch(`/api/todos/${todoId}/status`)
        .set(authHeader)
        .send({ status: 'in_progress' })
        .expect(200);

      expect(response.body.status).toBe('in_progress');
    });

    it('should transition from in_progress to done', async () => {
      const authHeader = await getAuthHeader();
      
      // First move to in_progress
      await request(app.getHttpServer())
        .patch(`/api/todos/${todoId}/status`)
        .set(authHeader)
        .send({ status: 'in_progress' });

      // Then move to done
      const response = await request(app.getHttpServer())
        .patch(`/api/todos/${todoId}/status`)
        .set(authHeader)
        .send({ status: 'done' })
        .expect(200);

      expect(response.body.status).toBe('done');
    });

    it('should allow moving back from done to todo', async () => {
      const authHeader = await getAuthHeader();
      
      // Move to done
      await request(app.getHttpServer())
        .patch(`/api/todos/${todoId}/status`)
        .set(authHeader)
        .send({ status: 'done' });

      // Move back to todo
      const response = await request(app.getHttpServer())
        .patch(`/api/todos/${todoId}/status`)
        .set(authHeader)
        .send({ status: 'todo' })
        .expect(200);

      expect(response.body.status).toBe('todo');
    });
  });

  describe.skip('Sorting and Filtering', () => {
    beforeAll(async () => {
      const authHeader = await getAuthHeader();
      
      // Create multiple todos
      await request(app.getHttpServer())
        .post('/api/todos')
        .set(authHeader)
        .send({ title: 'Todo 1', description: 'First', status: 'todo' });

      await request(app.getHttpServer())
        .post('/api/todos')
        .set(authHeader)
        .send({ title: 'Todo 2', description: 'Second', status: 'in_progress' });

      await request(app.getHttpServer())
        .post('/api/todos')
        .set(authHeader)
        .send({ title: 'Todo 3', description: 'Third', status: 'done' });
    });

    it('should return todos sorted by createdAt descending', async () => {
      const authHeader = await getAuthHeader();
      
      const response = await request(app.getHttpServer())
        .get('/api/todos')
        .set(authHeader)
        .expect(200);

      expect(response.body.length).toBeGreaterThanOrEqual(3);
      
      // Check if sorted by createdAt descending (newest first)
      for (let i = 0; i < response.body.length - 1; i++) {
        const current = new Date(response.body[i].createdAt);
        const next = new Date(response.body[i + 1].createdAt);
        expect(current.getTime()).toBeGreaterThanOrEqual(next.getTime());
      }
    });
  });
});
