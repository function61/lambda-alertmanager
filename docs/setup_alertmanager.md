Setting up AlertManager
=======================


Create Lambda function
----------------------

- Go to `Lambda > Create a Lambda function > Blank function`.
- Do not configure any triggers at this time (just hit next).
- Name: `AlertManager`
- Description: `AlertManager main: ingestor & alerter`
- Runtime: `Node.js 4.3`
- Download
  [alertmanager-2017-01-16.zip](https://s3.amazonaws.com/files.function61.com/alertmanager/alertmanager-2017-01-16.zip)
  to your desktop and then upload to Lambda

Env variables:

- `ALERT_TOPIC` = ARN of your alert topic (mine looked like `arn:aws:sns:us-west-2:426466625513:AlertManager-alert`)

Role config:

- Handler: (leave as is)
- Role: leave as is (`Choose existing role`)
- Existing role: `AlertManager`

Advanced config:

- Memory (MB): leave as is (`128`)
- Timeout: `1 min`

Okay now hit `[ Create function ]`.


Add trigger for "alertmanager_alerts" DynamoDB table
----------------------------------------------------

Go to `Triggers > Add > DynamoDB`:

- Table = `alertmanager_alerts`
- Batch size = `1`
- Starting position = `Trim horizon`


Add trigger for "AlertManager-ingest" SNS topic
-----------------------------------------------

Go to `Triggers > Add > SNS`:

- Topic = `AlertManager-ingest`

This topic allows the ingestor to receive alerts from Canary, CloudWatch & other SNS-compatible sources.
