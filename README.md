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
