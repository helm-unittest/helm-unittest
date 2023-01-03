# Frequently Asked Questions

## Installation

Q: **Is is possible to completely remove the plugin installation?** <br/>
A: Yes. Using the below command you can uninstall the plugin from usage
```
$ helm plugin uninstall helm-unittest
```
After the plugin removal, make sure the plugin cache is also cleaned (or at least the folder which contains the unittest plugin). See https://helm.sh/docs/faq/#xdg-base-directory-support for more information to identify the cache location for Helm 2 and Helm 3.

## Debugging
Q: **My test is failing but the expected and actual results are the same, what is happening?** <br/>
A: The error output is formatted for better readabillity. The result of the formatting is that it removes spaces and line endings, which can result in the same values between the expected and actual results.
With the debug option it is possible to see the the expected and actual content before the formatting is done.
```
$ helm helm-unittest ... -d
```

## DevOps
Q: **How can I setup the helm-unittest plugin in a build environment**
A: The helm-unittest plugin has the options _-t, --output-type_ and _-o, --output-file_ which can be use to generate testresults in a file. Most of the Buildservers have a task that can upload the testresult into the server and generate a buildreport, or determine the success or failure of the tests.
```
$ helm helm-unittest ... -t JUnit -o junit-results.xml
```
