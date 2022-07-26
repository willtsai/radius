import radius as radius

@description('Specifies the location for resources.')
param location string = 'global'

@description('Specifies the environment for resources.')
param environment string

resource app 'Applications.Core/applications@2022-03-15-privatepreview' = {
  name: 'corerp-resources-container-traffictarget'
  location: location
  properties: {
    environment: environment
  }
}

resource backend 'Applications.Core/containers@2022-03-15-privatepreview' = {
  name: 'backend'
  location: location
  properties: {
    application: app.id
    container: {
      image: 'jkotalik.azurecr.io/backend:latest'
      ports: {
        web: {
          containerPort: 80
          provides: backendhttp.id
        }
      }
      volumes:{
        'my-volume':{
          kind: 'ephemeral'
          mountPath:'/tmpfs'
          managedStore:'memory'
        }
      }
    }
    connections: {}
  }
}

resource backendhttp 'Applications.Core/httpRoutes@2022-03-15-privatepreview' = {
  name: 'backend'
  location: location
  properties: {
    application: app.id
  }
}

resource frontend 'Applications.Core/containers@2022-03-15-privatepreview' = {
  name: 'frontend'
  location: location
  properties: {
    application: app.id
    container: {
      image: 'jkotalik.azurecr.io/frontend:latest'
      ports: {
        web: {
          containerPort: 80
          provides: frontendhttp.id
        }
      }
      env: {
        // Always set these environment variables to show
        // the frontend failing to connect when a "connection" isn't present
        CONNECTION__BACKEND__HOSTNAME: backendhttp.properties['hostname']
        CONNECTION__BACKEND__PORT: '${backendhttp.properties.port}'
      }
    }
    // Uncomment me to allow connection between frontend and backend
    connections: {
      backend: {
        source: backendhttp.id
      }
    }
  }
}

resource frontendhttp 'Applications.Core/httpRoutes@2022-03-15-privatepreview' = {
  name: 'frontend'
  location: location
  properties: {
    application: app.id
  }
}
