if [ -z "$1" ]; then
    echo "Please provide a version number as an argument."
    exit 1
fi

# won't be as strict w/ database versioning
TAG=$1

docker build \
    -t ${GAR_REGION}-docker.pkg.dev/${GAR_PROJECT_ID}/${GAR_REPOSITORY}/${PROJECT_NAME}:${TAG} \
    -t ${GAR_REGION}-docker.pkg.dev/${GAR_PROJECT_ID}/${GAR_REPOSITORY}/${PROJECT_NAME}:latest \
    -f Dockerfile_migrate \
    .
