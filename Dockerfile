FROM debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

COPY /bin/volta /bin/volta

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT [ "/bin/volta"]