FROM alpine:latest
RUN   apk add --no-cache tzdata ca-certificates
WORKDIR /app
EXPOSE 8090

COPY bilive-auth /bilive-auth
# start PocketBase
ENTRYPOINT [ "/bilive-auth"]
CMD []
