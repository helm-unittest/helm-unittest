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
    - matchValue:
        path:
        value:
    - matchPattern:
        path:
        pattern:
    - contain:
    - containMap:
    - isNotNull:
    - isNotEmpty:
    - isKindOf:
    - isApiVersion:
    - hasDocumentsCount:
```
