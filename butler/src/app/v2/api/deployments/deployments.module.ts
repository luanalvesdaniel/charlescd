/*
 * Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { CdConfigurationEntity } from '../../../v1/api/configurations/entity';
import { CdConfigurationsRepository } from '../../../v1/api/configurations/repository';
import { ConsoleLoggerService } from '../../../v1/core/logs/console';
import { DeploymentsController } from './controller/deployment.controller';
import { NotificationsController } from './controller/notification.controller';
import { ComponentEntityV2 as ComponentEntity } from './entity/component.entity';
import { DeploymentEntityV2 as DeploymentEntity } from './entity/deployment.entity';
import { Execution } from './entity/execution.entity';
import { PgBossWorker } from './jobs/pgboss.worker';
import { DeploymentHandler } from './use-cases/deployment-handler';
import { DeploymentUseCase } from './use-cases/deployment-use-case';


@Module({
  imports: [
    TypeOrmModule.forFeature([
      DeploymentEntity,
      ComponentEntity,
      Execution,
      CdConfigurationEntity,
      CdConfigurationsRepository
    ])
  ],
  controllers: [
    DeploymentsController,
    NotificationsController
  ],
  providers: [
    PgBossWorker,
    DeploymentUseCase,
    DeploymentHandler,
    ConsoleLoggerService
  ],
  exports: [
  ]
})
export class DeploymentsModule { }