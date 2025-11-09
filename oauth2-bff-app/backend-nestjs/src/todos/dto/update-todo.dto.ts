import { IsString, IsOptional, IsEnum } from 'class-validator';
import { TodoStatus, TodoPriority } from '../entities/todo.entity';

export class UpdateTodoDto {
  @IsString()
  @IsOptional()
  title?: string;

  @IsString()
  @IsOptional()
  description?: string;

  @IsEnum(['todo', 'in_progress', 'done'])
  @IsOptional()
  status?: TodoStatus;

  @IsEnum(['low', 'medium', 'high'])
  @IsOptional()
  priority?: TodoPriority;
}
