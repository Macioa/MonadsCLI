<div style="background-color: white;">

## Docker test matrix

This folder contains Linux distro Dockerfiles that build and run `monadscli`.

### Images
- `Dockerfile.ubuntu`
- `Dockerfile.debian`
- `Dockerfile.alpine`
- `Dockerfile.fedora`
- `Dockerfile.ubuntu-pwsh` (PowerShell on Ubuntu; built as `linux/amd64`)

### Build and run a single image
```bash
docker build -f test/docker/Dockerfile.ubuntu -t monadscli-test:ubuntu .
docker run --rm monadscli-test:ubuntu
```

### Build and run all images
```bash
./test/docker/run-all.sh
```

The PowerShell image runs under amd64 emulation on Apple Silicon.

### Go version override
All Dockerfiles install Go from go.dev and accept `GO_VERSION` (default: 1.24.0):
```bash
docker build -f test/docker/Dockerfile.debian \
  --build-arg GO_VERSION=1.24.0 \
  -t monadscli-test:debian .
```

</div>
