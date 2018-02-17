# helm test

write the test:
```yaml
name: name
files:
tests:
- it: should ...
  values:
  set:
  asserts:
    - matchSnapshot:
      documentIndex:
      not: true
    - equal:
        path:
        value:
    - notEqual:
    - matchRegex:
        path:
        pattern:
    - notMatchRegex:
    - contains:
    - notContains:
    - isNull:
    - isNotNull:
    - empty:
    - isNotEmpty:
    - isKind:
    - isApiVersion:
    - hasDocuments:
```

## Assertion

Define assertions to validate the manifests rendered with values provided in the test jobs. The example below tests the resouces name with 2 `equal` assertion.

```yaml
suite: test deploy and service
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

- **not**: *bool*. Set to `true` to assert contrarily, default to `false`. The second assertion in the example above asserts that the service name is **NOT** *your-service*.

- **template**: *string*. The template file which render the manifest to be asserted, default to the first template file defined in `templates` of suite file. For example the first assertion above with no `template` specified asserts `deployment.yaml` by default. If no template file specified in neither suite and assertion, the assertion returns an error and fail the test.

- **documentIndex**: *int*. The index of rendered documents (devided by `---`) to be asserted, default to 0. Generally you can ignored this field if the template file render only one document.

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

Notice that there are some antonym assertions, the following two assertions actually have same effect :
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
