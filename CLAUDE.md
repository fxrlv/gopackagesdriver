# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go packages driver for gopls that provides fast Bazel support by creating overlay filesystems for generated files. Instead of running expensive `bazel query` and `bazel build` commands, it directly reads from Bazel's cache directory and maps generated files to workspace locations.

The codebase is designed with reusable components that could be applied to other applications beyond just the gopls driver.

## Build and Development Commands

```bash
# Build the driver binary
go build -o gopackagesdriver .

# Install globally
go install .

# Check module dependencies
go mod tidy
```

## High-Level Architecture

**🚨 CRITICAL**: This is a one-shot command invoked by gopls for each request, not a long-running server. Any caching must be file-based, not in-memory.

**🚨 PERFORMANCE REALITY**: Based on actual measurements, `packages.Load()` is the main bottleneck and cannot be optimized away. Directory walking is fast by comparison. Bazel calls are heavy but minimized through file caching.

### Core Components

1. **main.go**: Entry point, handles JSON communication with gopls
2. **driver.go**: Core orchestration, workspace validation, overlay creation
3. **bazel.go**: Cache directory walking, finds generated `.go` files
4. **overlay.go**: File I/O operations for overlay system
5. **load.go**: Package loading wrapper around `packages.Load()`
6. **modfile.go**: Go module file parsing for protobuf support

### Environment & Configuration

- **`GOPACKAGESDRIVER_WORKSPACE`** (required): Workspace root directory
- **`GOPACKAGESDRIVER_BAZEL`** (optional): Bazel command, defaults to "bazel"
- **File-based caching**: Uses temp directories to cache `bazel info` results

## Essential Implementation Notes

**Protobuf Handling**: `CutProtoPrefix()` handles `_go_proto_` generated files for proper path mapping.

**Performance Focus**:
- `packages.Load()` dominates execution time (measured)
- Directory walking provides minimal optimization benefit
- File-based caching is essential for bazel calls
- In-memory caching is wrong for this one-shot command architecture

The overlay system is the key component - it maps Bazel cache files to workspace paths for `go list -overlay`.
