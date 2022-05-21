FROM golang:1.18.2 AS build
WORKDIR /src
COPY . .
# Dont run your programs as root.
RUN useradd -u 10001 nonprivuser
# CGO_ENABLED=0    -Disable CGO to discard CGO stuff and completely use pure Go net stack.
# -trimpath        -Skip your local paths in error stacktraces, take only from the root of the repo.
# -tags timetzdata -Add tzdata for be docker image indepent. Around ~800KB more size.
RUN CGO_ENABLED=0 go build -trimpath -tags timetzdata -o proxy

FROM scratch
# Copy CA certificates from build image.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy users table from build image
COPY --from=build /etc/passwd /etc/passwd
# Copy binary build image
COPY --from=build /src/proxy /app/proxy

# metrics scrape port
EXPOSE 9971
# serving port
EXPOSE 8080
# Go pprof port
EXPOSE 6060


USER nonprivuser
WORKDIR /app
CMD ["./proxy"]