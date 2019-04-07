# ------------------------------------------------------------------------------
# Test image
# ------------------------------------------------------------------------------
FROM golang:alpine AS test_img

RUN apk update && apk upgrade && apk add --no-cache git

COPY . /src
COPY .git/refs/heads/master /commit-hash

WORKDIR /src

CMD ["go", "test"]


# ------------------------------------------------------------------------------
# Development image
# ------------------------------------------------------------------------------
FROM test_img AS dev_img

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
   go build -gcflags "all=-N -l" -o /gobench

ENTRYPOINT /gobench


# ------------------------------------------------------------------------------
# Production image
# ------------------------------------------------------------------------------
FROM alpine:3.7 as prod_img
COPY --from=test_img /commit-hash /
COPY --from=dev_img /gobench /

ENTRYPOINT /gobench
