FROM golang:1.26-alpine AS build-env
ARG TARGETOS=linux
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/sis ./cmd/sis

FROM scratch
COPY --from=build-env /out/sis /sis
USER 65534:65534
EXPOSE 8080
ENTRYPOINT ["/sis"]
