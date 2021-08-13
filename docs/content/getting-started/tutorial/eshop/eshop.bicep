param ESHOP_EXTERNAL_DNS_NAME_OR_IP string = 'localhost'

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
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'OrchestratorType': 'K8S'
            'IsClusterEnv': 'True'
            'DPConnectionString': ''
            'ApplicationInsights__InstrumentationKey': ''
            'XamarinCallback': ''
            'EnableDevspaces': 'False'
          }
        }
      }
      bindings: {
        http: {
          kind: 'http'
          targetPort: 80
          port: 5105
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
          binding: sqldbi.properties.bindings.sql
          env:{
            'ConnectionString': sqldbi.properties.bindings.sql.connectionString
          }
        }
        {
          binding: webmvc.properties.bindings.http
          env: {
            'MvcClient': 'http://${webmvc.properties.bindings.http.host}:${webmvc.properties.bindings.http.port}'
          }
        }
        {
          binding: webspa.properties.bindings.http
          env: {
            'SpaClient': 'http://${webspa.properties.bindings.http.host}:${webspa.properties.bindings.http.port}'
          }
        }
        {
          binding: basket.properties.bindings.http
          env: {
            'BasketApiClient': 'http://${basket.properties.bindings.http.host}:${basket.properties.bindings.port}'
          }
        }
        {
          binding: ordering.properties.bindings.http
          env:{
            'OrderingApiClient': 'http://${ordering.properties.bindings.http.host}:${ordering.properties.bindings.http.port}'
          }
        }
        {
          binding: webshoppingagg.properties.bindings.http
          env:{
            'WebShoppingAggClient': 'http://${webshoppingagg.properties.bindings.http.host}:${webshoppingagg.properties.bindings.http.port}'
          }
        }
        {
          binding: webhooks.properties.bindings.http
          env: {
            'WebhooksApiClient': 'http://${webhooks.properties.bindings.http.host}:${webhooks.properties.bindings.http.port}'
          }
        }
        {
          binding: webhooksclient.properties.bindings.http
          env: {
            'WebhooksWebClient': 'http://${webhooksclient.properties.bindings.http.host}:${webhooksclient.properties.bindings.http.port}'
          }
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
            'UseCustomizationData': 'False'
            'PATH_BASE': '/catalog-api'
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'OrchestratorType': 'K8S'
            'PORT': '80'
            'GRPC_PORT': '81'
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
          port: 5101
        }
        grpc: {
          kind: 'http'
          targetPort: 81
          port: 9101
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
          binding: sqldbc.properties.bindings.sql
          env:{
            'ConnectionString': sqldbc.properties.bindings.sql.connectionString
          }
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
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'UseCustomizationData': 'False'
            'AzureServiceBusEnabled': 'True'
            'CheckUpdateTime': '30000'
            'ApplicationInsights__InstrumentationKey': ''
            'OrchestratorType': 'K8S'
            'UseLoadTest': 'False'
            'Serilog__MinimumLevel__Override__Microsoft.eShopOnContainers.BuildingBlocks.EventBusRabbitMQ': 'Verbose'
            'Serilog__MinimumLevel__Override__ordering-api': 'Verbose'
            'PATH_BASE': '/ordering-api'
            'GRPC_PORT': '81'
            'PORT': '80'
          }
        }
      }
      bindings: {
        http: {
          kind: 'http'
          targetPort: 80
          port: 5102
        }
        grpc: {
          kind:  'http'
          targetPort: 81
          port: 9102
        }
      }
      traits: [
      ]
      uses:[
        {
          binding: sqldbo.properties.bindings.sql
          env:{
            'ConnectionString': sqldbo.properties.bindings.sql.connectionString
          }
        }
        {
          binding: servicebus.properties.bindings.default
          env:{
            'EventBusConnection': servicebus.properties.bindings.default.connectionString
          }
        }
        {
          binding: identity.properties.bindings.http
          env: {
            'identityUrl': 'http://${identity.properties.bindings.http.host}'
            'IdentityUrlExternal': 'http://${identity.properties.bindings.http.host}:${identity.properties.bindings.http.port}'
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
            'ASPNETCORE_ENVIRONMENT': 'Development'
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'ApplicationInsights__InstrumentationKey': ''
            'UseLoadTest': 'False'
            'PATH_BASE': '/basket-api'
            'OrchestratorType': 'K8S'
            'PORT': '80'
            'GRPC_PORT': '81'
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
        http: {
          kind: 'http'
          targetPort: 80
          port: 5103
        }
        grpc: {
          kind: 'http'
          targetPort: 81
          port: 9103 
        }
      }
      traits: [
      ]
      uses: [
        {
          binding: redis.properties.bindings.default
          env:{
            'ConnectionString': redis.properties.bindings.default.connectionString
          }
        }
        {
          binding: identity.properties.bindings.http
        }
        {
          binding: servicebus.properties.bindings.default
          env: {
            'EventBusConnection': servicebus.properties.bindings.default.connectionString
          }
        }
        {
          binding: identity.properties.bindings.http
          env: {
            'identityUrl': 'http://${identity.properties.bindings.http.host}'
            'IdentityUrlExternal': 'http://${identity.properties.bindings.http.host}:${identity.properties.bindings.http.port}'
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
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'OrchestratorType': 'K8S'
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
        http:{
          kind: 'http'
          targetPort: 80
          port: 5113
        }
      }
      traits: [
      ]
      uses: [
        {
          binding: sqldbw.properties.bindings.sql
          env: {
            'ConnectionString': sqldbw.properties.bindings.sql.connectionString
          }
        }
        {
          binding: servicebus.properties.bindings.default
          env:{
            'EventBusConnection': servicebus.properties.bindings.default.connectionString
          }
        }
        {
          binding: identity.properties.bindings.http
          env: {
            'identityUrl': 'http://${identity.properties.bindings.http.host}'
            'IdentityUrlExternal': 'http://${identity.properties.bindings.http.host}:${identity.properties.bindings.http.port}'
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
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'ApplicationInsights__InstrumentationKey': ''
            'Serilog__MinimumLevel__Override__payment-api.IntegrationEvents.EventHandling': 'Verbose'
            'Serilog__MinimumLevel__Override__Microsoft.eShopOnContainers.BuildingBlocks.EventBusRabbitMQ': 'Verbose'
            'OrchestratorType': 'K8S'
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
        http:{
          kind: 'http'
          targetPort: 80
          port: 5108
        }
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
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'UseCustomizationData': 'False'
            'CheckUpdateTime': '30000'
            'GracePeriodTime': '1'
            'ApplicationInsights__InstrumentationKey': ''
            'UseLoadTest': 'False'
            'Serilog__MinimumLevel__Override__Microsoft.eShopOnContainers.BuildingBlocks.EventBusRabbitMQ': 'Verbose'
            'OrchestratorType': 'K8S'
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
        http:{
          kind: 'http'
          targetPort: 80
          port: 5111
        }
      }
      traits: [
      ]
      uses: [
        {
          binding: sqldbo.properties.bindings.sql
          env:{
            'ConnectionString': sqldbo.properties.bindings.sql.connectionString
          }
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
            'urls__basket': 'http://basket-api'
            'urls__catalog': 'http://catalog-api'
            'urls__orders': 'http://ordering-api'
            'urls__identity': 'http://identity-api'
            'urls__grpcBasket': 'http://basket-api:81'
            'urls__grpcCatalog': 'http://catalog-api:81'
            'urls__grpcOrdering': 'http://ordering-api:81'
            'CatalogUrlHC': 'http://catalog-api/hc'
            'OrderingUrlHC': 'http://ordering-api/hc'
            'IdentityUrlHC': 'http://identity-api/hc'
            'BasketUrlHC': 'http://basket-api/hc'
            'PaymentUrlHC': 'http://payment-api/hc'
            'IdentityUrlExternal': 'http://${ESHOP_EXTERNAL_DNS_NAME_OR_IP}:5105'
          }
        }
      }
      bindings: {
        http: {
          kind: 'http'
          targetPort: 80
          port: 5121
        }
      }
      traits: [
      ]
      uses: [
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
        http: {
          kind: 'http'
          targetPort: 80
          port: 5202
        }
        http2: {
          kind: 'http'
          targetPort: 8001
          port: 15202
        }
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
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'ApplicationInsights__InstrumentationKey': ''
            'OrchestratorType': 'K8S'
            'IsClusterEnv': 'True'
            'AzureServiceBusEnabled': 'True'
          }
        }
      }
      bindings: {
        http: {
          kind: 'http'
          targetPort: 80
          port: 5112
        }
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
        {
          binding: identity.properties.bindings.http
          env: {
            'identityUrl': 'http://${identity.properties.bindings.http.host}'
            'IdentityUrlExternal': 'http://${identity.properties.bindings.http.host}:${identity.properties.bindings.http.port}'
          }
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
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'Token': '6168DB8D-DC58-4094-AF24-483278923590' // Webhooks are registered with this token (any value is valid) but the client won't check it
            'CallBackUrl': 'http://${ESHOP_EXTERNAL_DNS_NAME_OR_IP}:5114'
            'SelfUrl': 'http://webhooks-client/'
          }
        }
      }
      bindings: {
        http: {
          kind: 'http'
          targetPort: 80
          port: 5114
        }
      }
      traits: [
      ]
      uses: [
        {
          binding: webhooks.properties.bindings.http
          env: {
            'WebhooksUrl': 'http://${webhooks.properties.bindings.http.host}'
          }
        }
        {
          binding: identity.properties.bindings.http
          env: {
            'IdentityUrl': 'http://${identity.properties.bindings.http.host}:${identity.properties.bindings.http.port}'
          }
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
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'HealthChecksUI__HealthChecks__0__Name': 'WebMVC HTTP Check'
            'HealthChecksUI__HealthChecks__0__Uri': 'http://webmvc/hc'
            'HealthChecksUI__HealthChecks__1__Name': 'WebSPA HTTP Check'
            'HealthChecksUI__HealthChecks__1__Uri': 'http://webspa/hc'
            'HealthChecksUI__HealthChecks__2__Name': 'Web Shopping Aggregator GW HTTP Check'
            'HealthChecksUI__HealthChecks__2__Uri': 'http://webshoppingagg/hc'
            'HealthChecksUI__HealthChecks__3__Name': 'Mobile Shopping Aggregator HTTP Check'
            'HealthChecksUI__HealthChecks__3__Uri': 'http://mobileshoppingagg/hc'
            'HealthChecksUI__HealthChecks__4__Name': 'Ordering HTTP Check'
            'HealthChecksUI__HealthChecks__4__Uri': 'http://ordering-api/hc'
            'HealthChecksUI__HealthChecks__5__Name': 'Basket HTTP Check'
            'HealthChecksUI__HealthChecks__5__Uri': 'http://basket-api/hc'
            'HealthChecksUI__HealthChecks__6__Name': 'Catalog HTTP Check'
            'HealthChecksUI__HealthChecks__6__Uri': 'http://catalog-api/hc'
            'HealthChecksUI__HealthChecks__7__Name': 'Identity HTTP Check'
            'HealthChecksUI__HealthChecks__7__Uri': 'http://identity-api/hc'
            'HealthChecksUI__HealthChecks__8__Name': 'Payments HTTP Check'
            'HealthChecksUI__HealthChecks__8__Uri': 'http://payment-api/hc'
            'HealthChecksUI__HealthChecks__9__Name': 'Ordering SignalRHub HTTP Check'
            'HealthChecksUI__HealthChecks__9__Uri': 'http://ordering-signalrhub/hc'
            'HealthChecksUI__HealthChecks__10__Name': 'Ordering HTTP Background Check'
            'HealthChecksUI__HealthChecks__10__Uri': 'http://ordering-backgroundtasks/hc'
            'ApplicationInsights__InstrumentationKey': ''
            'OrchestratorType': 'K8S'
          }
        }
      }
      bindings: {
        http:{
          kind: 'http'
          targetPort: 80
          port: 5107
        }
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
            'ASPNETCORE_ENVIRONMENT': 'Production'
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'UseCustomizationData': 'False'
            'ApplicationInsights__InstrumentationKey': ''
            'OrchestratorType': 'K8S'
            'IsClusterEnv': 'True'
          }
        }
      }
      bindings: {
        http:{
          kind: 'http'
          targetPort: 80
          port: 5104
        }
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
        {
          binding: identity.properties.bindings.http
          env:{
            'IdentityUrl': 'http://${identity.properties.bindings.http.host}:${identity.properties.bindings.http.port}'
            'IdentityUrlHC': 'http://${identity.properties.bindings.http.host}/hc'
          }
        }
        {
          binding: webshoppingapigw.properties.bindings.http
          env:{
            'PurchaseUrl': 'http://${webshoppingapigw.properties.bindings.http.host}:${webshoppingapigw.properties.bindings.http.port}'
          }
        }
        {
          binding: orderingsignalrhub.properties.bindings.http
          env: {
            'SignalrHubUrl': 'http://${orderingsignalrhub.properties.bindings.http.host}:${orderingsignalrhub.properties.bindings.http.port}'
          }
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
            'ASPNETCORE_URLS': 'http://0.0.0.0:80'
            'UseCustomizationData': 'False'
            'ApplicationInsights__InstrumentationKey': ''
            'UseLoadTest': 'False'
            'OrchestratorType': 'K8S'
            'IsClusterEnv': 'True'
          }
        }
      }
      bindings: {
        http:{
          kind: 'http'
          targetPort: 80
          port: 5100
        }
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
        {
          binding: identity.properties.bindings.http
          env:{
            'IdentityUrl': 'http://${identity.properties.bindings.http.host}:${identity.properties.bindings.http.port}'
            'IdentityUrlHC': 'http://${identity.properties.bindings.http.host}/hc'
          }
        }
        {
          binding: webshoppingapigw.properties.bindings.http
          env:{
            'PurchaseUrl': 'http://${webshoppingapigw.properties.bindings.http.host}'
          }
        }
        {
          binding: orderingsignalrhub.properties.bindings.http
          env:{
            'SignalrHubUrl': 'http://${orderingsignalrhub.properties.bindings.http.host}:${orderingsignalrhub.properties.bindings.http.port}'
          }
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
            'ACCEPT_EULA': 'Y'
          }
        }
      }
      bindings: {
        web:{
          kind: 'http'
          targetPort: 5340
          port: 80
        }
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

  resource sqldbi 'Components' = {
    name: 'sqldb-identity'
    kind: 'azure.com/CosmosDBSQL@v1alpha1'
    properties: {
      config: {
        managed: true
      }
    }
  }

  resource sqldbc 'Components' = {
    name: 'sqldb-catalog'
    kind: 'azure.com/CosmosDBSQL@v1alpha1'
    properties: {
      config: {
        managed: true
      }
    }
  }

  resource sqldbo 'Components' = {
    name: 'sqldb-ordering'
    kind: 'azure.com/CosmosDBSQL@v1alpha1'
    properties: {
      config: {
        managed: true
      }
    }
  }

  resource sqldbw 'Components' = {
    name: 'sqldb-webhooks'
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
