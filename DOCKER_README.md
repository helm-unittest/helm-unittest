# Kubernetes Helm Unittest plugin #

Auto trigger docker build for [kubernetes helm unittest plugin](https://github.com/helm-unittest/helm-unittest) when a new release is created. </br>
The build comes with latest 3 version. </br>
More information on how to use the helm unittest plugin see https://github.com/helm-unittest/helm-unittest/blob/main/DOCUMENT.md

# README #

The latest docker tag is the latest helm release (https://github.com/helm/helm/releases/latest) containing the latest helm unittest plugin (https://github.com/helm-unittest/helm-unittest/releases/latest)

Please be aware to use the latest tag, as it can change the helm client and the helm unittest plugin version. Tag with the right versions is the proper way, such as ``` helmunittest/helm-unittest:3.11.1-0.3.0 ```

## Github Repo ##

https://github.com/helm-unittest/helm-unittest/ </br>
*Location of the repo containing the plugin and scripts to generate the docker images*

## Docker image tags ##

https://hub.docker.com/r/helmunittest/helm-unittest/tags/ </br>
*Overview of the available version combinations*

# Usage #
``` 
# run help of latest helm with latest helm unittest plugin
docker run -ti --rm -v $(pwd):/apps helmunittest/helm-unittest .

# run help of specific helm version with specific helm unittest plugin version
docker run -ti --rm -v $(pwd):/apps helmunittest/helm-unittest:3.11.1-0.3.0 .

# run unittests of a helm 3 chart
# make sure to mount local folder to /apps in container
docker run -ti --rm -v $(pwd):/apps helmunittest/helm-unittest:3.11.1-0.3.0 .

# run unittests of a helm 3 chart with Junit output for CI validation
# make sure to mount local folder to /apps in container
# the test-output.xml will be available in the local folder.
docker run -ti --rm -v $(pwd):/apps helmunittest/helm-unittest:3.11.1-0.3.0 -o test-output.xml .
```
*More information on how to use the helm unittest plugin see https://github.com/helm-unittest/helm-unittest/blob/main/DOCUMENT.md*

# Who can benefit from these image(s) #

All people who are using automated builds or as part of CI (continuous integration), as the focus on these docker containers is more to unittesting the helm charts.
