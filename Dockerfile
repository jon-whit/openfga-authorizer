FROM cgr.dev/chainguard/go:1.21@sha256:40fb38b3f61d1ecdecd2bfe3d14fa621ab5e9ecb4c9ebba590276c2992f3e282 AS builder

WORKDIR /app

RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# build with cache
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/authorizer ./cmd/authorizer

FROM cgr.dev/chainguard/go:latest-dev

EXPOSE 9443
EXPOSE 9002
COPY --from=builder /bin/authorizer /authorizer

ENTRYPOINT ["/authorizer"]