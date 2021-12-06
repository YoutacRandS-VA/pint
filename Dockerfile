FROM golang:1.17.4-alpine
COPY . /src
WORKDIR /src
RUN go build ./cmd/pint

FROM debian:stable
RUN apt-get update --yes && \
    apt-get install --no-install-recommends --yes git ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY --from=0 /src/pint /usr/local/bin/pint
WORKDIR /code
CMD ["/usr/local/bin/pint"]
