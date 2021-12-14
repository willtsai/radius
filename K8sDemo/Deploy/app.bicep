/**
 * The DEV team uses this file to specify how their application is wired: i.e.
 * what container connects to what databases, what pubsub queue to use, and etc.
 * Ideally when we change the infrastructure we don't ever need to touch this
 * file.
 */
 resource app 'radius.dev/Application@v1alpha3' = {
  name: 'demo'

  resource redis 'redislabs.com.RedisComponent' existing = {
    name: 'redis'
  }

  resource http 'HttpRoute' = {
    name: 'http'
    properties: {
      port: 8000
      gateway: {
        hostname: '*'
      }
    }
  }
  resource container 'ContainerComponent' = {
    name: 'todo'
    properties: {
      container: {
        image: 'jkotalik/demo:latest'
        ports: {
          http: {
            provides: http.id
            containerPort: 8000
          }
        }
      }
      connections: {
        redis: {
          kind: 'redislabs.com/Redis'
          source: redis.id
        }
      }
    }
  }
}
