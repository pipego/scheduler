FROM golang:latest AS build-stage
WORKDIR /go/src/app
COPY . .
RUN apt update && \
    apt install -y upx
RUN make build && \
    make plugin

FROM gcr.io/distroless/base-debian11 AS production-stage
WORKDIR /
COPY --from=build-stage /go/src/app/bin/plugin-fetch /
COPY --from=build-stage /go/src/app/bin/plugin-filter /
COPY --from=build-stage /go/src/app/bin/plugin-score /
COPY --from=build-stage /go/src/app/bin/scheduler /
COPY --from=build-stage /go/src/app/config/config.yml /
USER nonroot:nonroot
EXPOSE 28082
CMD ["/scheduler", "--config-file=/config.yml", "--listen-url=:28082"]
