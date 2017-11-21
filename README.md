# helm test

write the test:
```yaml
suite: name
range:
tests:
- it: should ...
  files:
  set:
  expects:
    - matchSnapshot:
      document:
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
