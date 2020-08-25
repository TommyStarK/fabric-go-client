#!/usr/bin/env bash
BUILD_NUMBER="${BUILD_NUMBER:-0}"
export COMPOSE_PROJECT_NAME=fabclient_${BUILD_NUMBER}

docker build . -t fabclient;

docker-compose -f testdata/hyperledger-fabric-network/docker-compose.yaml --project-name $COMPOSE_PROJECT_NAME up -d;

sleep 20;

check=$(docker ps -aq -f status=exited  | wc -l);
check=${check##*( )};

if [[ "$check" -ne 0 ]]; then
  exit 1;
fi

docker run --rm --network=${COMPOSE_PROJECT_NAME}_default -v `pwd`:/go/src/github.com/TommyStarK/fabric-go-client \
  fabclient bash -c "go test -v -race -failfast --cover -covermode=atomic -coverprofile=coverage.out -mod=vendor; exit $?";

rc=$?;

# TOODO: grep chaincode , rm container / images
docker-compose -f testdata/hyperledger-fabric-network/docker-compose.yaml --project-name $COMPOSE_PROJECT_NAME down;
docker images | grep dev-peer | awk '{print $3}' | xargs docker rmi -f;
docker ps -aq -f status=exited | xargs docker rm -f;
docker images -qf dangling=true | xargs docker rmi -f;
docker volume prune -f;
docker network prune -f;

exit $rc;
