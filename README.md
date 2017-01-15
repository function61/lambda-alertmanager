lambda-alertmanager?
--------------------

- Provides simple & reliable alerting for your infrastructure.
- Uses so little resources that it is practically free to run.
- Monitors your web properties for being up, receive alerts from Prometheus, Amazon CloudWatch alarms, alarms via SNS topic
  or any custom HTTP integration (as JSON).
- Runs **entirely** on AWS' reliable infrastructure (after setup nothing for you to manage or fix). The compute part is Lambda,
  but we also use DynamoDB + streams (for state), IAM (for sandboxing AlertManager), API Gateway (for inbound https integrations),
  CloudWatch Events (for scheduling) and SNS (inbound alarm receiving, outbound alert delivery).
- Acknowledge -model: each separate alarm is alerted only once until it is acknowledged from UI,
  even if the same alarm is submitted again. F.ex. Prometheus sends the same alert continuously
  until the issue is resolved, but of course you want to receive the alert only once).
- Rate limiting: if shit hits the fan and your hundreds of alarms trigger at once, you only get alerts
  for the first, say, 10 alarms. The rate limit is configurable.


Can send alerts to you (or many people) via:
--------------------------------------------

- SMS ([free: <= 100 alerts/month](https://aws.amazon.com/sns/sms-pricing/))
- Email
- Webhook
- Push to mobile device (though SMS is better in cases when you are travelling or otherwise not reachable via mobile data)
- Any combination of these (I use SMS + Email)
- Or [anything that SNS supports](https://aws.amazon.com/sns/details/) (the above are just SNS transports)


Can directly monitor:
---------------------

- http/https checks via AlertManager-Canary component (included but optional):
  checks that your web properties are up and triggering an alert if not. Can even check all your properties
  at 1 minute intervals, and runs efficiently because all the checks are executed in parallel. Tries to minimize
  false positives by retrying each failed check once before generating an alarm.


Integrates with:
----------------

- Supports receiving alerts from [Prometheus](https://prometheus.io/).
- Supports receiving alerts via SNS (= directly plugs into Amazon CloudWatch Alerts)
  or any other SNS-publishing source. For example we receive alerts from CloudWatch -> AlertManager if our
  queue processors stop processing work.
- Supports receiving alerts over https as JSON.


Diagram
-------

![Graph](https://g.gravizo.com/g?
  digraph G {
  	Prometheus;
  	custom [label="Custom integration"];
  	cloudwatch_alarms [label="Cloudwatch Alarms"];
  	alertmanager_canary [label="HTTP(S) monitoring%5CnLambda: AlertManager Canary"];
  	sns_ingest [label="SNS topic:%5CnAlertManager-ingest"];
  	http [label="HTTPS%5Cn- API gateway"];
  	receive_alarm [label="Receive alarm%5CnLambda: AlertManager"];
  	alarm_already_triggering [label="Alarm already triggering?"];
  	Discard;
  	rate_limit_exceeded [label="Rate limit exceeded?"];
  	store_alarm_dynamodb [label="Store alarm%5Cn- DynamoDB"];
  	dynamodb_trigger [label="DynamoDB trigger%5Cn- Row inserted: send alert"];
  	sns_alert [label="SNS topic:%5CnAlertManager-alert"];
  	sns_email [label="Email%5Cnops@example.com"];
  	sns_sms [label="SMS%5Cn+358 40 123 456"];
  	Prometheus -> http;
  	custom -> http;
  	cloudwatch_alarms -> sns_ingest;
  	alertmanager_canary -> sns_ingest;
  	sns_ingest -> receive_alarm;
  	http -> receive_alarm;
  	receive_alarm -> alarm_already_triggering;
  	alarm_already_triggering -> Discard [label=" yes"];
  	alarm_already_triggering -> rate_limit_exceeded [label=" no"];
  	rate_limit_exceeded -> Discard [label=" yes"];
  	rate_limit_exceeded -> store_alarm_dynamodb;
  	store_alarm_dynamodb -> dynamodb_trigger;
  	dynamodb_trigger -> sns_alert;
  	sns_alert -> sns_email;
  	sns_alert -> sns_sms;
  }
)
