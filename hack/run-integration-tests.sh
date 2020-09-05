#!/usr/bin/env bash

BUILD_NUMBER="${BUILD_NUMBER:-0}"
export COMPOSE_PROJECT_NAME=fabclient_${BUILD_NUMBER}

docker build . -t fabclient;

for i in {1..3}
do
  if [[ $i -eq 1 ]]; then
    docker-compose -f testdata/hyperledger-fabric-network/docker-compose.yaml --project-name $COMPOSE_PROJECT_NAME up -d orderer.dummy.com;
  else
    docker-compose -f testdata/hyperledger-fabric-network/docker-compose.yaml --project-name $COMPOSE_PROJECT_NAME up -d orderer${i}.dummy.com;
  fi
  sleep 5;
done

docker-compose -f testdata/hyperledger-fabric-network/docker-compose.yaml --project-name $COMPOSE_PROJECT_NAME up -d peer0.org1.dummy.com peer0.org2.dummy.com;
sleep 5;

check=$(docker ps -aq -f status=exited  | wc -l);
check=${check##*( )};

if [[ "$check" -ne 0 ]]; then
  exit 1;
fi

docker run --rm --network=${COMPOSE_PROJECT_NAME}_default -v `pwd`:/go/src/github.com/TommyStarK/fabric-go-client \
  fabclient bash -c "go test -v -race -failfast --cover -covermode=atomic -coverprofile=coverage.out -mod=vendor; exit $?";

rc=$?;

XARGS="xargs -r"
if [[ "$OSTYPE" == "darwin"* ]]; then
  XARGS="xargs"
fi

docker-compose -f testdata/hyperledger-fabric-network/docker-compose.yaml --project-name $COMPOSE_PROJECT_NAME down;
docker ps -a | grep "dev-peer0.\(org1\|org2\).dummy.com-fcacc_\(1.0\|2.0\)" | awk '{print $1}'| $XARGS docker rm -f;
docker images | grep "dev-peer0.\(org1\|org2\).dummy.com-fcacc_\(1.0\|2.0\)" | awk '{print $3}'| $XARGS docker rmi -f;
docker rmi -f fabclient;
docker volume prune -f;
docker network prune -f;

exit $rc;
