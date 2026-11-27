# Frequently Asked Questions

## Installation

Q: **Is it possible to completely remove the plugin installation?** <br/>
A: Yes. Using the below command you can uninstall the plugin from usage
```bash
helm plugin uninstall helm-unittest
```
After the plugin removal, make sure the plugin cache is also cleaned (or at least the folder which contains the unittest plugin). See https://helm.sh/docs/faq/#xdg-base-directory-support for more information to identify the cache location for Helm 3.

## Debugging
Q: **My test is failing but the expected and actual results are the same, what is happening?** <br/>
A: The error output is formatted for better readability. The result of the formatting is that it removes spaces and line endings, which can result in the same values between the expected and actual results.
With the debug option it is possible to see the expected and actual content before the formatting is done.
```bash
helm helm-unittest ... -d
```

## DevOps
Q: **How can I setup the helm-unittest plugin in a build environment** <br/>
A: The helm-unittest plugin has the options _-t, --output-type_ and _-o, --output-file_ which can be use to generate testresults in a file. Most of the Buildservers have a task that can upload the testresult into the server and generate a buildreport, or determine the success or failure of the tests.
```bash
helm helm-unittest ... -t JUnit -o junit-results.xml
```

Q: **Can I integrate helm-unittest plugin with the Helm Chart-Testing tool?** <br/>
A: Yes. The Helm Chart-Testing tool as the option to add [additional-commands](https://github.com/helm/chart-testing/blob/main/doc/ct_lint.md#options) when linting is used. This makes it possible to run helm unittest as part of the ct command. <br/>
_Be aware! the required tools like helm and the unittest plugin need to be available on the system_

## Asserting on parsed JSON/YAML
Q: **Why does my test suite fail to load when I add `parse` or `innerPath` to an assertion?** <br/>
A: The `parse` and `innerPath` options are checked while the test suite file is loaded, before anything is rendered. A wrong value makes the whole suite file fail to parse, so none of its tests run. This prevents a typo from making an assertion pass without checking anything. The suite is rejected when:
- `parse` or `innerPath` is used on an assertion that does not support it, like `matchRegex`, `matchSnapshot` or the `Raw` assertions;
- `parse` is not `json` or `yaml`;
- `innerPath` is used without `parse`, or is blank, or starts with a `.`.

The assertions that support `parse` are listed in [DOCUMENT](./DOCUMENT.md#asserting-on-embedded-json-or-yaml).

Q: **Why do I have to choose between `json` and `yaml`, can't the format be detected?** <br/>
A: The format is explicit on purpose. JSON is valid YAML, so a YAML parser would also accept malformed JSON (like `{"port": 8080,}`) and would turn plain text into a string instead of failing. Auto-detection would hide these errors.

Q: **Why does `value: 1` match a parsed `1.0`?** <br/>
A: Numbers are typed by their value, not by how they are written: a whole number is an integer and anything else is a float. So `8080`, `1.0` and `1e3` are integers and `1.5` is a float. This is the same for `parse: json`, `parse: yaml` and an ordinary `path`. Large integers keep their exact value, and a whole number too large for a 64-bit integer stays a float.

Q: **Can I combine `decodeBase64`, `parse` and `innerPath`?** <br/>
A: Yes, on `equal` and `notEqual`. They always run in the same order: first the value at `path` is base64 decoded, then it is parsed, then `innerPath` is looked up in the parsed content. This lets you assert on a Secret that holds base64 encoded JSON or YAML. The order cannot be changed, so it is not possible to parse first and decode a nested field afterwards.

_Be aware! `decodeBase64` is not available on the other assertions, like `isType`. Unlike `parse`, it is then silently ignored instead of reported as an error. The [schema](./README.md#test-suite-code-completion-and-validation) does flag it in your editor._
```yaml
- equal:
    path: data.config
    decodeBase64: true
    parse: json
    innerPath: server.port
    value: 8080
```

Q: **Why does `notExists` (or `not: true`) fail when the content cannot be parsed?** <br/>
A: When the value at `path` is not a string, or is not valid JSON/YAML, an error is reported before the assertion itself runs. This error is not reversed by `not: true` or by the `not*` assertions, so a broken document is always reported, even by `notExists`.

Q: **Why does one invalid base64 value fail only that value, while an invalid JSON/YAML value fails the whole document?** <br/>
A: A failed base64 decode only fails that single value; the other values found at `path` are still decoded, parsed and checked. A parse error, on the other hand, stops checking the whole rendered document (manifest) and reports the error.

Q: **What happens when `innerPath` matches more than one value?** <br/>
A: All matches, from all values found at `path`, are put into one list, and every value in that list must pass the assertion. This works the same as a `path` that matches several values, including `--failfast` and `not`. When nothing matches, the assertion reports `unknown path <path> innerPath <innerPath>` (the inverse assertions, like `notExists`, pass instead).

Q: **The value index in the failure message does not match my `path` matches, why?** <br/>
A: The index counts positions in the combined list of `innerPath` matches, not the matches of `path`. For example, when `path` matches two strings and `innerPath` matches three values in each, the indexes run from 0 to 5.

Q: **How do I add `parse` support to another assertion?** <br/>
A: The supported assertions follow from the code itself: a validator supports `parse` when it embeds `ParseOptions`. To add support:
1. embed `ParseOptions` in the validator and call `resolveActuals` after reading the values at `path`;
1. add `parse` and `innerPath` to the assertion in `schema/helm-testsuite.json`;
1. add the assertion to the list in `pkg/unittest/schema_parse_fields_test.go`, which checks that the schema and the code agree;
1. cover it in the fixture chart `test/data/v3/parse-structured`;
1. add the assertion to the supported list in `DOCUMENT.md`.

See [ARCHITECTURE](./ARCHITECTURE.md#assert-on-parsed-jsonyaml) for the full order of execution, and [#924](https://github.com/helm-unittest/helm-unittest/pull/924) for the design notes.
