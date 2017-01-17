var AWS = require('aws-sdk');

var ALERT_TOPIC = process.env.ALERT_TOPIC;

if (!ALERT_TOPIC) {
	throw new Error("ALERT_TOPIC not defined");
}

var MAX_FIRING_ALERTS = +process.env.MAX_FIRING_ALERTS || 5;

var sns = new AWS.SNS();

var dynamodb = new AWS.DynamoDB();

// DynamoDB records are typed by this horrible convention:
//     { "name": { S: "a string value" }, "age": { N: "123" }  }
// that actually means
//     { "name": "a string value", "age": 123 }
// read more @ https://www.npmjs.com/package/dynamodb-data-types
function unwrapDynamoDBTypedObject(kv) {
	var ret = {};
	
	if (typeof (kv) !== 'object') {
		throw new Error('Can only unwrapDynamoDBTypedObject key-value objects');
	}

	for (var key in kv) {
		var value = kv[key];
		
		if (!('S' in value)) {
			throw new Error('Support only strings at the moment');
		}
		
		ret[key] = value.S;
	}
	
	return ret;
}

function httpSucceedAndLog(context, succeedResult) {
	console.log(succeedResult);

	context.succeed({
		statusCode: 200,
		body: JSON.stringify(succeedResult)
	});
}

function failAndLog(context, failResult) {
	console.log(failResult);
	context.fail(failResult);
	// TODO: do we have to respond to failures with the HTTP statusCode
	//       wrapper when using Lambda-proxy in ApiGateway?
	/*
	context.fail({
		statusCode: 500,
		body: JSON.stringify(failResult)
	});
	*/
}

