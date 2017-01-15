Setting up AlertManager
=======================


Create Lambda function
----------------------

- Go to `Lambda > Create a Lambda function > Blank function`.
- Do not configure any triggers at this time (just hit next).
- Name: `AlertManager`
- Description: `AlertManager main: ingestor & alerter`
- Runtime: `Node.js 4.3`
- FIXME Code entry type: `Upload a file from Amazon S3`
- FIXME S3 link URL: `https://s3.amazonaws.com/files.function61.com/lambda-canary/2017-01-13.zip`
- Paste code from `ingestor/index.js`
- Enable encryption helpers: leave unchecked

Env variables:

- `ALERT_TOPIC` = `arn:aws:sns:__REGION__:__AWS_ACCOUNT_ID__:AlertManager-alert` (replace `__AWS_ACCOUNT_ID__` with your ID and `__REGION__` with your region)

Role config:

- Handler: (leave as is)
- Role: leave as is (`Choose existing role`)
- Existing role: `AlertManager`

Advanced config:

- Memory (MB): leave as is (`128`)
- Timeout: `1 min`

Okay now hit `Create function`.


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
