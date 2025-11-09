import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { HttpModule } from '@nestjs/axios';
import { AuthController } from './auth.controller';
import { AuthService } from './auth.service';
import { SessionService } from './session.service';
import { AuthGuard } from './guards/auth.guard';
import { RefreshGuard } from './guards/refresh.guard';
import { SharedModule } from '../shared/shared.module';

@Module({
  imports: [ConfigModule, HttpModule, SharedModule],
  controllers: [AuthController],
  providers: [AuthService, SessionService, AuthGuard, RefreshGuard],
  exports: [AuthGuard, RefreshGuard],
})
export class AuthModule {}
