import { Injectable, OnModuleInit, OnModuleDestroy, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { MongoClient, Db } from 'mongodb';

@Injectable()
export class DatabaseService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(DatabaseService.name);
  private client!: MongoClient;
  private db!: Db;

  constructor(private readonly configService: ConfigService) {}

  async onModuleInit(): Promise<void> {
    await this.connect();
  }

  async onModuleDestroy(): Promise<void> {
    await this.disconnect();
  }

  private async connect(): Promise<void> {
    try {
      const uri = this.configService.get<string>('database.uri');
      const dbName = this.configService.get<string>('database.name');

      if (!uri || !dbName) {
        throw new Error('MongoDB configuration is missing');
      }

      this.logger.log(`Connecting to MongoDB at ${uri}`);
      this.client = new MongoClient(uri);
      await this.client.connect();
      this.db = this.client.db(dbName);
      this.logger.log(`Connected to MongoDB database: ${dbName}`);

      await this.createIndexes();
    } catch (error) {
      this.logger.error('Failed to connect to MongoDB', error);
      throw error;
    }
  }

  private async createIndexes(): Promise<void> {
    try {
      this.logger.log('Creating database indexes...');
      await this.db
        .collection('todos')
        .createIndex({ userId: 1, createdAt: -1 }, { name: 'userId_createdAt_idx' });
      this.logger.log('Database indexes created successfully');
    } catch (error: any) {
      // Ignore error if index already exists
      if (error.code === 85 || error.codeName === 'IndexOptionsConflict') {
        this.logger.log('Index already exists, skipping creation');
      } else {
        this.logger.error('Failed to create indexes', error);
        throw error;
      }
    }
  }

  getDatabase(): Db {
    if (!this.db) {
      throw new Error('Database not initialized');
    }
    return this.db;
  }

  async disconnect(): Promise<void> {
    if (this.client) {
      this.logger.log('Disconnecting from MongoDB...');
      await this.client.close();
      this.logger.log('Disconnected from MongoDB');
    }
  }
}
