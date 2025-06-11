FROM alpine:latest
RUN   apk add --no-cache tzdata ca-certificates
WORKDIR /app
VOLUME [ "/app/pb_data/" ]
EXPOSE 9096

COPY bilive-auth /app/bilive-auth
# start PocketBase
ENTRYPOINT [ "/app/bilive-auth", "serve", "--http=0.0.0.0:9096" ]
CMD []
