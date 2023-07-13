function(base, extension) base {
  spec+: {
    template+: {
      spec+: {
        topologySpreadConstraints: if std.objectHas(base.spec.template.spec, 'topologySpreadConstraints') then base.spec.template.spec.topologySpreadConstraints else if [
            {
              maxSkew: 1,
              topologyKey: 'kubernetes.io/hostname',
              matchLabelKeys: [ 'pod-template-hash' ],
            }
          ]
      },
    },
  },
}