import { ModuleDeployment } from './module-deployment.entity'
import { BaseEntity, Column, Entity, OneToMany, PrimaryGeneratedColumn } from 'typeorm'
import { ReadDeploymentDto } from '../dto'
import { DeploymentModuleResponse } from '../interface'
import { CircleDeployment } from './circle-deployment.entity'

@Entity('deployments')
export class Deployment extends BaseEntity {

  @PrimaryGeneratedColumn('uuid')
  public id: string

  @OneToMany(
    type => ModuleDeployment,
    moduleDeployment => moduleDeployment.deployment,
    { cascade: true, eager: true }
  )
  public modules: ModuleDeployment[]

  @Column({ name: 'user_id' })
  public authorId: string

  @Column({ name: 'description'} )
  public description: string

  @Column({ name: 'callback_url'} )
  public callbackUrl: string

  @Column({ type: 'jsonb', name: 'circles'} )
  public circles: CircleDeployment[]

  constructor(
    modules: ModuleDeployment[],
    authorId: string,
    description: string,
    callbackUrl: string,
    circles: CircleDeployment[]
  ) {
    super()
    this.modules = modules
    this.authorId = authorId
    this.description = description
    this.callbackUrl = callbackUrl
    this.circles = circles
  }

  private getDeploymentModulesResponseArray(): DeploymentModuleResponse[] {
    return this.modules.map(module => ({
      id: module.id,
      moduleId: module.moduleId,
      buildImageTag: module.buildImageTag
    }))
  }

  public toReadDto(): ReadDeploymentDto {
    return new ReadDeploymentDto(
      this.id,
      this.getDeploymentModulesResponseArray(),
      this.authorId,
      this.description,
      this.circles
    )
  }
}
