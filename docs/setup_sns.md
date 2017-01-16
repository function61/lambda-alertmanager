Settings up SNS
===============

SNS is basically a pub/sub solution from AWS.


Create topic "AlertManager-ingest"
----------------------------------

In `SNS > Topics > Create new topic`:

- Topic name = `AlertManager-ingest`
- Display name = (leave blank)

Write the `Topic ARN` down - you'll need this when setting up Lambda.


Create topic "AlertManager-alert"
---------------------------------

In `SNS > Topics > Create new topic`:

- Topic name = `AlertManager-alert`
- Display name = `ALERT` (this is shown in SMS message prefix etc.)

Write the `Topic ARN` down (for this topic as well) - you'll need this when setting up Lambda.


Add first subscriber to alert topic
-----------------------------------

Now, `SNS > Topics > AlertManager-alert > Actions > Subscribe`:

- Protocol: `Email`
- Endpoint: `your.email@example.com`

AWS just sent you an email. Open that email and confirm your subscription.
This has to be done only one per subscription.

You can later set up SMS delivery by adding a new subscription to the `AlertManager-alert` topic.


What is the difference between "ingest" and "alert" topics?
-----------------------------------------------------------

The diagram in [README](../README.md) explains this the best! Look for the SNS topics.

TL;DR: `ingest` processes high-bandwith alarms and `alert` delivers filtered low-bandwith alerts.
