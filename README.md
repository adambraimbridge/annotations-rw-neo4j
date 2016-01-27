# Annotations Reader/Writer for Neo4j (annotations-rw-neo4j)

__An API for reading/writing annotations into Neo4j. Expects the annotations json supplied to be in the format that comes out of the annotations consumer.__

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

## Endpoints

### PUT
/content/{annotatedContentId}/annotations

Each annotation is added as a MENTIONS relationship between the content and a concept.

This acts as a replace - all existing annotations are removed, and the new ones are created. This is because we get these
annotations wholesale from the concept extraction service, which annotates the whole content on each publish.

Supplying an empty list as the request body will remove all annotations for the content.

A successful PUT results in 200.

We run queries in batches. If a batch fails, all failing requests will get a 500 server error response.

Invalid json body input will result in a 400 bad request response.

NB: annotations don't have identifiers themselves currently - the id in the json is the id of the concept that is annotating the content.

See [this doc](https://docs.google.com/document/d/1FE-JZDYJlKsxOIuQQkPwyyzcOkJQn8L3nNy1H8A8eDo) for more details.

Example:

    curl -XPUT -H "X-Request-Id: 123" -H "Content-Type: application/json" localhost:8080/content/3fa70485-3a57-3b9b-9449-774b001cd965/annotations --data
    "@examplePutBody.json"

NB: Although provenances are supplied is a list, we don't expect to get more than one provenance: we will take the scores from that one
and apply them to the MENTIONS relationship that we are creating for that annotation.  If there is no provenance, we'll still
create a MENTIONS relationship, it just won't have score, agent and time properties.

### GET
/content/{annotatedContentId}/annotations
This internal read should return what got written (i.e., this isn't the public annotations read API)

If not found, you'll get a 404 response.

Empty fields are omitted from the response.
`curl -H "X-Request-Id: 123" localhost:8080/content/3fa70485-3a57-3b9b-9449-774b001cd965/annotations`

### DELETE
/content/{contentId}/annotations/{annotationType}/{conceptId}

conceptId here is just the uuid, not the full URI

NB: /content/{contentId}/annotations/mentions/{conceptId} is used to allow annotations to be removed in Spyglass (not sure whether this is much used because if the content is republished, we lose the fact an annotation was deleted).

Other paths are only added here for ease of testing.

Will return 204 if successful, 404 if not found

See [this doc](https://docs.google.com/document/d/1cySUlTuSYlv8ANikLlfToezSiRERa0sBdO2eVqy1FXM) for more details.

`curl -XDELETE -H "X-Request-Id: 123" localhost:8080/3fa70485-3a57-3b9b-9449-774b001cd965/annotations/2e8d937e-935c-3586-9137-eff2bc1cdd8d`


## Healthchecks
* Check connectivity [http://localhost:8080/__health](http://localhost:8080/__health)
* Ping: [http://localhost:8080/__ping](http://localhost:8080/__ping)

## TODO
### Things to resolve, check or otherwise investigate
* Handle DELETE
* Handle GET
* Properties on annotation
  * annotatedDate - is this required, what Format, set as current date if missing, duplicate as long for sorting ?
  * originatingSystem - required ?
  * annotatedBy - required ?
* Write an ANNOTATED_BY (aka isAnnotatedBy) relationship for all annotation relationships so we can count them ?

### API specific
* Complete Test cases
* Runbook
* Update or new API documentation based on original google docs

### Cross cutting concerns
* Allow service to start if neo4j is unavailable at startup time
