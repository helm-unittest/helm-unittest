# Frequently Asked Questions

## Installation

Q: **Is is possible to completely remove the plugin installation?** <br/>
A: Yes. Using the below command you can uninstall the plugin from usage
```
$ helm plugin uninstall helm-unittest
```
After the plugin removal, make sure the plugin cache is also cleaned (or at least the folder which contains the unittest plugin). See https://helm.sh/docs/faq/#xdg-base-directory-support for more information to identify the cache location for Helm 2 and Helm 3.

