FROM golang:1.18.2 AS build

ARG ssh_prv_key
ARG ssh_pub_key
RUN git config --global url.git@github.com:.insteadOf https://github.com/
# Authorize SSH Host
RUN mkdir -p /root/.ssh && \
    chmod 0700 /root/.ssh && \
    ssh-keyscan github.com > /root/.ssh/known_hosts

# Add the keys and set permissions
RUN echo "$ssh_prv_key" > /root/.ssh/id_rsa && \
    echo "$ssh_pub_key" > /root/.ssh/id_rsa.pub && \
    chmod 600 /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa.pub

RUN go env -w GOPRIVATE=go.eloylp.dev/goomerang

WORKDIR /src
COPY . .
# Dont run your programs as root.
RUN useradd -u 10001 nonprivuser


# CGO_ENABLED=0    -Disable CGO to discard CGO stuff and completely use pure Go net stack.
# -trimpath        -Skip your local paths in error stacktraces, take only from the root of the repo.
# -tags timetzdata -Add tzdata for be docker image indepent. Around ~800KB more size.
RUN CGO_ENABLED=0 go build -trimpath -tags timetzdata -o client

FROM scratch
# Copy CA certificates from build image.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy users table from build image
COPY --from=build /etc/passwd /etc/passwd
# Copy binary build image
COPY --from=build /src/client /app/client

# metrics scrape port
EXPOSE 9971
# Go pprof port
EXPOSE 6060

USER nonprivuser
WORKDIR /app
CMD ["./client"]