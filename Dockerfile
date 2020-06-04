FROM coco/go-alpine-plus-seabolt:v1.3.0

ENV PROJECT=annotations-rw-neo4j
ENV ORG_PATH="github.com/Financial-Times"
ENV SRC_FOLDER="${GOPATH}/src/${ORG_PATH}/${PROJECT}"

COPY . /${SRC_FOLDER}/
WORKDIR ${SRC_FOLDER}

RUN BUILDINFO_PACKAGE="${ORG_PATH}/service-status-go/buildinfo." \
    && VERSION="version=$(git describe --tag --always 2> /dev/null)" \
    && DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)" \
    && REPOSITORY="repository=$(git config --get remote.origin.url)" \
    && REVISION="revision=$(git rev-parse HEAD)" \
    && BUILDER="builder=$(go version)" \
    && LDFLAGS="-s -w -X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'" \
    && CGO_ENABLED=1 go build -mod=readonly -a -o /artifacts/${PROJECT} -ldflags="${LDFLAGS}" -tags seabolt_static

COPY ./suggestion-config.json /artifacts/suggestion-config.json
COPY ./annotation-config.json /artifacts/annotation-config.json

FROM scratch
WORKDIR /
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /artifacts/* /

CMD [ "/annotations-rw-neo4j" ]
