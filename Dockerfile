FROM debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

COPY /bin/volta /bin/volta

EXPOSE 8080

ENTRYPOINT [ "/bin/volta"]