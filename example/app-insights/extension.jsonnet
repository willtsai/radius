function(base, extension) base {
  assert !std.objectHas(extension, 'connectionString') : "connectionString is a required field",

  spec+: {
    template+: {
      spec+: {
        containers: [container {
          env: container.env + [{ name: 'APPLICATION_INSIGHTS_CONNECTIONSTRING', value: extension.connectionString }],
        } for container in base.spec.template.spec.containers]
      },
    },
  },
}
