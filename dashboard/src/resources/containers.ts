import { isResourceType, Resource } from "./resources"

export interface ContainerProperties {
  application: string
  environment: string
  container: ContainerPropertiesContainer
  connections?: {[name: string]: ContainerConnection}
}

export interface ContainerConnection {
  source: string
}

export interface ContainerPropertiesContainer {
  ports?: {[name: string]: ContainerPort}
}

export interface ContainerPort {
  containerPort?: number
  provides?: string
}

export const isContainerType = (resource: Resource<any>): boolean => {
  return isResourceType(resource, 'Applications.Core/containers');
}