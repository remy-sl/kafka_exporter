FROM golang:1.15-alpine AS builder

ARG CGO_ENABLED=0

RUN apk --no-cache add git
WORKDIR /code
# Download dependencies separately to cache this layer as
# dependencies do not change that often.
COPY go.* ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=$CGO_ENABLED go build -o kafka_exporter ./cmd/kafka_exporter
RUN chmod +x kafka_exporter

FROM alpine:latest AS runtime
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /code/kafka_exporter /usr/local/bin
ENTRYPOINT ["kafka_exporter"]