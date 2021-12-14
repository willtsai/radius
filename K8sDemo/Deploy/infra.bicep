/**
 * This file contains infrastructure specification: i.e. how databases are
 * setup, how are they load-balanced, and where are the passwords stored. It is
 * owned by the OPS team.
 */
 import kubernetes from kubernetes

 resource redisService 'kubernetes.core/Service@v1' existing = {
   metadata: {
     name: 'redis-master'
   }
 }
 
 resource redisSecret 'kubernetes.core/Secret@v1' existing = {
   metadata: {
     name: 'redis'
   }
 }
 
 resource app 'radius.dev/Application@v1alpha3' = {
   name: 'demo'
 
   resource redis 'redislabs.com.RedisComponent' = {
     name: 'redis'
     properties: {
       // Now we provide data from the OPS team's Kubernetes resources to Radius.
       host: '${redisService.metadata.name}.${redisService.metadata.namespace}.svc.cluster.local'
       port: redisService.spec.ports[0].port
       secrets: {
         password: base64ToString(redisSecret.data['redis-password'])
         connectionString: '${redisService.metadata.name}.${redisService.metadata.namespace}.svc.cluster.local:${redisService.spec.ports[0].port},password=${base64ToString(redisSecret.data['redis-password'])}'
       }
     }
   }
 }

