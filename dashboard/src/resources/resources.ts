
export interface Resource<TProperties> {
  id: string
  name: string
  type: string
  properties: TProperties
}

export const isResourceType = (resource: Resource<any>, type: string): boolean => {
  return resource.type.toLowerCase() === type.toLowerCase();
}

export const extractResourceGroup = (id: string): string | null => {
  const parts = id.split('/').filter(v => v.length > 0)
  for (let i = 0; i < parts.length; i++) {
    if (parts[i] === 'providers') {
      return parts[i - 1]
    }
  }

  return null
}

export const extractResourceName = (id: string | undefined): string | null => {
  if (!id) {
    return null;
  }

  const parts = id.split('/').filter(v => v.length > 0)
  for (let i = 0; i < parts.length; i++) {
    if (parts[i] === 'providers') {
      return parts[i + 3]
    }
  }

  return null
}