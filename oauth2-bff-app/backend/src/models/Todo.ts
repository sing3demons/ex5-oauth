import mongoose, { Schema, Document } from 'mongoose';

export enum TodoStatus {
  TODO = 'todo',
  IN_PROGRESS = 'in_progress',
  DONE = 'done'
}

export interface ITodo extends Document {
  userId: string;
  title: string;
  description?: string;
  status: TodoStatus;
  position: number;
  createdAt: Date;
  updatedAt: Date;
}

const TodoSchema = new Schema<ITodo>(
  {
    userId: {
      type: String,
      required: [true, 'User ID is required'],
      index: true,
    },
    title: {
      type: String,
      required: [true, 'Title is required'],
      trim: true,
      minlength: [1, 'Title must be at least 1 character'],
      maxlength: [200, 'Title cannot exceed 200 characters'],
    },
    description: {
      type: String,
      trim: true,
      maxlength: [1000, 'Description cannot exceed 1000 characters'],
      default: '',
    },
    status: {
      type: String,
      enum: {
        values: Object.values(TodoStatus),
        message: '{VALUE} is not a valid status',
      },
      default: TodoStatus.TODO,
      index: true,
    },
    position: {
      type: Number,
      default: 0,
      min: [0, 'Position cannot be negative'],
    },
  },
  {
    timestamps: true,
    toJSON: {
      transform: (_doc, ret: any) => {
        ret._id = ret._id.toString();
        return ret;
      },
    },
  }
);

// Compound indexes for performance
TodoSchema.index({ userId: 1, status: 1 });
TodoSchema.index({ userId: 1, position: 1 });
TodoSchema.index({ userId: 1, createdAt: -1 });

// Pre-save hook to set position for new todos
TodoSchema.pre('save', async function (next) {
  if (this.isNew && this.position === 0) {
    // Find the highest position for this user and status
    const Todo = mongoose.model<ITodo>('Todo');
    const maxPositionDoc = await Todo.findOne({
      userId: this.userId,
      status: this.status,
    })
      .sort({ position: -1 })
      .select('position')
      .lean();

    this.position = maxPositionDoc ? maxPositionDoc.position + 1 : 0;
  }
  next();
});

export const Todo = mongoose.model<ITodo>('Todo', TodoSchema);
