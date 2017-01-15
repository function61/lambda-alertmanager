Setting up API gateway
======================

Foreword: Amazon API gateway is not the most elegant of products that AWS has shipped.
It still isn't as easy to use with Lambda as it should be, but it has improved greatly.


Create API & configure it as Lambda proxy
-----------------------------------------

Go To `API gateway > Create new API`

- New API
- API Name = `AlertManager`
- Description = leave empty

In `API Gateway > AlertManager > Resources > Actions > Create resource`:

- Configure as proxy resource: `check`
- Resource name: `proxy`
- Resource path: `{proxy+}`
- Enable API Gateway CORS: leave unchecked
- `[ Create resource ]`

In integration setup:

- Type = `Lambda function proxy`
- Lambda region = choose the region your Lambda function is in
- Lambda function = `AlertManager`
- `[ Save ]`
- Add Permission to Lambda Function: OK

Now go to `Actions > Deploy API`:

- Deployment stage: `prod` (or `[New Stage]` if prod does not exist yet)
- `[ Deploy ]`


Testing if this works
---------------------

After deploying, you should now see the `Invoke URL`. Mine was `https://REDACTED.execute-api.us-west-2.amazonaws.com/prod`.

Sidenote: the `REDACTED` part is the API's ID. Currently we use that as our access control mechanism to submit alerts,
so treat the API ID as secret. We will probably implement some kind of API token authentication in the future.

Open that URL, you should see: `{"message":"Missing Authentication Token"}`.
That is to be expected. Append `/alerts` to that URL, resulting in `https://REDACTED.execute-api.us-west-2.amazonaws.com/prod/alerts`.

That URL should give you:

```
[

]
```

Meaning that there are no triggered alerts. If you get anything else, carefully read the
instructions again because something is wrong.


Submit a test alert
-------------------

Now you should try test raising an alert to the `/alerts/ingest` endpoint:

```
$ curl -H 'Content-Type: application/json' -X POST -d '{"subject": "www.example.com", "details": "I dont like the page"}' https://REDACTED.execute-api.us-west-2.amazonaws.com/prod/alerts/ingest
"OK => alert saved to database and queued for delivery"
```

Now you should hear a **bling** from your inbox, meaning that AlertManager used SNS to deliver your first alert.

Now, revisit that `/alerts` endpoint from the previous heading again - you should see:

```
[
	{
		"alert_key": "1",
		"subject": "www.example.com",
		"timestamp": "2017-01-15T12:12:04.018Z",
		"details": "I dont like the page"
	}
]
```

Now, if you try to submit the same exact alert again (repeat the `$ curl ...` command),
it should not be re-accepted (because it has the same `subject`):

```
$ curl -H 'Content-Type: application/json' -X POST -d '{"subject": "www.example.com", "details": "I dont like the page"}' https://REDACTED.execute-api.us-west-2.amazonaws.com/prod/alerts/ingest
"This alert is already firing. Discarding the submitted alert."
```

Congrats, everything seems to be working!


Acknowledging the alert
-----------------------

Okay now we learned that AlertManager won't accept alarms with the same subject again, before they are acknowledged.

How do I acknowledge the alert (after I've fixed the root cause that made the alarm go off)?

The most basic way to acknowledge the alert is to remove the row from
`DynamoDB > alertmanager_alerts > Items > (choose the alert) > Actions > Delete`.

There's also an API for doing the same:

```
$ curl -H 'Content-Type: application/json' -X POST -d '{"alert_key": "1"}' https://REDACTED.execute-api.us-west-2.amazonaws.com/prod/alerts/acknowledge
"Alert 1 deleted"
```

We use the aforementioned APIs to show/acknowledge the alerts in a central dashboard. We'll probably
open source that project in the future, but in the meantime you can use either the DynamoDB UI as your
acknowledge tool or make your own with the APIs.
