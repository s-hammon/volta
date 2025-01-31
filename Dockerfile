FROM debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

COPY /bin/volta /bin/volta

ENTRYPOINT [ "/bin/volta"]