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

# Advanced Usage Examples #

## Testing Multiple Charts ##
```bash
# Test all charts in a monorepo structure
for chart in charts/*; do
    if [ -d "$chart" ]; then
        echo "Testing chart: $chart"
        docker run -ti --rm -v $(pwd):/apps helmunittest/helm-unittest:3.11.1-0.3.0 "$chart"
    fi
done
```

## CI/CD Integration Examples ##

### GitHub Actions ###
```yaml
name: Helm Chart Tests
on: [push, pull_request]

jobs:
  unittest:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Helm Unit Tests
        run: |
          docker run --rm -v $(pwd):/apps helmunittest/helm-unittest:3.11.1-0.3.0 \
            -o junit.xml \
            --strict \
            --file 'tests/*.yaml' \
            .
      - name: Publish Test Results
        uses: mikepenz/action-junit-report@v3
        if: always()
        with:
          report_paths: 'junit.xml'

```

### GitLab CI ###
```yaml
helm-unittest:
  image: helmunittest/helm-unittest:3.11.1-0.3.0
  script:
    - helm unittest -o junit.xml --strict .
  artifacts:
    reports:
      junit: junit.xml
  rules:
    - changes:
      - charts/**/*
```

## Best Practices ##

1. **Version Pinning**:
   - Always use specific version tags (e.g., `3.11.1-0.3.0`) instead of `latest`
   - Document the version combinations in your CI configuration

2. **Security**:
   - Mount volumes as read-only when possible: `-v $(pwd):/apps:ro`
   - Use non-root user when running tests
   - Scan the Docker images for vulnerabilities in your CI pipeline

3. **Test Organization**:
   - Keep tests in a `tests/` directory within each chart
   - Use meaningful test file names (e.g., `values-validation_test.yaml`)
   - Group related test cases using test suites

4. **Performance**:
   - Use `.helmignore` to exclude unnecessary files
   - Consider using Docker layer caching in CI
   - Run tests in parallel for multiple charts when possible

5. **Validation**:
   - Enable strict mode (`--strict`) to catch potential issues
   - Use consistent assertion messages
   - Include tests for both success and failure cases

## Troubleshooting ##

Common issues and solutions:

1. **Permission Issues**:
   ```bash
   # Fix permission issues with mounted volumes
   docker run -ti --rm -v $(pwd):/apps:Z helmunittest/helm-unittest:3.11.1-0.3.0 .
   ```

2. **Path Resolution**:
   - Ensure paths are relative to the mounted `/apps` directory
   - Use absolute paths when referring to files outside the chart

3. **Output Formatting**:
   - JUnit: `-o junit.xml`
   - NUnit: `-o nunit.xml`
   - XUnit: `-o xunit.xml`
   - Sonar: `-o sonar.xml`
