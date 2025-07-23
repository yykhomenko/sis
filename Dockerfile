FROM golang:alpine AS build-env
RUN mkdir /main
WORKDIR /main
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -v -a -installsuffix cgo -o main ./cmd/sis

FROM scratch
COPY --from=build-env /main .
ENTRYPOINT ["/main"]
