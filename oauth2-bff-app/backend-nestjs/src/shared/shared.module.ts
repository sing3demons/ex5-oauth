import { Module, Global } from '@nestjs/common';
import { HttpModule } from '@nestjs/axios';
import { CryptoService } from './services/crypto.service';
import { OidcService } from './services/oidc.service';
import { TokenService } from './services/token.service';

@Global()
@Module({
  imports: [HttpModule],
  providers: [CryptoService, OidcService, TokenService],
  exports: [CryptoService, OidcService, TokenService],
})
export class SharedModule {}
