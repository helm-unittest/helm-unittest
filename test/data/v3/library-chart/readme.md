# Helm library Chart

- [What is library chart](https://helm.sh/docs/topics/library_charts/)

## Library Chart testing folder structure

```
.
├── Chart.yaml                # type: library (required)
├── readme.md
├── templates                 # common templates
│   ├── _annotations.tpl
│   └── _names.tpl
└── tests
    └── chart
        ├── Chart.yaml             # type: application (required)
        ├── charts
        │   └── common-1.0.0.tgz   # chart must be installed required as
        ├── templates
        │   └── pod.yaml           # helm template with testing logic
        └── tests
            └── unit               # tests folder (required)
                └── metadata_test.yaml
```

## Library Chart testing commands

```sh
# from the project root directory
helm dependency build test/data/v3/library-chart/tests/chart
helm unittest -f 'tests/unit/*.yaml' --color test/data/v3/library-chart/tests/chart

# or from testing directory
cd test/data/v3
helm dependency build library-chart/tests/chart
helm unittest -f 'tests/unit/*.yaml' --color library-chart/tests/chart
```

## TODO

- [ ] Helm `unit-test` to accept a flag to pull dependencies
