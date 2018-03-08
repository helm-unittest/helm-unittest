# helm unittest

Unit test for *helm chart* in YAML with ease to keep your chart functional and robust!

Feature:
  - write test file in pure YAML
  - render locally with no need of *tiller*
  - create **nothing** on your cluster
  - define values and release options
  - [snapshot testing](#snapshot-testing)

## Install

```
$ helm plugin install https://github.com/lrills/helm-unittest
```

It will install latest version of binary into helm plugin directory.

## Get Started

Add `tests` in `.helmignore` of your chart, and create the following test file at `$YOUR_CHART/tests/deployment_test.yaml`:

```yaml
suite: test deployment
templates:
  - deployment.yaml
tests:
  - it: should work
    asserts:
      - isKind:
          of: Deployment
      - matchRegex:
          path: metadata.name
          pattern: -my-chart$
      - equal:
          path: spec.template.spec.containers[0].image
          value: nginx:stable
```
and run:

```
$ helm unittest $YOUR_CHART
```

Now there is your first test! ;)  
Please read the brief [document](#testing-document) below to learn writing your own tests.

## Usage

```
$ helm unittest [flags] CHART [...]
```

This renders your charts locally (without tiller) and runs tests
defined in test suite files.

### Flags

```
--color              enforce printing colored output even stdout is not a tty. Set to false to disable color
-f, --file stringArray   glob paths of test files location, default to tests/*_test.yaml (default [tests/*_test.yaml])
-h, --help               help for unittest
-u, --update-snapshot    update the snapshot cached if needed, make sure you review the change before update
```

## Example

Check [`__fixtures__/basic/`](https://github.com/lrills/helm-unittest/tree/master/__fixtures__/basic) for some basic use cases of a simple chart.

# Testing Document

## Test Suite

A test suite is a collection of tests with the same purpose and scope defined in one single file. The test suite file is written in pure YAML, and default placed under the `tests/` directory of the chart with suffix `_test.yaml`. You can also have your own suite files arrangement with `-f, --file` option of cli set as the glob patterns of test suite files, like:

```bash
$ helm unittest -f 'my-tests/*.yaml' -f 'more-tests/*.yaml' my-chart
```

The root structure of a test suite is like below:

```yaml
suite: test deploy and service
templates:
  - deployment.yaml
  - service.yaml
tests:
  - it: should test something
    ...
```

- **suite**: *string, optional*. The suite name to show on test result output.

- **templates**: *array of string, recommended*. The template files scope to test in this suite, only the ones specified here is rendered during testing. If omitted, all template files are rendered. File suffixed with `.tpl` is added automatically, you don't need to add them again.

- **tests**: *array of test job, required*. Where you define your test jobs to run, check [Test Job](#test-job).

## Test Job

The test job is the base unit for rendering chart for testing. Your chart is **rendered each time a test job run**, and validated with assertions defined in the test. You can setup your values used to render the chart in the test job with external values files or directly in the test job definition. Below is a test job example with all of its options defined:

```yaml
...
tests:
  - it: should pass
    values:
      - ./values/staging.yaml
    set:
      image.pullPolicy: Always
      resources:
        limits:
          memory: 128Mi
    release:
      name: my-release
      namespace:
      revision: 9
      isUpgrade: true
    asserts:
      - equal:
          path: metadata.name
          value: my-deploy
```

- **it**: *string, recommended*. Define the name of the test with TDD style or any message you like.

- **values**: *array of string, optional*. The values files used to render the chart, think it as the `-f, --values` options of `helm install`. The file path should be the relative path from the test suite file itself.

- **set**: *object of any, optional*. Set the values directly in suite file. The key is the value path with the format just like `--set` option of `helm install`, for example `image.pullPolicy`. The value is anything you want to set to the path specified by the key, which can be even an array or an object.

- **release**: *object, optional*. Define the `{{ .Release }}` object.
  - **name**: *string, optional*. The release name, default to `"RELEASE-NAME"`.
  - **namespace**: *string, optional*. The namespace which release be installed to, default to `"NAMESPACE"`.
  - **revision**: *string, optional*. The revision of current build, default to `0`.
  - **isUpgrade**: *bool, optional*. Whether the build is an upgrade, default to `false`.

- **asserts**: *array of assertion, required*. The assertions to validate the rendered chart, check [Assertion](#assertion).

## Assertion

Define assertions in the test job to validate the manifests rendered with values provided. The example below tests the instances' name with 2 `equal` assertion.

```yaml
templates:
  - deployment.yaml
  - service.yaml
tests:
  - it: should pass
    asserts:
      - equal:
          path: metadata.name
          value: my-deploy
      - equal:
          path: metadata.name
          value: your-service
        not: true
        template: service.yaml
        documentIndex: 0
```

The assertion is defined with the assertion type as the key and its parameters as value, there can be only one assertion type key exists in assertion definition object. And there are three more options can be set at root of assertion definition:

- **not**: *bool, optional*. Set to `true` to assert contrarily, default to `false`. The second assertion in the example above asserts that the service name is **NOT** *your-service*.

- **template**: *string, optional*. The template file which render the manifest to be asserted, default to the first template file defined in `templates` of suite file. For example the first assertion above with no `template` specified asserts `deployment.yaml` by default. If no template file specified in neither suite and assertion, the assertion returns an error and fail the test.

- **documentIndex**: *int, optional*. The index of rendered documents (devided by `---`) to be asserted, default to 0. Generally you can ignored this field if the template file render only one document.

### Assertion Types

Available assertion types are listed below:

| Assertion Type | Parameters | Description | Example |
|----------------|------------|-------------|---------|
| `equal` | **path**: *string*. The `set` path to assert.<br/>**value**: *any*. The expected value. | Assert the value of specified **path** equal to the **value**. | <pre>equal:<br/>  path: metadata.name<br/>  value: my-deploy</pre> |
| `notEqual` | **path**: *string*. The `set` path to assert.<br/>**value**: *any*. The value expected not to be. | Assert the value of specified **path** NOT equal to the **value**. | <pre>notEqual:<br/>  path: metadata.name<br/>  value: my-deploy</pre> |
| `matchRegex` | **path**: *string*. The `set` path to assert, the value must be a *string*. <br/>**pattern**: *string*. The regex pattern to match (without quoting `/`). | Assert the value of specified **path** match **pattern**. | <pre>matchRegex:<br/>  path: metadata.name<br/>  value: -my-chart$</pre> |
| `notMatchRegex` | **path**: *string*. The `set` path to assert, the value must be a *string*. <br/>**pattern**: *string*. The regex pattern NOT to match (without quoting `/`). | Assert the value of specified **path** NOT match **pattern**. | <pre>notMatchRegex:<br/>  path: metadata.name<br/>  pattern: -my-chat$</pre> |
| `contains` | **path**: *string*. The `set` path to assert, the value must be an *array*. <br/>**content**: *any*. The content to be contained. | Assert the array as the value of specified **path** contains the **content**. |<pre>contains:<br/>  path: spec.ports<br/>  content:<br/>    name: web<br/>    port: 80<br/>    targetPort: 80<br/>    protocle:TCP</pre> |
| `notContains` | **path**: *string*. The `set` path to assert, the value must be an *array*. <br/>**content**: *any*. The content NOT to be contained. | Assert the array as the value of specified **path** NOT contains the **content**. |<pre>notContains:<br/>  path: spec.ports<br/>  content:<br/>    name: server<br/>    port: 80<br/>    targetPort: 80<br/>    protocle: TCP</pre> |
| `isNull` | **path**: *string*. The `set` path to assert. | Assert the value of specified **path** is `null`. |<pre>isNull:<br/>  path: spec.strategy</pre> |
| `isNotNull` | **path**: *string*. The `set` path to assert. | Assert the value of specified **path** is NOT `null`. |<pre>isNotNull:<br/>  path: spec.replicas</pre> |
| `isEmpty` | **path**: *string*. The `set` path to assert. | Assert the value of specified **path** is empty (`null`, `""`, `0`, `[]`, `{}`). |<pre>isEmpty:<br/>  path: spec.tls</pre> |
| `isNotEmpty` | **path**: *string*. The `set` path to assert. | Assert the value of specified **path** is NOT empty (`null`, `""`, `0`, `[]`, `{}`). |<pre>isNotEmpty:<br/>  path: spec.selector</pre> |
| `isKind` | **of**: *String*. Expected `kind` of manifest. | Assert the `kind` value **of** manifest, is equilevant to:<br/><pre>equal:<br/>  path: kind<br/>  value: ...<br/> | <pre>isKind:<br/>  of: Deployment</pre> |
| `isApiVersion` | **of**: *string*. Expected `apiVersion` of manifest. | Assert the `apiVersion` value **of** manifest, is equilevant to:<br/><pre>equal:<br/>  path: apiVersion<br/>  value: ...<br/> | <pre>isApiVersion:<br/>  of: v2</pre> |
| `hasDocuments` | **count**: *int*. Expected count of documents rendered. | Assert the documents count rendered by the `template` specified. The `documentIndex` option is ignored here. | <pre>hasDocuments:<br/>  count: 2</pre> |
| `matchSnapshot` | **path**: *string*. The `set` path for snapshot. | Assert the value of **path** is the same as snapshotted last time. Check [doc](#snapshot-testing) below. | <pre>matchSnapshot:<br/>  path: spec</pre> |

### Antonym and `not`

Notice that there are some antonym assertions, the following two assertions actually have same effect:
```yaml
- equal:
    path: kind
    value: Pod
  not: true
# works the same as
- notEqual:
    path: kind
    value: Pod
```

### Snapshot Testing

Sometimes you just want to keep the rendered manifest not changed between editings without every details asserted. That's the reason for snapshot testing! Check the tests below:

```yaml
templates:
  - deployment.yaml
tests:
  - it: pod spec should match snapshot
    asserts:
      - matchSnapshot:
          path: spec.template.spec
  # or you can snapshot the whole manifest
  - it: manifest should match snapshot
    asserts:
      - matchSnapshot: {}
```

The `matchSnapshot` assertion validate the content rendered the same as cached last time. It fails if the content changed, and you should check and update the cache with `-u, --update-snapshot` option of cli.

```
$ helm unittest -u my-chart
```
The cache files is stored as `__snapshot__/*_test.yaml.snap` at the directory your test file placed, you should add them in version control with your chart.

## Related Projects / Commands

This plugin is inspired by [helm-template](https://github.com/technosophos/helm-template), and the idea of snapshot testing and some printing format comes from [jest](https://github.com/facebook/jest).


And there are some other helm commands you might want to use:

- [`helm template`](https://github.com/kubernetes/helm/blob/master/docs/helm/helm_template.md): render the chart and print the output.

- [`helm lint`](https://github.com/kubernetes/helm/blob/master/docs/helm/helm_lint.md): examines a chart for possible issues, useful to validate chart dependencies.

- [`helm test`](https://github.com/kubernetes/helm/blob/master/docs/helm/helm_test.md): test a release with testing pod defined in chart. Note this does create resources on your cluster to verify if your release is correct. Check the [doc](https://github.com/kubernetes/helm/blob/master/docs/chart_tests.md).

## License

MIT

## Contributing

Issues and PRs are welcome!  
Before start developing this plugin, you must have [go](https://golang.org/doc/install) and [dep](https://github.com/golang/dep#installation) installed, and run:

```
git clone git@github.com:lrills/helm-unittest.git
cd helm-unittest
dep ensure
```

And please make CI passed when request a PR which would check following things:

- `dep status` passed. Make sure you run `dep ensure` if new dependencies added.
- `gofmt` no changes needed. Please run `gofmt -w -s` before you commit.
- `go test ./unittest/...` passed.

In some cases you might need to manually fix the tests in `*_test.go`. If the snapshot tests (of the plugin's test code) failed you need to run:

```
UPDATE_SNAPSHOTS=true go test ./unittest/...
```

This update the snapshot cache file and please add them before you commit.
