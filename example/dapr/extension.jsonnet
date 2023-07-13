function(base, extension) base {
  spec+: {
    template+: {
      metadata+: {
        annotations+: {
          'dapr.io/app-id': if std.objectHas(extension, 'appId') then extension.appId else base.metadata.name,
          'dapr.io/app-port': if std.objectHas(extension, 'appPort') then extension.appPort,
          'dapr.io/config': if std.objectHas(extension, 'config') then extension.config,
          'dapr.io/enabled': 'true',
        },
      },
    },
  },
}
