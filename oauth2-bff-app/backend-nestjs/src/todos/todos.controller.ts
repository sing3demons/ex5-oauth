import {
  Controller,
  Get,
  Post,
  Put,
  Delete,
  Patch,
  Body,
  Param,
  UseGuards,
  HttpCode,
  HttpStatus,
} from '@nestjs/common';
import { TodosService } from './todos.service';
import { AuthGuard } from '../auth/guards/auth.guard';
import { User } from '../common/decorators/user.decorator';
import { CreateTodoDto } from './dto/create-todo.dto';
import { UpdateTodoDto } from './dto/update-todo.dto';
import { UpdateStatusDto } from './dto/update-status.dto';
import { Todo } from './entities/todo.entity';

@Controller('api/todos')
@UseGuards(AuthGuard)
export class TodosController {
  constructor(private readonly todosService: TodosService) {}

  /**
   * Get all todos for the authenticated user
   * Sorted by createdAt descending (newest first)
   */
  @Get()
  async findAll(@User() userId: string): Promise<Todo[]> {
    return this.todosService.findAllByUser(userId);
  }

  /**
   * Get a specific todo by id
   * Verifies ownership before returning
   */
  @Get(':id')
  async findOne(@User() userId: string, @Param('id') id: string): Promise<Todo> {
    return this.todosService.findOne(userId, id);
  }

  /**
   * Create a new todo
   * Returns 201 status code with created todo
   */
  @Post()
  @HttpCode(HttpStatus.CREATED)
  async create(@User() userId: string, @Body() createTodoDto: CreateTodoDto): Promise<Todo> {
    return this.todosService.create(userId, createTodoDto);
  }

  /**
   * Update a todo
   * Verifies ownership and updates allowed fields
   */
  @Put(':id')
  async update(
    @User() userId: string,
    @Param('id') id: string,
    @Body() updateTodoDto: UpdateTodoDto,
  ): Promise<Todo> {
    return this.todosService.update(userId, id, updateTodoDto);
  }

  /**
   * Delete a todo
   * Returns 204 No Content on success
   */
  @Delete(':id')
  @HttpCode(HttpStatus.NO_CONTENT)
  async remove(@User() userId: string, @Param('id') id: string): Promise<void> {
    return this.todosService.remove(userId, id);
  }

  /**
   * Update todo status (for drag & drop functionality)
   * Verifies ownership before updating
   */
  @Patch(':id/status')
  async updateStatus(
    @User() userId: string,
    @Param('id') id: string,
    @Body() updateStatusDto: UpdateStatusDto,
  ): Promise<Todo> {
    return this.todosService.updateStatus(userId, id, updateStatusDto.status);
  }
}
