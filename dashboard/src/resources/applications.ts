import { isResourceType, Resource } from "./resources"

export interface ApplicationProperties {
  environment: string
}

export const isApplicationType = (resource: Resource<any>): boolean => {
    return isResourceType(resource, 'Applications.Core/applications');
  }