resource eshop 'radius.dev/Applications@v1alpha1' = {
  name: 'eShop'

  // APIs -----------------------------------------------

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/identity-api
  resource identity 'Components' = {
    name: 'identity-api'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/identity.api:latest'
          env: {
            'PATH_BASE': ''
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'OrchestratorType': 'K8S'
            'IsClusterEnv': 'True'
            'ConnectionString': ''
            'DPConnectionString': ''
            'ApplicationInsights__InstrumentationKey': ''
            'MvcClient': ''
            'SpaClient': ''
            'BasketApiClient': ''
            'OrderingApiClient': ''
            'MobileShoppingAggClient': ''
            'WebShoppingAggClient': ''
            'XamarinCallback': ''
            'WebhooksApiClient': ''
            'WebhooksWebClient': ''
            'EnableDevspaces': 'False'
          }
        }
      }
      bindings: {
        http: {
          kind: 'http'
          targetPort: 80
        }
      }
      traits: [
        {
          kind: 'radius.dev/InboundRoute@v1alpha1'
          binding: 'http'
        }
      ]
      uses:[
        {
          binding: sqldb.properties.bindings.sql
        }
        {
          binding: servicebus.properties.bindings.default
        }
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/catalog-api
  resource catalog 'Components' = {
    name: 'catalog-api'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/catalog.api:latest'
          env: {
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'OrchestratorType': 'K8S'
            'PORT': '80'
            'GRPC_PORT': '81'
            'ConnectionString': ''
            'PicBaseUrl': ''
            'AzureStorageEnabled': 'False'
            'ApplicationInsights__InstrumentationKey': ''
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
        http: {
          kind: 'http'
          targetPort: 80
        }
        grpc: {
          kind: 'http'
          targetPort: 81
        }
      }
      traits: [
        {
          kind: 'radius.dev/InboundRoute@v1alpha1'
          binding: 'http'
        }
      ]
      uses:[
        {
          binding: sqldb.properties.bindings.sql
        }
        {
          binding: servicebus.properties.bindings.default
          env:{
            'EventBusConnection': servicebus.properties.bindings.default.connectionString
          }
        }
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/ordering-api
  resource ordering 'Components' = {
    name: 'ordering-api'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/ordering.api:latest'
          env: {
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses:[
        {
          binding: sqldb.properties.bindings.sql
        }
        {
          binding: servicebus.properties.bindings.default
          env:{
            'EventBusConnection': servicebus.properties.bindings.default.connectionString
          }
        }
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/basket-api
  resource basket 'Components' = {
    name: 'basket-api'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/basket.api:latest'
          env: {
            'OrchestratorType': 'K8S'
            'PORT': '80'
            'GRPC_PORT': '81'
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
        {
          binding: redis.properties.bindings.default
        }
        {
          binding: identity.properties.bindings.web
        }
        {
          binding: servicebus.properties.bindings.default
          env: {
            'EventBusConnection': servicebus.properties.bindings.default.connectionString
          }
        }
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/webhooks-api
  resource webhooks 'Components' = {
    name: 'webhooks-api'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/webhooks.api:latest'
          env: {
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'OrchestratorType': 'K8S'
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
        {
          binding: sqldb.properties.bindings.sql
        }
        {
          binding: servicebus.properties.bindings.default
          env:{
            'EventBusConnection': servicebus.properties.bindings.default.connectionString
          }
        }
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/payment-api
  resource payment 'Components' = {
    name: 'payment-api'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/payment.api:latest'
          env: {
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'OrchestratorType': 'K8S'
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
        {
          binding: servicebus.properties.bindings.default
          env: {
            'EventBusConnection': servicebus.properties.bindings.default.connectionString
          }
        }
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/ordering-backgroundtasks
  resource orderbgtasks 'Components' = {
    name: 'ordering-backgroundtasks'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/ordering.backgroundtasks:latest'
          env: {
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'OrchestratorType': 'K8S'
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
        {
          binding: sqldb.properties.bindings.sql
        }
        {
          binding: servicebus.properties.bindings.default
          env: {
             'EventBusConnection': servicebus.properties.bindings.default.connectionString
          }
        }
      ]
    }
  }

  // Other ---------------------------------------------

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/webshoppingagg
  resource webshoppingagg 'Components' = {
    name: 'webshoppingagg'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/webshoppingagg:latest'
          env: {
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'OrchestratorType': 'K8S'
            'IsClusterEnv': 'True'
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
        {
          binding: sqldb.properties.bindings.sql
        }
        {
          binding: mongodb.properties.bindings.mongo
        }
        {
          binding: servicebus.properties.bindings.default
        }
        {
          binding: identity.properties.bindings.http
        }
        {
          binding: ordering.properties.bindings.http
        }
        {
          binding: catalog.properties.bindings.http
        }
        {
          binding: basket.properties.bindings.http
        }
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/apigwws
  resource webshoppingapigw 'Components' = {
    name: 'webshoppingapigw'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'envoyproxy/envoy:v1.11.1'
          env: {
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/ordering-signalrhub
  resource orderingsignalrhub 'Components' = {
    name: 'ordering-signalrhub'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/ordering.signalrhub:latest'
          env: {
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'OrchestratorType': 'K8S'
            'IsClusterEnv': 'True'
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
        {
          binding: sqldb.properties.bindings.sql
        }
        {
          binding: mongodb.properties.bindings.mongo
        }
        {
          binding: servicebus.properties.bindings.default
          env: {
            'EventBusConnection': servicebus.properties.bindings.default.connectionString
          }
        }
        {
          binding: identity.properties.bindings.http
        }
        {
          binding: ordering.properties.bindings.http
        }
        {
          binding: catalog.properties.bindings.http
        }
        {
          binding: basket.properties.bindings.http
        }
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/webhooks-web
  resource webhooksclient 'Components' = {
    name: 'webhooks-client'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/webhooks.client:latest'
          env: {
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
        {
          binding: webhooks.properties.bindings.http
        }
      ]
    }
  }

  // Sites ----------------------------------------------

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/webstatus
  resource webstatus 'Components' = {
    name: 'webstatus'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/webstatus:latest'
          env: {
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'OrchestratorType': 'K8S'
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/webspa
  resource webspa 'Components' = {
    name: 'web-spa'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/webspa:latest'
          env: {
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'OrchestratorType': 'K8S'
            'IsClusterEnv': 'True'
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
        {
          binding: webshoppingagg.properties.bindings.http
        }
        {
          binding: webshoppingapigw.properties.bindings.http
        }
      ]
    }
  }

  // Based on https://github.com/dotnet-architecture/eShopOnContainers/tree/dev/deploy/k8s/helm/webmvc
  resource webmvc 'Components' = {
    name: 'webmvc'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'eshop/webmvc:latest'
          env: {
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'OrchestratorType': 'K8S'
            'IsClusterEnv': 'True'
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
        {
          binding: webshoppingagg.properties.bindings.http
        }
        {
          binding: webshoppingapigw.properties.bindings.http
        }
      ]
    }
  }

  // Logging --------------------------------------------

  resource seq 'Components' = {
    name: 'seq'
    kind: 'radius.dev/Container@v1alpha1'
    properties: {
      run: {
        container: {
          image: 'datalust/seq:latest'
          env: {
          }
        }
      }
      bindings: {
      }
      traits: [
      ]
      uses: [
      ]
    }
  }

  // Resources ------------------------------------------

  resource servicebus 'Components' = {
    name: 'servicebus'
    kind: 'azure.com/ServiceBusQueue@v1alpha1'
    properties: {
      config: {
        managed: true
        queue: '' 
      }
    }
  }

  resource sqldb 'Components' = {
    name: 'sqldb'
    kind: 'azure.com/CosmosDBSQL@v1alpha1'
    properties: {
      config: {
        managed: true
      }
    }
  }
  
  resource redis 'Components' = {
    name: 'redis'
    kind: 'redislabs.com/Redis@v1alpha1'
    properties: {
      config: {
        managed: true
      }
    }
  }

  resource mongodb 'Components' = {
    name: 'mongodb'
    kind: 'mongodb.com/Mongo@v1alpha1'
    properties: {
      config: {
        managed: true
      }
    }
  }

}
