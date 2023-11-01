#!/bin/bash

MODULE=$(grep module go.mod | cut -d\  -f2)
BINBASE=${MODULE##*/}
VERSION=${VERSION:-$GITHUB_REF_NAME}
VERSION=${VERSION:-0.0.0}
COMMIT_HASH="$(git rev-parse --short HEAD 2>/dev/null)"
COMMIT_HASH=${COMMIT_HASH:-00000000}
DIRTY=$(git diff --quiet 2>/dev/null || echo '-dirty')
BUILD_TIMESTAMP=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILDER=$(go version)

[ -d dist ] && rm -rf dist
mkdir dist

# For version in sub module
# "-X '${MODULE}/main.Version=${VERSION}'"

LDFLAGS=(
  "-X 'main.Version=${VERSION}'"
  "-X 'main.CommitHash=${COMMIT_HASH}${DIRTY}'"
  "-X 'main.BuildTimestamp=${BUILD_TIMESTAMP}'"
  "-X 'main.Builder=${BUILDER}'"
)
echo "[*] Build info"
echo  "   Version=${VERSION}"
echo  "   CommitHash=${COMMIT_HASH}${DIRTY}"
echo  "   BuildTimestamp=${BUILD_TIMESTAMP}"
echo  "   Builder=${BUILDER}"

#echo "[*] go get"
#go get .

echo "[*] go builds:"
#set -x
for DIST in {linux,openbsd,freebsd,windows}/{amd64,arm,arm64,386} darwin/{amd64,arm64}; do
#for DIST in linux/{amd64,386}; do
  GOOS=${DIST%/*}
  GOARCH=${DIST#*/}
  echo "[+]   $DIST:"
  echo "[-]    - build"
  SUFFIX=""
  [ "$GOOS" = "windows" ] && SUFFIX=".exe"
  TARGET=${BINBASE}-${GOOS}-${GOARCH}
  env CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="${LDFLAGS[*]}" -mod vendor -o dist/${TARGET}${SUFFIX}
  if [ -z "$NOCOMPRESS" ]; then
    echo "[-]    - compress"
    if [ "$GOOS" = "windows" ]; then
      xz --keep dist/${TARGET}${SUFFIX}
      (cd dist; zip -qm9 ${TARGET}.zip ${TARGET}${SUFFIX})
    else
      xz dist/${TARGET}
    fi
  fi
done

echo "[*] sha256sum"
(cd dist; sha256sum *) | tee ${BINBASE}.sha256sum
mv ${BINBASE}.sha256sum dist/

#echo "[*] pack"
#tar -cvf all.tar -C dist/ . && mv all.tar dist

echo "[*] done"
