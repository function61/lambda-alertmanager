NOT PRODUCTION READY
====================

Disclaimer: work-in-progress. Not ready for general use.

lambda-canary
=============

This is a [Lambda](https://aws.amazon.com/lambda/)-function to poll your important web properties'
online status for example every five minutes.

Features:

- Runs entirely on AWS's super-reliable architecture
- Costs practically nothing to run
- Supports `http` and `https` checks
- Runs multiple checks in parallel - it is super fast
- Modular architecture. Plugs into [Prometheus](https://prometheus.io/docs/alerting/clients/)-compatible alert API.
- Lambda-canary therefore does not deliver alerts - only invokes the alerting process of another system.
- Tries to minimize false positives by retrying one time

function61 has another product: alert-ingestor, which you can connect to lambda-canary. Its features:

- Rate limiting for alerts (so you don't get 1 000 alerts at once for one problem)
- Runs also on AWS Lambda - no infrastructure to manage and operate
- Delivers alerts to [AWS SNS](https://aws.amazon.com/sns/) (supports transports like email, SMS, webhook etc.)
- Easy to configure from AWS' UI which email addresses and SMS numbers get notified on alerts
- SMS messages are practically free (at least my alerting volume does not cost me anything)

Configuration
-------------

Todo: how to separate code from configuration in Lambda?

Architecture
------------

```
     +-----------------------+
     |                       |
     | lambda-canary         |                                                               +-------+
     | - Ping web properties +-----------+                                                   |       |
     |                       |           |                                                   | Email |
     +-----------------------+           |                                                   |       |
                                         |                                                   +---^---+
                                         |                                                       |
+----------------------------+         +-v--------------------+      +------------------+        |
|                            |         |                      |      |                  +--------+
|  Prometheus                +---------> alert-ingestor       +------> SNS topic: alert |
|  - Alerts based on metrics |         | - Rate limiting etc. |      |                  +--------+
|                            |         |                      |      +------------------+        |
+----------------------------+         +-^--------+----^------+                                  |
                                         |        |    |                                     +---v-+
                                         |        |    |                                     |     |
     +-----------------------+           |     +--v----+--+                                  | SMS |
     |                       |           |     |          |                                  |     |
     | Other alert producers +-----------+     | DynamoDB |                                  +-----+
     |                       |                 | - State  |
     +-----------------------+                 |          |
                                               +----------+
```

Alerting
--------

Purposefully, the alerting pipeline has been separated from this.

The alerting currently invokes another Lambda function (via API gateway) to do the alerting.

### API for alerting

It should respond to HTTP requests in the following format:

```
Content-Type: application/json

{
	"subject": "http://www.example.com/",
	"details": "Timeout encountered (8000 ms)"
}
```
