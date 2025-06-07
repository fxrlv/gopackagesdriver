# gopackagesdriver

A [driver](https://pkg.go.dev/golang.org/x/tools/go/packages#hdr-The_driver_protocol) for `gopls` with Bazel support.

## Why

The [rules_go's driver](https://github.com/bazel-contrib/rules_go/wiki/Editor-and-tool-integration) requires Bazel to execute `bazel query` and `bazel build` which takes a lot of time. \
This completely ruins the entire experience and joy of writing in `go`.

## Usage

The driver must be installed in `PATH` or named explicitly with the `GOPACKAGESDRIVER` environment variable.

```console
go install github.com/fxrlv/gopackagesdriver@latest
```

Once installed, point the `GOPACKAGESDRIVER_WORKSPACE` environment variable to a Bazel workspace.
