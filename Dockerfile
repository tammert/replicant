FROM golang:1.18 as builder

WORKDIR /src/
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

RUN CGO_ENABLED=0 go build -ldflags '-w -s' -o replicant

FROM gcr.io/distroless/static
COPY --from=builder /src/replicant /bin/
ENTRYPOINT ["/bin/replicant"]
