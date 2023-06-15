import radius as radius

param environment string

resource app 'Applications.Core/applications@2022-03-15-privatepreview' = {
  name: 'my-app'
  properties: {
    environment: environment
  }
}

resource container 'Applications.Core/containers@2022-03-15-privatepreview' = {
  name: 'my-backend'
  properties: {
    application: app.id
    container: {
      image: 'myimage'
    }
    connections: {
      blob: {
        source: blobContainer.id
        iam: {
          kind: 'azure'
          roles: [
            'Storage Blob Data Reader'
          ]
        }
      }
    }
  }
}

resource blobContainer 'Microsoft.Storage/storageAccounts/blobServices/containers@2021-06-01' existing = {
  name: 'mystorage/default/mycontainer'
}