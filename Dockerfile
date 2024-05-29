FROM golang:latest as builder

# install xz
# Uncomment if using upx to compress the binary
# RUN apt-get update && apt-get install -y \
#     xz-utils \
#     && rm -rf /var/lib/apt/lists/*
# # install UPX
# ADD https://github.com/upx/upx/releases/download/v3.94/upx-3.94-amd64_linux.tar.xz /usr/local
# RUN xz -d -c /usr/local/upx-3.94-amd64_linux.tar.xz | \
#     tar -xOf - upx-3.94-amd64_linux/upx > /bin/upx && \
#     chmod a+x /bin/upx

ARG GITHUB_TOKEN

ENV GITHUB_TOKEN=$GITHUB_TOKEN

# create a working directory
COPY . /exemplar
WORKDIR /exemplar


# Force token authentication for fetching LambdatestIncPrivate repos
RUN git config --global  url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/LambdatestIncPrivate".insteadOf "https://github.com/LambdatestIncPrivate"

# Add src

# Build binary
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o exemplar
# Uncomment only when build is highly stable. Compress binary.
# RUN strip --strip-unneeded ts
# RUN upx ts


# use a minimal alpine image
FROM alpine:latest
# add ca-certificates in case you need them
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
# set working directory
WORKDIR /root
# copy the binary from builder
COPY --from=builder /exemplar .
RUN touch .exemplar.yml
# run the binary
CMD ["./exemplar"]