FROM golang:1.10 as build-env
ARG svc
WORKDIR /go/src/github.com/traiana/okro/hellod
COPY ./vendor vendor
COPY ./api api
COPY ./$svc $svc
RUN CGO_ENABLED=0 go build --ldflags '-extldflags "-static"' -o /tmp/app ./$svc

FROM gcr.io/distroless/base:debug
COPY --from=build-env /tmp/app /app
ENTRYPOINT ["/app"]
