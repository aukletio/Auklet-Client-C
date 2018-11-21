#!/bin/bash
set -e
if [[ "$1" == "" ]]; then
  echo "ERROR: env not provided."
  exit 1
fi
ENVDIR=$1
VERSION="$(cat ~/.version)"
VERSION_SIMPLE=$(cat ~/.version | xargs | cut -f1 -d"+")
export TIMESTAMP="$(date --rfc-3339=seconds | sed 's/ /T/')"
if [[ "$1" == "staging" ]]; then
  BASE_URL='https://api-staging.auklet.io'
elif [[ "$1" == "qa" ]]; then
  BASE_URL='https://api-qa.auklet.io'
else
  BASE_URL='https://api.auklet.io'
fi

echo 'Gathering license files for dependencies...'
REPO_DIR=$(eval cd $CIRCLE_WORKING_DIRECTORY ; pwd)
LICENSES_DIR="$REPO_DIR/cmd/client/licenses"
cp LICENSE $LICENSES_DIR
cd .devops
npm install --no-spin follow-redirects@1.5.0 > /dev/null 2>&1
node licenses.js "$REPO_DIR" "$LICENSES_DIR"
rm -rf node_modules package-lock.json
cd ..
echo

echo 'Generating packed resource files...'
curl -sSL https://github.com/gobuffalo/packr/releases/download/v1.11.0/packr_1.11.0_linux_amd64.tar.gz | tar -xz packr
./packr -v -z
echo

echo 'Compiling client for target architectures...'
echo
GO_LDFLAGS="-X github.com/aukletio/Auklet-Client-C/version.Version=$VERSION -X github.com/aukletio/Auklet-Client-C/version.BuildDate=$TIMESTAMP -X github.com/aukletio/Auklet-Client-C/config.StaticBaseURL=$BASE_URL"
PREFIX='auklet-client'
S3_BUCKET='auklet'
S3_PREFIX='client'
export GOOS=linux
declare -a archs=("amd64" "arm" "arm64" "mips" "mipsle" "mips64" "mips64le")
for a in "${archs[@]}"
do
  echo "=== $GOOS/$a ==="
  if [[ "$a" == "arm" ]]; then
    # We don't support ARM 5 or 6.
    export GOARM=7
  fi
  GOARCH=$a go build -ldflags "$GO_LDFLAGS" -o $PREFIX-$GOOS-$a-$VERSION ./cmd/client
done

echo 'Installing AWS CLI...'
sudo apt-get -y install awscli > /dev/null 2>&1

if [[ "$ENVDIR" == "release" ]]; then
  echo 'Erasing production client binaries in public S3...'
  aws s3 rm s3://$S3_BUCKET/$S3_PREFIX/latest/ --recursive
fi

echo 'Uploading client binaries to S3...'
# Decide which S3 subdir to put the binaries in.
case "$ENVDIR" in
  beta)
    S3_ENVDIR='staging'
    ;;
  rc)
    S3_ENVDIR='qa'
    ;;
  release)
    S3_ENVDIR='production'
    ;;
  *)
    echo "ERROR: invalid ENVDIR '$ENVDIR' provided."
    exit 1
    ;;
esac
# Iterate over each file and upload it to S3.
for f in ${PREFIX}-*; do
  # Upload to the internal bucket.
  S3_LOCATION="s3://auklet-profiler/$S3_ENVDIR/$S3_PREFIX/$VERSION/$f"
  aws s3 cp $f $S3_LOCATION
  # Upload to the public bucket for production builds.
  if [[ "$ENVDIR" == "release" ]]; then
    # Copy to the public versioned directory.
    VERSIONED_NAME="${f/$VERSION/$VERSION_SIMPLE}"
    aws s3 cp $S3_LOCATION s3://$S3_BUCKET/$S3_PREFIX/$VERSION_SIMPLE/$VERSIONED_NAME
    # Copy to the public "latest" directory.
    LATEST_NAME="${f/$VERSION/latest}"
    aws s3 cp $S3_LOCATION s3://$S3_BUCKET/$S3_PREFIX/latest/$LATEST_NAME
  fi
done
