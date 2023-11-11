FROM alpine:latest
RUN   apk add --no-cache  ca-certificates
WORKDIR /app
VOLUME [ "/app/pb_data/" ]
EXPOSE 8090

COPY bilive-auth /app/bilive-auth
# start PocketBase
ENTRYPOINT [ "/app/bilive-auth", "serve", "--http=0.0.0.0:8090" ]
CMD []
