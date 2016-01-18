# Annotations Reader/Writer for Neo4j (annotations-rw-neo4j)

__An API for reading/writing annotations into Neo4j. Expects the people json supplied to be in the format that comes out of the annotations consumer.__

## Build & deployment etc:
*TODO*
_NB You will need to tag a commit in order to build, since the UI asks for a tag to build / deploy_
* [Jenkins view](http://ftjen10085-lvpr-uk-p:8181/view/annotations-private-rw)
* [Build and publish to forge](http://ftjen10085-lvpr-uk-p:8181/job/annotations-private-rw)
* [Deploy to test or production](http://ftjen10085-lvpr-uk-p:8181/job/annotations-private-rw)


## Installation & running locally
* `go get -u github.com/Financial-Times/annotations-private-rw`
* `cd $GOPATH/src/github.com/Financial-Times/annotations-private-rw`
* `go test ./...`
* `go install`
* `$GOPATH/bin/annotations-private-rw --neo-url={neo4jUrl} --port={port} --log-level={DEBUG|INFO|WARN|ERROR}`
_All arguments are optional.
--neo-url defaults to http://localhost:7474/db/data, which is the out of box url for a local neo4j instance.
--port defaults to 8080.
--log-level defaults to INFO
See help text for other arguments._
* curl http://localhost:8080/annotations/{content_uuid} | json_pp

## API definition
Based on the following google docs:
* [Replace](https://docs.google.com/document/d/1FE-JZDYJlKsxOIuQQkPwyyzcOkJQn8L3nNy1H8A8eDo)
  `PUT /content/{annotatedContentId}/annotations`
* Read is the inverse of the PUT / Replace (i.e. this is not the public annotation reader)
  `GET /content/{annotatedContentId}/annotations`
* [Delete](https://docs.google.com/document/d/1cySUlTuSYlv8ANikLlfToezSiRERa0sBdO2eVqy1FXM)
  `DELETE /content/{contentId}/annotations/mentions/{conceptId}`

## Healthchecks
* Check connectivity [http://localhost:8080/__health](http://localhost:8080/__health)
* Ping: [http://localhost:8080/__ping](http://localhost:8080/__ping)

## TODO
### Things to resolve, check or otherwise investigate
* Do we need to do the V1-V2 facade (Internal) request for GET requests ?

### API specific
* Complete Test cases
* Runbook
* Update or new API documentation based on original google docs

### Cross cutting concerns
* Allow service to start if neo4j is unavailable at startup time
* Rework build / deploy (low priority)
  * Suggested flow:
    1. Build & Tests
    1. Publish Release (using konstructor to generate vrm)
    1. Deploy vrm/hash to test/prod
