FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /api ./cmd/api

# Export a cross-compiled binary without running a foreign-platform image.
FROM scratch AS binary
COPY --from=build /api /api

FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS runtime
RUN apk add --no-cache ca-certificates tzdata && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /api /app/api
USER app
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=30s --retries=3 CMD wget -qO- http://127.0.0.1:8080/readyz || exit 1
ENTRYPOINT ["/app/api"]
