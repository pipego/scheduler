FROM golang:latest AS build-stage
WORKDIR /go/src/app
COPY . .
RUN apt update && \
    apt install -y upx
RUN make build
RUN URL=$(curl -L -s https://api.github.com/repos/pipego/plugin-fetch/releases/latest | grep -o -E "https://(.*)plugin-fetch_(.*)_linux_amd64.tar.gz") && \
    curl -L -s $URL | tar xvz -C bin && \
    URL=$(curl -L -s https://api.github.com/repos/pipego/plugin-filter/releases/latest | grep -o -E "https://(.*)plugin-filter_(.*)_linux_amd64.tar.gz") && \
    curl -L -s $URL | tar xvz -C bin && \
    URL=$(curl -L -s https://api.github.com/repos/pipego/plugin-score/releases/latest | grep -o -E "https://(.*)plugin-score_(.*)_linux_amd64.tar.gz") && \
    curl -L -s $URL | tar xvz -C bin

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
