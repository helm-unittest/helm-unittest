# Contributing

Issues and PRs are welcome!

## Issues and Improvements

When you find an Issue or Improvement, please chech first if it already occurs
otherwise create a [New Issue](https://github.com/helm-unittest/helm-unittest/issues/new/choose)

If you have a Issue related to security, please follow our [Security Policy](./SECURITY.md)

## Developing

Before start developing this plugin, you must have [Go](https://golang.org/doc/install) >= 1.24 installed, and run:

```
git clone git@github.com:helm-unittest/helm-unittest.git
cd helm-unittest
```

And please make CI passed when request a PR which would check following things:

- `gofmt` no changes needed. Please run `gofmt -w -s .` before you commit.
- `go test ./pkg/unittest/...` passed.

In some cases you might need to manually fix the tests in `*_test.go`. If the snapshot tests (of the plugin's test code) failed, you need to run:

```
UPDATE_SNAPSHOTS=true go test ./...
```

This update the snapshot cache file and please add them before you commit.

In order to run the post-render tests successful, make sure the tool yq is installed on the path.
