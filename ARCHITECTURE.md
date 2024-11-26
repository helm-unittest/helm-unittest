# Helmunittest Architecture

This page gives a short overview on how the Helm Unittest is being setup.
This can help identifying elements for resolving issues and creating new features.

## Overview

![Architectural Overview](./.images/helmunittest-arch.svg)

### Input
This section describes all input required to let helm-unittest do its work.
For each section there is a short summary on the reason for its existence.
Details on what the configure can be found in [Documents](./DOCUMENT.md)

#### Test Suite(s)
This input contains the definition of the test suite.
A helmunittest project can have multiple testsuites.
The testsuite is the starting point, the properties which you can set here are defined for the whole helmcharts, which you want to unittest.

| Features/Quirks | The feature | The quirk |
| ------ | ----- | ----- |
| When templates are defined, it filters only those templates | You can focus on only templates to be tests | templates that depends on other templates, need to be added to the list |
| When release and/or capabillities are defined, it overrides the behavour of the chart itself | You can control the behaviour of the template rendering and have a predictable outcome | |
| When KubernetesProvider is used, it generates a Mock that can be used for e.g. Lookups | You can control the outcome of the rendering of dependend resources that are not part of the helm template | Only use the KubernetesProvider when you expect output, without the provider dependend resources return an empty object |

#### Test Job(s)
This input contains the definition of the test job.
Within the Testsuite it is possible (highly recommended) to have multiple testjobs.

| Features/Quirks | The feature | The quirk |
| ------ | ----- | ----- |
| For each testjob the complete helmchart will be reloaded | Isolation of the tests | Large helmchart project can take a bit longer to load |
| When values or sets is used, the values will be merged with the upper values | Similar approach as Helm itself | Unsetting specific values can be tricky |

#### Assertion(s)
This input contains the definition of the validation to be executed.
Within the Testjob it is possible (highly recommended) to have multiple assertions.

| Features/Quirks | The feature | The quirk |
| ------ | ----- | ----- |
| When the template is not used, the assertion will check all filtered templates | Maximum reuse of assertion | When filtering different resources, be aware the validation is correct for all types |
| When values or sets is used, the values will be merged with the upper values | Similar approach as Helm itself | Unsetting specific values can be tricky |

### Core
The core logic of the Helmunittest.

#### Testrunner
The testrunner is running al the logic in the following steps:

1. parsing the Input;
1. loads the helmchart;
1. filters only the templates identified;
1. renders the identified templates;
1. executes all validators;
1. generates output;

| Features/Quirks | The feature | The quirk |
| ------ | ----- | ----- |
| Where possible helm package is used to render the resources  | Be as closest to the helm behaviour | |

#### Validators
The validator is the implementation of a specific assertion.
Each assertion has it's own validator containing a basic Interface.
```go
type Validatable interface {
	Validate(context *ValidateContext) (bool, []string)
}
```

Each validator follows a similar pattern on evaluation (happy flow)
1. retrieve rendered resources (manifests)
1. validationSuccess is false
1. for each manifest
    1. retrieve values of set path
    1. check found values contains results
    1. foreach found value
        1. check singleValueValidation
        1. singleValueValidationSuccess is true
    1. determine validationSuccess
    1. validationSuccess is true
1. determine validationSuccess
1. validationSuccess is true

| Features/Quirks | The feature | The quirk |
| ------ | ----- | ----- |
| validation fails by default | To prevent false positive on validations |  |
| singleValueValidation also validates inverse checks | | |
| validations have a failfast option | When a singleValueValidation fails, it stops directly, which can speed-up the test | Not all failures will be shown | 

#### Command
The Command containing the commandline parameters.

| Features/Quirks | The feature | The quirk |
| ------ | ----- | ----- |
| For debugging a deviating name is used | Explicit debugging of the plugin to gather as much information as possible | Helm is passing all parameters to the plugin, except for the debug flag |

### Output
The output is generating the results.
The testresults can be used to output to console or output to file.

#### Testresults
The testresults object containing all required information to have human readable or test automation based output. 
