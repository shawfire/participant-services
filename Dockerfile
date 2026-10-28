# syntax=docker/dockerfile:1
FROM golang:1.27.3-alpine3.23 AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY graph/ ./graph/
RUN go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && go build -trimpath -ldflags="-s -w" -o /out/mockmulesoft ./cmd/mockmulesoft \
 && go build -trimpath -ldflags="-s -w" -o /out/localissuer ./cmd/localissuer

FROM alpine:3.23 AS runtime
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 65532:65532

FROM runtime AS localissuer
COPY --from=build /out/localissuer /usr/local/bin/localissuer
EXPOSE 8443
ENTRYPOINT ["/usr/local/bin/localissuer"]

FROM runtime AS mockmulesoft
COPY --from=build /out/mockmulesoft /usr/local/bin/mockmulesoft
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/mockmulesoft"]

FROM runtime AS api
COPY --from=build /out/api /usr/local/bin/api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
