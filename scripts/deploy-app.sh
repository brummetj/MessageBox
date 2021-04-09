#!/usr/bin/env bash

DEPLOY_DB=false
PUSH=false
IMAGE_NAME="mbapp"
VERSION="latest"
DOCKERHUB_URI="Some-private-hub.com"
BASEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
ZONE="blue"
NETWORK="mb"
BASE_IMAGE=""
CONF=""
RUN="all"
BUILD=false
STOP=
NGINX_IMAGE_NAME="mb-nginx"
NGINX_CONTAINER_NAME="mb"
FQUIT=false
while true; do
    opt="$1"
    [[ -z "${opt}" ]] && break
    case "${opt}" in
        --help|-h)
            cat <<EOM
Build and run MessageBox Application for deployment

Usage: $0 [<options>]

where:
  <options> may be:
    --zone            | -z   Tag the image with a specific zone it should be deployed too. Blue is default
    --push            | -p   Push the image to the registry.
    --version         | -v   The version to build and tag with
    --build-app       | -b   Build the image!
    --base-image      | -i   Give a specific base build image, useful for adding new base images with tools
    --deploy-db       | -db  Start the database container... useful for dev but it should already be deployed and running!
    --network         | -n   The docker network to run on... really we want to use kube tho :)
    --run             | -r   Run MB images, options ['all', 'blue', 'green']
    --stop-image      | -s   Stop a image by zone or all , options ['all', 'blue', 'green']
    --force-quit      | -q   Quit the application after running --stop-image or if you just wanna not run anything :D
EOM
            exit 0
            ;;
        --push|-p)
            PUSH=true
            shift
            ;;
        --force-quit|-q)
            FQUIT=true
            shift
            ;;
        --version|-v)
            VERSION=$2
            shift
            shift
            ;;
        --base-image|-i)
            BASE_IMAGE=$2
            shift
            shift
            ;;
        --zone|-z)
            ZONE=$2
            shift
            shift
            ;;
        --network|-n)
            NETWORK=$2
            shift
            shift
            ;;
        --build|-b)
            BUILD=true
            shift
            ;;
        --deploy-db|-db)
            DEPLOY_DB=true
            shift
            ;;
        --run|-r)
            RUN=$2
            shift
            shift
            ;;
        --stop-image|-s)
            STOP=$2
            shift
            shift
            ;;
        -*)
            shift
            ;;
        *)
            break
            ;;
    esac
done
CONTAINER_NAME_BLUE="${IMAGE_NAME}-blue"
CONTAINER_NAME_GREEN="${IMAGE_NAME}-green"
IMAGE_NAME="${IMAGE_NAME}":"${VERSION}"

if [[ ${STOP} == 'blue' ]]; then
  docker stop "${CONTAINER_NAME_BLUE}"
  docker rm "${CONTAINER_NAME_BLUE}"
fi

if [[ ${STOP} == 'green' ]]; then
  docker stop "${CONTAINER_NAME_GREEN}"
  docker rm "${CONTAINER_NAME_GREEN}"
fi

if [[ ${STOP} == "all" ]]; then
  docker stop "${CONTAINER_NAME_BLUE}"
  docker rm "${CONTAINER_NAME_BLUE}"
  docker stop "${CONTAINER_NAME_GREEN}"
  docker rm "${CONTAINER_NAME_GREEN}"
  docker stop "${NGINX_CONTAINER_NAME}"
  docker rm "${NGINX_CONTAINER_NAME}"
fi

if [[ ${FQUIT} == true ]]; then
  echo "Quitting!"
  exit 0
fi

# Need to create the network first
echo "Creating docker network: ${NETWORK}"
docker network create "${NETWORK}" || true

if [[ $DEPLOY_DB == true ]]; then
  "${BASEDIR}"/deploy-db.sh "${NETWORK}"
  sleep 10 # The DB takes a bit to init so lets chill
fi

if [[ ${BUILD} == true ]]; then
  echo "Building MB image: ${IMAGE_NAME}"
  echo "${BASEDIR}"/../deploy/docker/Dockerfile.prod
  docker build -f "${BASEDIR}"/../deploy/docker/Dockerfile -t "${IMAGE_NAME}" "${BASEDIR}"/../ \
  --build-arg BUILD_IMAGE="${BASE_IMAGE}" --no-cache
fi

if [[ ${PUSH} == true ]]; then
  echo "Oops!! we aren't pushing anywhere yet :)"
fi

echo "Building nginx image"
if [[ ${ZONE} == "blue" ]]; then
  CONF="mb-blue.conf"
elif [[ ${ZONE} == "green" ]]; then
  CONF="mb-green.conf"
else
  echo "unknown zone"
  exit 1
fi

docker build -f "${BASEDIR}"/../deploy/nginx/Dockerfile "${BASEDIR}"/../deploy/nginx/ --build-arg CONF="${CONF}" -t "${NGINX_IMAGE_NAME}"

if [[ ${RUN} == 'all' ]]; then

  docker run -d --name "${CONTAINER_NAME_BLUE}" -v /tmp:/tmp --net="${NETWORK}" "${IMAGE_NAME}" \
      mb -e /usr/local/mb/src/config/serviceConfigBlue.yml
  docker run -d --name "${CONTAINER_NAME_GREEN}" -v /tmp:/tmp --net="${NETWORK}" "${IMAGE_NAME}" \
      mb -e /usr/local/mb/src/config/serviceConfigGreen.yml

elif [[ ${RUN} == 'blue' ]]; then

    docker run -d --name "${CONTAINER_NAME_BLUE}" -v /tmp:/tmp --net="${NETWORK}" "${IMAGE_NAME}" \
      mb -e /usr/local/mb/src/config/serviceConfigBlue.yml

elif [[ ${RUN} == 'green' ]]; then

    docker run -d --name "${CONTAINER_NAME_GREEN}" -v /tmp:/tmp --net="${NETWORK}" "${IMAGE_NAME}" \
      mb -e /usr/local/mb/src/config/serviceConfigGreen.yml

fi

sleep 5

docker run  -d --name mb --net="${NETWORK}"  -p 3001:80 ${NGINX_IMAGE_NAME}

sleep 5 # sleeps are to give app some startup time

echo "Application Deployed and running in a docker environment"