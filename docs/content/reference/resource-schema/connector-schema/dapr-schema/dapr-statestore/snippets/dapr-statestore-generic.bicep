import radius as radius

param environment string

resource app 'Applications.Core/applications@2022-03-15-privatepreview' = {
  name: 'dapr-statestore-generic'
  location: 'global'
  properties: {
    environment: environment
  }
}
//SAMPLE
resource statestore 'Applications.Connector/daprStateStores@2022-03-15-privatepreview' = {
  name: 'statestore'
  location: 'global'
  properties: {
    environment: environment
    application: app.id
    kind: 'generic'
    type: 'state.couchbase'
    metadata: {
      couchbaseURL: '***'
      username: 'admin'
      password: '***'
      bucketName: 'dapr'
    }
    version: 'v1'
  }
}
//SAMPLE
