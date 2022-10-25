import { isResourceType, Resource } from "./resources"

export interface EnvironmentProperties {
}

export const isEnvironmentType = (resource: Resource<any>): boolean => {
  return isResourceType(resource, 'Applications.Core/environments');
}