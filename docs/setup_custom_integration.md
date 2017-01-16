Setting up custom integration over HTTPS
========================================

This API is pretty much explained in the [Setting up API gateway](./setup_apigateway.md) guide,
but here's a short summary until I get time to write a better tutorial on this.

NOTE: replace `REDACTED` with your `API ID` from API Gateway.


List firing alerts
------------------

AKA un-acknowledged alerts.

```
$ curl https://REDACTED.execute-api.us-west-2.amazonaws.com/prod/alerts
[
	{
		"alert_key": "1",
		"subject": "www.example.com",
		"timestamp": "2017-01-15T12:12:04.018Z",
		"details": "I dont like the page"
	}
]
```


Submit an alarm
---------------

```
$ curl -H 'Content-Type: application/json' -X POST -d '{"subject": "www.example.com", "details": "I dont like the page"}' https://REDACTED.execute-api.us-west-2.amazonaws.com/prod/alerts/ingest
"OK => alert saved to database and queued for delivery"
```

Or if the alert is already firing, you'll get back response `This alert is already firing. Discarding the submitted alert.`.

Alternate way: if you app uses AWS-SDK, you can also submit the alarm for ingestion by posting to the `AlertManager-ingest` SNS topic.


Acknowledge an alert
--------------------

```
$ curl -H 'Content-Type: application/json' -X POST -d '{"alert_key": "1"}' https://REDACTED.execute-api.us-west-2.amazonaws.com/prod/alerts/acknowledge
"Alert 1 deleted"
```


Receiving fired alerts via webhook
----------------------------------

Go to `SNS > Topics > AlertManager-alert > Create subscription > HTTP or HTTPS`.

Firing alerts are sent to any subscribers listed in this topic. If I added a webhook, I would have these subscriptions:

- Email: ops@example.com
- SMS: +358 40 123 456
- Webhook(HTTPS): https://example.com/api/alert-firing-webhook
