docker run --rm \
    --network host \
    -v ${HOME}/.config/gcloud:/root/.config/gcloud \
    $(basename $(pwd)):latest serve -d ${DATABASE_URL}