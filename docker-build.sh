#!/usr/bin/env bash

# Prerequisite
# Make sure you set secret enviroment variables in Circle CI
# DOCKER_USERNAME
# DOCKER_PASSWORD
# GITHUB_TOKEN

# set -ex

build() {

  echo "Found new version, building the image ${image}:${tag}"
  docker build --no-cache --build-arg HELM_VERSION=${helmVersion} --build-arg PLUGIN_VERSION=${pluginVersion} -t ${image}:${tag} .

  # run test
  version=$(docker run -ti --entrypoint "helm" --rm ${image}:${tag} version --client)
  #Client: &version.Version{SemVer:"v2.9.0-rc2", GitCommit:"08db2d0181f4ce394513c32ba1aee7ffc6bc3326", GitTreeState:"clean"}
  if [[ "${version}" == *"Error: unknown flag: --client"* ]]; then
    echo "Detected Helm3+"
    version=$(docker run -ti --entrypoint "helm" --rm ${image}:${tag} version)
    #version.BuildInfo{Version:"v3.0.0-beta.2", GitCommit:"26c7338408f8db593f93cd7c963ad56f67f662d4", GitTreeState:"clean", GoVersion:"go1.12.9"}
  fi
  version=$(echo ${version}| awk -F \" '{print $2}')
  if [ "${version}" == "v${helmVersion}" ]; then
    echo "matched"
  else
    echo "unmatched"
    exit
  fi

  if [[ ! -z "$DOCKER_PASSWORD" ]] && [[ ! -z "$DOCKER_USERNAME" ]]; then
    echo "$DOCKER_PASSWORD" | docker login --username $DOCKER_USERNAME --password-stdin
    docker push ${image}:${tag}
  fi
}

image="quintush/helm-unittest"
helmRepo="helm/helm"
pluginRepo="quintush/helm-unittest"

if [[ ${CI} == 'true' ]]; then
  helmLatest=`curl -sL -H "Authorization: token ${GITHUB_TOKEN}"  https://api.github.com/repos/${helmRepo}/tags?per_page=50 |jq -r ".[].name"|sed 's/^v//'|sort -V |grep -v -`
  pluginLatest=`curl -sL -H "Authorization: token ${GITHUB_TOKEN}"  https://api.github.com/repos/${pluginRepo}/tags?per_page=2 |jq -r ".[].name"|sed 's/^v//'|sort -V |grep -v -`
else
  helmLatest=`curl -sL https://api.github.com/repos/${helmRepo}/tags?per_page=50 |jq -r ".[].name"| sed 's/^v//'| sort -V |grep -v -`
  pluginLatest=`curl -sL https://api.github.com/repos/${pluginRepo}/tags?per_page=2 |jq -r ".[].name"| sed 's/^v//'| sort -V |grep -v -`
fi

for helmVersion in ${helmLatest}
do
  echo "Found helm version: $helmVersion"
  for pluginVersion in ${pluginLatest}
  do
    echo "Found helm-unittest plugin version: $pluginVersion"
    tag="$helmVersion-$pluginVersion"
    echo $tag 
    status=$(curl -sL https://hub.docker.com/v2/repositories/${image}/tags/${tag})
    echo $status
    if [[ "${status}" =~ "not found" ]]; then
        build
    fi
  done
done

echo "Update latest image with latest release"
# output format for reference:
# <html><body>You are being <a href="https://github.com/helm/helm/releases/tag/v2.14.3">redirected</a>.</body></html>
helmLatestRelease=$(curl -s https://github.com/${helmRepo}/releases)
helmLatestVersion=$(echo $helmLatestRelease\" |grep -oP '(?<=tag\/v)[0-9][^"]*'|grep -v \-|sort -Vr|head -1)
pluginLatestRelease=$(curl -s https://github.com/${pluginRepo}/releases)
pluginLatestVersion=$(echo $pluginLatestRelease\" |grep -oP '(?<=tag\/v)[0-9][^"]*'|grep -v \-|sort -Vr|head -1)
latest="$helmLatestVersion-$pluginLatestVersion"
echo $latest

if [[ ! -z "$DOCKER_PASSWORD" ]] && [[ ! -z "$DOCKER_USERNAME" ]]; then
  echo "$DOCKER_PASSWORD" | docker login --username $DOCKER_USERNAME --password-stdin
  docker pull ${image}:${latest}
  docker tag ${image}:${latest} ${image}:latest
  docker push ${image}:latest
fi