# Design discussion on InboundRoute

This document goes over the current design of InboundRoute, what it looks like app model v3, and the future direction we should take.

## What is an InboundRoute/Gateway/Ingress?

- Ability to expose a component to the public internet
- Allows for routing, load balancing, ssl termination, etc.
- nginx, API Management, Envoy, etc.

## Today

An InboundRoute is a trait, looking like.

```bicep
  resource frontend 'Components' = {
    name: 'frontend'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'rynowak/frontend:0.5.0-dev'
        }
      }
      bindings: {
        web: {
          kind: 'http'
          targetPort: 80
        }
      }
      traits: [
        {
          kind: 'radius.dev/InboundRoute@v1alpha1'
          binding: 'web'
        }
      ]
    }
  }
```

By adding InboundRoute, we are specifying that the frontend should be publicly exposed. This currently only supports container components.

## Changes with AppModelv3

In app model v3, the inbound route is still currently a trait, which makes it ever more odd as the binding itself has moved out of the component.

```bicep
resource http 'radius.dev/Application/HttpRoute@v1alpha3a' = {
  name: 'catalog-http'
  parent: app
  properties: {
    port: 5101
  }
}

resource catalog 'radius.dev/Application/ContainerComponent@v1alpha3a' = {
  name: 'catalog-api'
  parent: app
  properties: {
    run: {
      container: {
        image: 'eshop/catalog.api:latest'
        ports: {
          http: {
            provides: http.id
            containerPort: 80
          }
        }
      }
    }
    traits: [
      {
        kind: 'radius.dev/InboundRoute@v1alpha1'
        binding: http.id
      }
    ]
  }
}
```

## Current design issues
- Inbound route lacks a lot of fundamental requirements for ingress/apigateways/reverse-proxies that an application developer will care about, including
  - Path (if I have multiple components I want to expose publicly, how does traffic get to each of them)
  - Prefix match (very common to match just the start of a URI path)
  - Remove matched prefix
  - Hostname
- InboundRoutes are part of the component, which makes it very difficult to rationalize what is publicly exposed or not
  - Also, with routes/bindings being moved out as well, becomes increasingly confusing
  
## Options
### InboundRoute++

This simply adds more to inbound route.
```bicep
resource http 'radius.dev/Application/HttpRoute@v1alpha3a' = {
  name: 'catalog-http'
  parent: app
  properties: {
    port: 5101
  }
}
resource catalog 'radius.dev/Application/ContainerComponent@v1alpha3a' = {
  name: 'catalog-api'
  parent: app
  properties: {
    run: {
      container: {
        image: 'eshop/catalog.api:latest'
        ports: {
          http: {
            provides: http.id
            containerPort: 80
          }
        }
      }
    }
    traits: [
      {
        kind: 'radius.dev/InboundRoute@v1alpha1'
        binding: 'http'
        path: '/foo'
        prefixMatch: true // or exactMatch: true
        removePrefixOnMatch: true // word smithing here
        // more options here as well
      }
    ]
  }
}
```

Benefits:
- clearly associates a component that should be publicly exposed with a binding

Downsides:
- What if people wanted to configure the reverse-proxy/gateway being used?
- Still isn't a great fit to be part of a component as bindings have moved out of a component definition.

### Gateway resource contains routes

```bicep
resource gateway 'radius.dev/Application/Gateway@v1alpha3a' {
  name: 'gateway'
  // Additional gateways specific config here.
  routes: [
    {
      // TODO do we need to reassociate the binding with the service here?
      path: '/foo'
      prefixMatch: true
      binding: http.id
    }
    ...
  ]
}
resource http 'radius.dev/Application/HttpRoute@v1alpha3a' = {
  name: 'catalog-http'
  parent: app
  properties: {
    port: 5101
  }
}
resource catalog 'radius.dev/Application/ContainerComponent@v1alpha3a' = {
  name: 'catalog-api'
  parent: app
  properties: {
    run: {
      container: {
        image: 'eshop/catalog.api:latest'
        ports: {
          http: {
            provides: http.id
            containerPort: 80
          }
        }
      }
    }
  }
}
```

Benefits:
- Very clear what will be publicly exposed

Downsides:
- I don't think this will scale well in multifile and multi bindings scenarios. Will be hard to keep track of everything as a developer.

### Can we add the "inbound" aspect to the route directly

Implicit gateway

```bicep
resource http 'radius.dev/Application/HttpRoute@v1alpha3a' = {
  name: 'catalog-http'
  parent: app
  properties: {
    port: 5101
    gateway: {
      prefix: '/foo'
      prefixMatch: true
    }
  }
}
resource catalog 'radius.dev/Application/ContainerComponent@v1alpha3a' = {
  name: 'catalog-api'
  parent: app
  properties: {
    run: {
      container: {
        image: 'eshop/catalog.api:latest'
        ports: {
          http: {
            provides: http.id
            containerPort: 80
          }
        }
      }
    }
  }
}
```

Explicit gateway

```bicep
resource gateway 'radius.dev/Application/Gateway@v1alpha3a' = {
  name: 'gateway'
  // ...
}

resource http 'radius.dev/Application/HttpRoute@v1alpha3a' = {
  name: 'catalog-http'
  parent: app
  properties: {
    port: 5101
    gateway: {
      prefix: '/foo'
      id: gateway.id // list (multiple parents)
    }
  }
}
resource catalog 'radius.dev/Application/ContainerComponent@v1alpha3a' = {
  name: 'catalog-api'
  parent: app
  properties: {
    run: {
      container: {
        image: 'eshop/catalog.api:latest'
        ports: {
          http: {
            provides: http.id
            containerPort: 80
          }
        }
      }
    }
  }
}
```

Benefits:
- Feels simple, associates the binding with an inbound aspect of it.

Open question: should gateway information be part of HttpRoute or should we create a new resource type called InboundHttpRoute which requires explicit info about the path, gateway, etc.

Open question: One to many and many to one relationships between routes and gateways?

## What goes in the gateway resource

- IPs, Hostname, port, protocol
- Type of reverse proxy (nginx, azure, etc)
- Platform specific things (or maybe more resource types).

```bicep
resource gateway 'radius.dev/Application/Gateway@v1alpha3a' = {
  name: 'gateway'
  endpoints/listeners/input/idk: [
    {
      hostname: foo.com
      port: 80
      protocol: HTTP
    },
    {
      hostname: bar.com
      port: 443
      protocol HTTP
      tls: {
        // ...
      }
    },
    {
      ipaddress: 9.9.9.9
      port: 443
      protocol HTTP
      tls: {
        // ...
      }
    }
  ]
}
```

## Kubernetes is going through an ingress renaissance

https://gateway-api.sigs.k8s.io/

The concept of routes is very similar to what we have already design in radius. Overall the design seems pretty solid.