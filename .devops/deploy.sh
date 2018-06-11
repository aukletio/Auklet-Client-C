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
GO_LDFLAGS="-X main.Version=$VERSION -X main.BuildDate=$TIMESTAMP -X github.com/ESG-USA/Auklet-Client/config.StaticBaseURL=$BASE_URL"

echo 'Compiling client for target architectures...'
echo
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

if [[ "$ENVDIR" == "production" ]]; then
  echo 'Erasing production client binaries in public S3...'
  aws s3 rm s3://$S3_BUCKET/$S3_PREFIX/latest/ --recursive
fi

echo 'Uploading client binaries to S3...'
# Iterate over each file and upload it to S3.
for f in ${PREFIX}-*; do
  # Upload to the internal bucket.
  S3_LOCATION="s3://auklet-profiler/$ENVDIR/$S3_PREFIX/$VERSION/$f"
  aws s3 cp $f $S3_LOCATION
  # Upload to the public bucket for production builds.
  if [[ "$ENVDIR" == "production" ]]; then
    # Copy to the public versioned directory.
    VERSIONED_NAME="${f/$VERSION/$VERSION_SIMPLE}"
    aws s3 cp $S3_LOCATION s3://$S3_BUCKET/$S3_PREFIX/$VERSION_SIMPLE/$VERSIONED_NAME
    # Copy to the public "latest" directory.
    LATEST_NAME="${f/$VERSION/latest}"
    aws s3 cp $S3_LOCATION s3://$S3_BUCKET/$S3_PREFIX/latest/$LATEST_NAME
  fi
done

# Push to public GitHub repo.
# The hostname "aukletio.github.com" is intentional and it matches the "ssh-config-aukletio" file.
if [[ "$ENVDIR" == "production" ]]; then
  echo 'Pushing production branch to github.com/aukletio...'
  mv ~/.ssh/config ~/.ssh/config-bak
  cp .devops/ssh-config-aukletio ~/.ssh/config
  chmod 400 ~/.ssh/config
  git remote add aukletio git@aukletio.github.com:aukletio/Auklet-Client-C.git
  git push aukletio HEAD:master
  git remote rm aukletio
  rm -f ~/.ssh/config
  mv ~/.ssh/config-bak ~/.ssh/config
fi
