# Kubernetes Helm Unittest plugin #

Auto trigger docker build for [kubernetes helm unittest plugin](https://github.com/quintush/helm-unittest) when a new release is created. </br>
The build comes with several helm 2 & helm 3 version, as the unittest plugin supports both helm versions. </br>
More information on how to use the helm unittest plugin see https://github.com/quintush/helm-unittest/blob/master/DOCUMENT.md

# README #

The latest docker tag is the latest helm release (https://github.com/helm/helm/releases/latest) containing the latest helm unittest plugin (https://github.com/quintush/helm-unittest/releases/latest)

Please be aware to use the latest tag, as it can change the helm client and the helm unittest plugin version. Tag with the right versions is the proper way, such as ``` quintus/helm-unittest:3.3.0-0.2.2 ```

## Github Repo ##

https://github.com/quintush/helm-unittest/ </br>
*Location of the repo containing the plugin and scripts to generate the docker images*

## Docker image tags ##

https://hub.docker.com/r/quintush/helm-unittest/tags/ </br>
*Overview of the available version combinations*

# Usage #
``` 
# run help of latest helm with latest helm unittest plugin
docker run -ti --rm -v $(pwd):/apps quintush/helm-unittest

# run help of specific helm version with specific helm unittest plugin version
docker run -ti --rm -v $(pwd):/apps quintush/helm-unittest:3.3.0-0.2.2

# run unittests of a helm 2 chart
# make sure to mount local folder to /apps in container
docker run -ti --rm -v $(pwd):/apps quintush/helm-unittest:2.16.10-0.2.2 .

# run unittests of a helm 3 chart
# make sure to mount local folder to /apps in container
docker run -ti --rm -v $(pwd):/apps quintush/helm-unittest:3.3.0-0.2.2 -3 .

# run unittests of a helm 3 chart with Junit output for CI validation
# make sure to mount local folder to /apps in container
# the test-output.xml will be available in the local folder.
docker run -ti --rm -v $(pwd):/apps quintush/helm-unittest:3.3.0-0.2.2 -3 -o test-output.xml .
```
*More information on how to use the helm unittest plugin see https://github.com/quintush/helm-unittest/blob/master/DOCUMENT.md*

# Who can benefit from these image(s) #

All people who are using automated builds or as part of CI (continuous integration), as the focus on these docker containers is more to unittesting the helm charts.
