import { Injectable, NotFoundException, ForbiddenException } from '@nestjs/common';
import { DatabaseService } from '../database/database.service';
import { TokenService } from '../shared/services/token.service';
import { Todo } from './entities/todo.entity';
import { CreateTodoDto } from './dto/create-todo.dto';
import { UpdateTodoDto } from './dto/update-todo.dto';
import { TodoStatus } from './entities/todo.entity';
import { v4 as uuidv4 } from 'uuid';

@Injectable()
export class TodosService {
  constructor(
    private readonly databaseService: DatabaseService,
    private readonly tokenService: TokenService,
  ) {}

  /**
   * Find all todos for a specific user
   * Sorted by createdAt descending (newest first)
   */
  async findAllByUser(userId: string): Promise<Todo[]> {
    const db = this.databaseService.getDatabase();
    const todos = await db.collection('todos').find({ userId }).sort({ createdAt: -1 }).toArray();

    return todos.map((todo) => ({
      id: todo.id,
      userId: todo.userId,
      title: todo.title,
      description: todo.description,
      status: todo.status,
      priority: todo.priority,
      createdAt: todo.createdAt,
      updatedAt: todo.updatedAt,
    }));
  }

  /**
   * Find a specific todo by id
   * Verifies ownership before returning
   */
  async findOne(userId: string, id: string): Promise<Todo> {
    return this.verifyOwnership(userId, id);
  }

  /**
   * Create a new todo for a user
   * Sets default status to 'todo' and priority to 'medium'
   */
  async create(userId: string, createTodoDto: CreateTodoDto): Promise<Todo> {
    const db = this.databaseService.getDatabase();
    const now = new Date();

    const todo: Todo = {
      id: uuidv4(),
      userId,
      title: createTodoDto.title,
      description: createTodoDto.description,
      status: 'todo',
      priority: createTodoDto.priority || 'medium',
      createdAt: now,
      updatedAt: now,
    };

    await db.collection('todos').insertOne(todo);

    return todo;
  }

  /**
   * Update a todo
   * Verifies ownership and validates title is not empty
   */
  async update(userId: string, id: string, updateTodoDto: UpdateTodoDto): Promise<Todo> {
    // Verify ownership first
    await this.verifyOwnership(userId, id);

    // Validate title is not empty if provided
    if (updateTodoDto.title !== undefined && updateTodoDto.title.trim() === '') {
      throw new ForbiddenException('Title cannot be empty');
    }

    const db = this.databaseService.getDatabase();
    const updateFields: any = {
      updatedAt: new Date(),
    };

    // Only update provided fields
    if (updateTodoDto.title !== undefined) {
      updateFields.title = updateTodoDto.title;
    }
    if (updateTodoDto.description !== undefined) {
      updateFields.description = updateTodoDto.description;
    }
    if (updateTodoDto.status !== undefined) {
      updateFields.status = updateTodoDto.status;
    }
    if (updateTodoDto.priority !== undefined) {
      updateFields.priority = updateTodoDto.priority;
    }

    await db.collection('todos').updateOne({ id }, { $set: updateFields });

    // Return updated todo
    const updatedTodo = await db.collection('todos').findOne({ id });
    if (!updatedTodo) {
      throw new NotFoundException(`Todo with id ${id} not found after update`);
    }

    return {
      id: updatedTodo.id,
      userId: updatedTodo.userId,
      title: updatedTodo.title,
      description: updatedTodo.description,
      status: updatedTodo.status,
      priority: updatedTodo.priority,
      createdAt: updatedTodo.createdAt,
      updatedAt: updatedTodo.updatedAt,
    };
  }

  /**
   * Delete a todo
   * Verifies ownership before deletion
   */
  async remove(userId: string, id: string): Promise<void> {
    // Verify ownership first
    await this.verifyOwnership(userId, id);

    const db = this.databaseService.getDatabase();
    await db.collection('todos').deleteOne({ id });
  }

  /**
   * Update todo status
   * Verifies ownership and validates status enum
   */
  async updateStatus(userId: string, id: string, status: TodoStatus): Promise<Todo> {
    // Verify ownership first
    await this.verifyOwnership(userId, id);

    // Validate status enum (should be validated by DTO, but double-check)
    const validStatuses: TodoStatus[] = ['todo', 'in_progress', 'done'];
    if (!validStatuses.includes(status)) {
      throw new ForbiddenException('Invalid status value');
    }

    const db = this.databaseService.getDatabase();
    await db.collection('todos').updateOne(
      { id },
      {
        $set: {
          status,
          updatedAt: new Date(),
        },
      },
    );

    // Return updated todo
    const updatedTodo = await db.collection('todos').findOne({ id });
    if (!updatedTodo) {
      throw new NotFoundException(`Todo with id ${id} not found after status update`);
    }

    return {
      id: updatedTodo.id,
      userId: updatedTodo.userId,
      title: updatedTodo.title,
      description: updatedTodo.description,
      status: updatedTodo.status,
      priority: updatedTodo.priority,
      createdAt: updatedTodo.createdAt,
      updatedAt: updatedTodo.updatedAt,
    };
  }

  /**
   * Verify that a todo belongs to the specified user
   * Throws NotFoundException if todo doesn't exist
   * Throws ForbiddenException if todo belongs to another user
   */
  private async verifyOwnership(userId: string, todoId: string): Promise<Todo> {
    const db = this.databaseService.getDatabase();
    const todo = await db.collection('todos').findOne({ id: todoId });

    if (!todo) {
      throw new NotFoundException(`Todo with id ${todoId} not found`);
    }

    if (todo.userId !== userId) {
      throw new ForbiddenException('You do not have permission to access this todo');
    }

    return {
      id: todo.id,
      userId: todo.userId,
      title: todo.title,
      description: todo.description,
      status: todo.status,
      priority: todo.priority,
      createdAt: todo.createdAt,
      updatedAt: todo.updatedAt,
    };
  }
}
