import { IsString, IsNotEmpty, IsOptional, IsEnum } from 'class-validator';
import { TodoPriority } from '../entities/todo.entity';

export class CreateTodoDto {
  @IsString()
  @IsNotEmpty()
  title!: string;

  @IsString()
  @IsOptional()
  description?: string;

  @IsEnum(['low', 'medium', 'high'])
  @IsOptional()
  priority?: TodoPriority;
}
