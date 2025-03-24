#!/bin/bash

X_Platform=${1:-"."}
X_Build_Dir=${2:-"build"}

mkdir -p $X_Build_Dir
echo "[output_dir]: $(realpath $X_Build_Dir)"

GitCommit=$(git rev-parse --short HEAD 2>/dev/null)
GitBranch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
X_AppVersion=${3:-"$(date -u +'Y%y.%m.%d').$GitCommit"}
# BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

build_x() {
  local x_dst=${1:-"$X_Build_Dir/mcapevent.$GitBranch.$GitCommit.x"}
  echo "[building]:$(realpath $x_dst)"
  go build -buildvcs=false -ldflags "\
    -X main.Version=$X_AppVersion \
    -X main.GitCommit=$GitBranch.$GitCommit \
    -X main.BuildTime=$(date +'%Y-%m-%dT%H:%M:%SZ')" \
    -o $x_dst

  echo "[finished]: $x_dst"
}

build_me() {
  local x_dst=${1:-"$X_Build_Dir/mcapevent"}
  echo "[building]:$(realpath $x_dst)"
  build_x $x_dst
}

build_all() {
  local x_dir=${1:-"$X_Build_Dir"}
  PLATFORMS=("linux/amd64" "windows/amd64" "darwin/arm64")
  for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT_NAME="${x_dir}/${GOOS}-${GOARCH}/app"
    if [ $GOOS = "windows" ]; then
      OUTPUT_NAME="$OUTPUT_NAME.exe"
    fi
    # GOOS=$GOOS GOARCH=$GOARCH go build -o $OUTPUT_NAME main.go
    GOOS=$GOOS GOARCH=$GOARCH build_me $OUTPUT_NAME
  done
}

if [[ "$X_Platform" = "x" ]]; then
  build_x
elif [[ "$X_Platform" = "." ]]; then
  build_me
else
  build_all
fi