var apis = {
	'GET /alerts': function (event, context) {
		dynamodb.scan({
			TableName: 'alertmanager_alerts'
		}, function (err, data){
			if (err) {
				failAndLog(context, err);
				return;
			}

			// httpSucceedAndLog(context, data.Items.map(unwrapDynamoDBTypedObject));
			context.succeed({
				statusCode: 200,
				body: JSON.stringify(data.Items.map(unwrapDynamoDBTypedObject))
			});
		});
	},

	'POST /alerts/acknowledge': function (event, context) {
		var eventBody = JSON.parse(event.body);

		dynamodb.deleteItem({
			TableName: 'alertmanager_alerts',
			Key: {
				alert_key: { S: eventBody.alert_key }
			}
		}, function (err, res){
			if (err) {
				failAndLog(context, err);
				return;
			}

			httpSucceedAndLog(context, 'Alert ' + eventBody.alert_key + ' deleted');
		});
	},

	'POST /alerts/ingest': function (event, context) {
		// Old logic:
		// var alert_key = event.body.subject.replace(/[^a-zA-Z0-9]/g, '_');

		var eventBody = JSON.parse(event.body);
		var ts = eventBody.timestamp ? new Date(eventBody.timestamp) : new Date();

		function trySaveOnce(tryNumber) {
			if (tryNumber >= 5) {
				failAndLog(context, new Error('Enough retries - should not happen'));
				return;
			}

			dynamodb.scan({
				TableName: 'alertmanager_alerts',
				Limit: 1000 // whichever comes first, 1 MB or 1 000 records
			}, function (err, firingAlertsResult){
				if (err) {
					context.fail(err);
					return;
				}

				if (firingAlertsResult.Items.length >= MAX_FIRING_ALERTS) {
					// should not context.fail(), as otherwise the submitter could re-try again (that would be undesirable)
					httpSucceedAndLog(context, "Max alerts already firing. Discarding the submitted alert.");
					return;
				}

				var items = firingAlertsResult.Items.map(unwrapDynamoDBTypedObject);

				var largestNumber = 0;

				for (var i = 0; i < items.length; ++i) {
					if (items[i].subject === eventBody.subject) {
						// should not context.fail(), as otherwise the submitter could re-try again (that would be undesirable)
						httpSucceedAndLog(context, "This alert is already firing. Discarding the submitted alert.");
						return;
					}

					largestNumber = Math.max(largestNumber, +items[i].alert_key);
				}

				// if you want to test ConditionalCheckFailedException, don't increment this
				var nextNumber = largestNumber + 1;

				dynamodb.putItem({
					Item: {
						alert_key: { S: nextNumber.toString() },
						timestamp: { S: ts.toISOString() },
						subject: { S: eventBody.subject },
						details: { S: eventBody.details }
					},
					TableName: 'alertmanager_alerts',

					// http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.SpecifyingConditions.html
					ConditionExpression: 'attribute_not_exists(alert_key)'
				}, function(err, data) {
					// ConditionalCheckFailedException means that betweem the scan() -> putItem() cycle somebody else
					// wrote a record, so we must run the scan again to calculate new sequence number to write the record with.
					//
					// why not just use GUIDs? that would solve the saving problem, but we actually want ConditionalCheckFailedException
					// to happen, so firing alert count has no race conditions. if we get 1 000 alerts firing at the same second,
					// we don't want to send 100 alerts (possibly SMS) to the ops team. instead, we want to send exactly the configured maximum count.
					if (err && err.code === 'ConditionalCheckFailedException') {
						console.log('ConditionalCheckFailedException -> trying again');
						trySaveOnce(tryNumber + 1);
						return;
					}
					else if (err) {
						failAndLog(context, err);
						return;
					}

					httpSucceedAndLog(context, 'OK => alert saved to database and queued for delivery');
				});
			});
		}

		trySaveOnce(1);
	},

	// Prometheus integration
	'POST /prometheus-alertmanager/api/v1/alerts': function (event, context) {
		var eventBody = JSON.parse(event.body);
		/*	eventBody=
			[
			  {
			    "labels": {
			      "alertname": "dummy_service_down",
			      "instance": "10.0.0.17:80",
			      "job": "prometheus-dummy-service"
			    },
			    "annotations": {
			      
			    },
			    "startsAt": "2017-01-17T08:42:07.804Z",
			    "endsAt": "2017-01-17T08:42:52.806Z",
			    "generatorURL": "http://f67e003689ac:9090/graph?g0.expr=fictional_healthmeter%7Bjob%3D%22prometheus-dummy-service%22%7D+%3C+50\\u0026g0.tab=0"
			  }
			]
		*/

		// FIXME: this only takes care of the first alert
		var subject = eventBody.length === 1 ?
			eventBody[0].labels.alertname :
			'Alert count not 1, was: ' + eventBody.length; // Fallback for actually letting us now

		// convert to simulated incoming HTTP message
		var simulatedHttpEvent = {
			httpMethod: 'POST',
			path: '/alerts/ingest',
			body: JSON.stringify({
				subject: subject,
				details: "Job: " + eventBody[0].labels.job + "\nInstance: " + eventBody[0].labels.instance,
				timestamp: eventBody[0].startsAt
			})
		};
		
		// run the main dispatcher again
		exports.handler(simulatedHttpEvent, context);
	},

	'SNS: ingest': function (event, context) {
		/* {
				Type: 'Notification',
				MessageId: 'ab47c45e-ba4a-5572-9694-94ac6ab4daa8',
				TopicArn: 'arn:aws:sns:us-east-1:...:AlertManager-ingest',
				Subject: 'http://whoami.prod4.fn61.net/',
				Message: 'find="_c70e24a08b3a" NOT in body: Hostname: c70e24a08b3a...',
				Timestamp: '2017-01-13T12:57:34.911Z',
				SignatureVersion: '1',
				Signature: '...',
				SigningCertUrl: '...',
				UnsubscribeUrl: '...',
				MessageAttributes: {} }
		*/
		
		// convert to simulated incoming HTTP message
		var simulatedHttpEvent = {
			httpMethod: 'POST',
			path: '/alerts/ingest',
			body: JSON.stringify({
				subject: event.Records[0].Sns.Subject,
				details: event.Records[0].Sns.Message,
				timestamp: new Date(event.Records[0].Sns.Timestamp).toISOString()
			})
		};
		
		// run the main dispatcher again
		exports.handler(simulatedHttpEvent, context);
	},

	'DynamoDB: alertmanager_alerts': function (event, context) {
		if (event.Records.length !== 1) { // should not happen, as trigger config: BatchSize=1
			failAndLog(context, new Error("Record count must be 1"));
			return;
		}

		if (event.Records[0].eventName !== 'INSERT') {
			httpSucceedAndLog(context, 'Not interested in eventName: ' + event.Records[0].eventName);
			return;
		}

		var record = unwrapDynamoDBTypedObject(event.Records[0].dynamodb.NewImage);

		sns.publish({
			Message: record.subject + "\n\n" + record.details,
			Subject: record.subject,
			TopicArn: ALERT_TOPIC
		}, function (err) {
			if (err) {
				failAndLog(context, err);
			} else {
				httpSucceedAndLog(context, "Alert sent with subject: " + record.subject);
			}
		});
	},

	'Unsupported record type': function (event, context) {
		console.log(event.Records[0]);
		failAndLog(context, new Error("Unsupported record type"));
	},

	'API not found': function (event, context) {
		console.log(event);
		failAndLog(context, new Error('API not found: ' + event.api));
	},

	'Unsupported event': function (event, context) {
		console.log(event);
		failAndLog(context, new Error("Unsupported event"));
	}
};

exports.handler = function(event, context) {
	var operation = 'Unsupported event';

	// HTTP request via API gateway, api = "<HTTP_METHOD> <URL>", example: "POST /ingest"
	if (event.httpMethod) {
		var apiName = event.httpMethod + ' ' + event.path;

		operation = apiName in apis ? apiName : 'API not found';
	}
	else if (event.Records) { // SNS or DynamoDB notification
		// curiously, SNS and DynamoDB event EventSource and EventVersion fields differ in case
		if (event.Records[0].EventSource && event.Records[0].EventSource === 'aws:sns' && event.Records[0].EventVersion === '1.0') {
			operation = 'SNS: ingest';
		} else if (event.Records[0].eventSource && event.Records[0].eventSource === 'aws:dynamodb' && event.Records[0].eventVersion === '1.1') {
			operation = 'DynamoDB: alertmanager_alerts';
		} else {
			operation = 'Unsupported record type';
		}
	}

	console.log(operation + " ->");

	apis[ operation ](event, context);
};
