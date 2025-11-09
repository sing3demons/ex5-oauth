import { IsEnum } from 'class-validator';
import { TodoStatus } from '../entities/todo.entity';

export class UpdateStatusDto {
  @IsEnum(['todo', 'in_progress', 'done'])
  status!: TodoStatus;
}
