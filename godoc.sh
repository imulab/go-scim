#/bin/bash

mkdir -p /tmp/tmpgoroot/doc
rm -rf /tmp/tmpgopath/src/github.com/imulab/go-scim
mkdir -p /tmp/tmpgopath/src/github.com/imulab/go-scim
tar -c --exclude='.git' --exclude='tmp' . | tar -x -C /tmp/tmpgopath/src/github.com/imulab/go-scim
echo -e "open http://localhost:6060/pkg/github.com/imulab/go-scim\n"
GOROOT=/tmp/tmpgoroot/ GOPATH=/tmp/tmpgopath/ godoc -http=localhost:6060