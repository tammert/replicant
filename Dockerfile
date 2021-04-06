FROM golang:1.16 as builder

WORKDIR /src/
ADD . .

RUN CGO_ENABLED=0 go build -o replicant

#######################################
FROM gcr.io/distroless/static
COPY --from=builder /src/replicant /
ENTRYPOINT ["/replicant"]
