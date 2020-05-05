# UPP - Annotations RW Neo4j

Annotations RW is a service and an API for reading/writing annotations into Neo4j

## Code

annotations-rw-neo4j

## Primary URL

https://upp-prod-delivery-glb.upp.ft.com/__annotations-rw-neo4j/

## Service Tier

Platinum

## Lifecycle Stage

Production

## Delivered By

content

## Supported By

content

## Known About By

- elitsa.pavlova
- kalin.arsov
- miroslav.gatsanoga
- ivan.nikolov
- marina.chompalova

## Host Platform

AWS

## Architecture

Annotations RW reads messages from the ConceptAnnotations queue, generates annotations in the graph database (Neo4j) and forwards them on to the PostConceptAnnotations queue. The Annotation Writer owns all annotations that are linked to a piece of content that are written with lifecycle=annotations-v1, annotations-next-video or annotations-pac.

## Contains Personal Data

No

## Contains Sensitive Data

No

## Dependencies

- upp-kafka
- upp-neo4j-cluster

## Failover Architecture Type

ActiveActive

## Failover Process Type

FullyAutomated

## Failback Process Type

FullyAutomated

## Failover Details

The service is deployed in both Delivery clusters. The failover guide for the cluster is located here:
https://github.com/Financial-Times/upp-docs/tree/master/failover-guides/delivery-cluster

## Data Recovery Process Type

NotApplicable

## Data Recovery Details

The service does not store data, so it does not require any data recovery steps.

## Release Process Type

PartiallyAutomated

## Rollback Process Type

Manual

## Release Details

If the new release does not change the way kafka messages are consumed and/or produce it's safe to deploy it without cluster failover.

## Key Management Process Type

Manual

## Key Management Details

To access the service clients need to provide basic auth credentials.
To rotate credentials you need to login to a particular cluster and update varnish-auth secrets.

## Monitoring

Service in UPP K8S delivery clusters:

- Delivery-Prod-EU health: https://upp-prod-delivery-eu.upp.ft.com/__health/__pods-health?service-name=annotations-rw-neo4j
- Delivery-Prod-US health: https://upp-prod-delivery-us.upp.ft.com/__health/__pods-health?service-name=annotations-rw-neo4j

## First Line Troubleshooting

https://github.com/Financial-Times/upp-docs/tree/master/guides/ops/first-line-troubleshooting

## Second Line Troubleshooting

Please refer to the GitHub repository README for troubleshooting information.
