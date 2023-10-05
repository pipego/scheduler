FROM golang:latest AS build-stage
WORKDIR /go/src/app
COPY . .
RUN make build && \
    make plugin

FROM gcr.io/distroless/base-debian11 AS production-stage
WORKDIR /
COPY --from=build-stage /go/src/app/bin/fetch-localhost /
COPY --from=build-stage /go/src/app/bin/fetch-metalflow /
COPY --from=build-stage /go/src/app/bin/filter-nodeaffinity /
COPY --from=build-stage /go/src/app/bin/filter-nodename /
COPY --from=build-stage /go/src/app/bin/filter-noderesourcesfit /
COPY --from=build-stage /go/src/app/bin/filter-nodeunschedulable /
COPY --from=build-stage /go/src/app/bin/score-noderesourcesbalancedallocation /
COPY --from=build-stage /go/src/app/bin/score-noderesourcesfit /
COPY --from=build-stage /go/src/app/bin/scheduler /
COPY --from=build-stage /go/src/app/config/config.yml /
USER nonroot:nonroot
EXPOSE 28082
CMD ["/scheduler", "--config-file=/config.yml", "--listen-url=:28082"]
