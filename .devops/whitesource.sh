#!/bin/bash
set -e
THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $HOME # Prevents codebase contamination.
CIRCLE_WORKDIR=$(eval cd $CIRCLE_WORKING_DIRECTORY ; pwd)

echo 'Installing Oracle JRE 8...'
JAVA_HOME="$HOME/jre"
mkdir $JAVA_HOME
curl -jLs -H "Cookie: oraclelicense=accept-securebackup-cookie" http://download.oracle.com/otn-pub/java/jdk/8u131-b11/d54c1d3a095b4ff2b6607d096fa80163/jre-8u131-linux-x64.tar.gz | tar -C $JAVA_HOME --strip-components=1 -xz

echo 'Installing WhiteSource FSA...'
WS_AGENT="$HOME/whitesource.jar"
curl -Ls 'https://s3.amazonaws.com/file-system-agent/whitesource-fs-agent-18.5.1.jar' > $WS_AGENT

echo 'Configuring WhiteSource FSA...'
WS_CONFIG_SRC="$CIRCLE_WORKDIR/whitesource.cfg"
WS_CONFIG="$HOME/whitesource.cfg"
cp $WS_CONFIG_SRC $WS_CONFIG
echo "apiKey=$WHITESOURCE_ORG_TOKEN" >> $WS_CONFIG
echo "productToken=$WHITESOURCE_PRODUCT_TOKEN" >> $WS_CONFIG

echo 'Starting WhiteSource FSA...'
set +e
$JAVA_HOME/bin/java -jar $WS_AGENT -c $WS_CONFIG -d $CIRCLE_WORKDIR
RESULT=$?
set -e
ls -al
ls -al $CIRCLE_WORKDIR
# TODO
# Add failure logic where applicable.
# Success=0, Error=-1, Policy Violation=-2, Client Failure=-3, Connection Failure=-4

echo 'Retrieving FSA results...'
PROJECT_FILE="$HOME/whitesource-results.json"
export WHITESOURCE_PROJECT_TOKEN=$(curl -H 'Content-Type: application/json' -X POST --data "{\"requestType\" : \"getOrganizationProjectVitals\",\"orgToken\" : \"$WHITESOURCE_ORG_TOKEN\"}" 'https://saas.whitesourcesoftware.com/api' | jq -r ".projectVitals[] | select(.name=='$WHITESOURCE_PROJECT_NAME') | .token")
curl -H 'Content-Type: application/json' -X POST --data "{\"requestType\" : \"getProjectHierarchy\",\"projectToken\" : \"$WHITESOURCE_PROJECT_TOKEN\"}" -o $PROJECT_FILE 'https://saas.whitesourcesoftware.com/api'

echo 'Stripping all transitive dependencies from the results...'
npm install --no-spin follow-redirects@1.5.0 > /dev/null 2>&1
node $THIS_DIR/whitesource.js $CIRCLE_WORKDIR

echo 'Cleaning up...'
rm -rf $JAVA_HOME
rm $WS_AGENT
rm $WS_CONFIG
