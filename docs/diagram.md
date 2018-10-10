[Gravizo](https://gravizo.com/) source code:

```
digraph G {
	Prometheus;
	custom [label="Custom integration"];
	cloudwatch_alarms [label="Cloudwatch Alarms"];
	alertmanager_canary [label="HTTP(S) monitoring%5CnLambda: AlertManager Canary"];
	sns_ingest [label="SNS topic:%5CnAlertManager-ingest"];
	http [label="HTTPS (API Gateway)%5Cn- POST /alerts/ingest"];
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
```
