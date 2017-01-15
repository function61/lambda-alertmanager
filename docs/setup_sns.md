Settings up SNS
===============

SNS is basically a pub/sub solution from AWS.


Create topic "AlertManager-ingest"
----------------------------------

In `SNS > Topics > Create new topic`:

- Topic name = `AlertManager-ingest`
- Display name = (leave blank)


Create topic "AlertManager-alert"
---------------------------------

In `SNS > Topics > Create new topic`:

- Topic name = `AlertManager-alert`
- Display name = `ALERT` (this is shown in SMS message prefix etc.)


Add first subscriber to alert topic
-----------------------------------

Now, `SNS > Topics > AlertManager-alert > Actions > Subscribe`:

- Protocol: `Email`
- Endpoint: `your.email@example.com`

AWS just sent you an email. Open that email and confirm your subscription.
This has to be done only one per subscription.

You can later set up SMS delivery by adding a new subscription to the `AlertManager-alert` topic.

What is the difference between ingest and alert topics?
-------------------------------------------------------

Ingest topic will be used to submit all alarms to AlertManager for ingestion.
Ingestion is the process in which we decide if we'll actually act on the alarm
or not. We'll discard the alarm if:

- We've seen it before or
- If too many alarms are firing at the moment (rate limiting)

In short, ingest topic receives unfiltered alarms (high bandwidth) and the alert
topic receives only a few alerts (low bandwidth).

This means that the alarm producers can send the alarm repeatedly to the ingest
topic safely as long as the alarm condition is firing. Many producers like Prometheus
do this to minimize state in their part.
