## Python

These settings apply only when `--python` is specified on the command line.
Please also specify `--python-sdks-folder=<path to the root directory of your azure-sdk-for-python clone>`.

``` yaml $(track2)
azure-arm: true
license-header: MICROSOFT_MIT_NO_VERSION
package-name: azure-mgmt-applications
no-namespace-folders: true
package-version: 1.0.0b1
clear-output-folder: true
```

``` yaml $(python-mode) == 'update' && $(track2)
no-namespace-folders: true
output-folder: $(python-sdks-folder)/applications/azure-mgmt-applications/azure/mgmt/applications
```

``` yaml $(python-mode) == 'create' && $(track2)
basic-setup-py: true
output-folder: $(python-sdks-folder)/applications/azure-mgmt-applications
```

### Tag: package-core-2022-03-15-privatepreview and python

These settings apply only when `--tag=package-core-2022-03-15-privatepreview --python` is specified on the command line.
Please also specify `--python-sdks-folder=<path to the root directory of your azure-sdk-for-python clone>`.

``` yaml $(tag) == 'package-core-2022-03-15-privatepreview'
namespace: azure.mgmt.applications.core.v2022_03_15_privatepreview
output-folder: $(python-sdks-folder)/applications/azure-mgmt-applications/azure/mgmt/applications/core/v2022_03_15_privatepreview
python:
  namespace: azure.mgmt.applications.core.v2022_03_15_privatepreview
  output-folder: $(python-sdks-folder)/applications/azure-mgmt-applications/azure/mgmt/applications/core/v2022_03_15_privatepreview
```

### Tag: package-link-2022-03-15-privatepreview and python

These settings apply only when `--tag=package-link-2022-03-15-privatepreview --python` is specified on the command line.
Please also specify `--python-sdks-folder=<path to the root directory of your azure-sdk-for-python clone>`.

``` yaml $(tag) == 'package-link-2022-03-15-privatepreview'
namespace: azure.mgmt.applications.link.v2022_03_15_privatepreview
output-folder: $(python-sdks-folder)/applications/azure-mgmt-applications/azure/mgmt/applications/link/v2022_03_15_privatepreview
python:
  namespace: azure.mgmt.applications.link.v2022_03_15_privatepreview
  output-folder: $(python-sdks-folder)/applications/azure-mgmt-applications/azure/mgmt/applications/link/v2022_03_15_privatepreview
```

### Python multi-api

Generate all API versions currently shipped for this package

```yaml $(multiapi) && $(track2)
clear-output-folder: true
batch:
  - tag: package-core-2022-03-15-privatepreview
  - tag: package-link-2022-03-15-privatepreview
  - multiapiscript: true
```

``` yaml $(multiapiscript)
output-folder: $(python-sdks-folder)/applications/azure-mgmt-applications/azure/mgmt/applications/
clear-output-folder: false
perform-load: false
```