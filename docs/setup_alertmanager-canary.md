Setting up AlertManager-Canary
==============================

Create the Lambda function
--------------------------

- Go to `Lambda > Create a Lambda function > Blank function`.
- Name: `AlertManager-canary`
- Description: `Checks that important web properties are working.`
- Runtime: `Node.js 4.3` (or higher)
- Code entry type: `Upload a .ZIP file`
- Download this to your desktop and then upload to Lambda: TODO
- FIXME_REMOVE: S3 link URL: `https://s3.amazonaws.com/files.function61.com/alertmanager-canary/alertmanager-canary-2017-01-14.zip`
- Enable encryption helpers: leave unchecked

Now, for each property that you want to monitor, add those checks as separate ENV variables. Example:

- `CHECK1` = `{"url":"https://example.com/"§"find":"This domain is established to be used for illustrative examples in documents."}`
- `INGEST_TOPIC` = `arn:aws:sns:us-west-2:426466625513:AlertManager-ingest` (replace your region and customer ID)

(there can be gaps in the check numbers, the numbers only have to be unique - luckily Lambda checks this)

Role config:

- Handler: leave as is (`index.handler`)
- Role: leave as is (`Choose existing role`)
- Existing role: `AlertManager`

Advanced config:

- Memory (MB): leave as is (`128`)
- Timeout: 1 min

Okay now hit `[ Create function ]`.


Test that the function works
----------------------------

Now hit `[ Test ]` so we can see that it is working. It'll ask you for a test event, but the content does not matter
(since our events will be schedule-based) so just accept the dummy event offered by Lambda.

You should get this log output from the test run:

```
START RequestId: ff8ffe53-db1d-11e6-8fda-35d4c5ac1dd6 Version: $LATEST
2017-01-15T12:27:37.417Z	ff8ffe53-db1d-11e6-8fda-35d4c5ac1dd6	Starting Canary. Check count: 1
2017-01-15T12:27:37.838Z	ff8ffe53-db1d-11e6-8fda-35d4c5ac1dd6	✓ https://example.com/ duration=419
2017-01-15T12:27:37.838Z	ff8ffe53-db1d-11e6-8fda-35d4c5ac1dd6	=> All passed. Awesome!
END RequestId: ff8ffe53-db1d-11e6-8fda-35d4c5ac1dd6
```

Now edit the check definition (`CHECK1`) to look like this:

```
{"url":"https://example.com/"§"find":"THIS TEXT WILL NOT BE FOUND"}
```

- `[ Save ]`
- `[ Test ]`

Your log output should now be:

```
START RequestId: 586059c3-db1e-11e6-ab0a-a37bca6277f6 Version: $LATEST
2017-01-15T12:30:06.403Z	586059c3-db1e-11e6-ab0a-a37bca6277f6	Starting Canary. Check count: 1
2017-01-15T12:30:06.837Z	586059c3-db1e-11e6-ab0a-a37bca6277f6	https://example.com/ failed once - re-trying (only once)
2017-01-15T12:30:06.948Z	586059c3-db1e-11e6-ab0a-a37bca6277f6	✗ https://example.com/ => find="THIS TEXT WILL NOT BE FOUND" NOT in body: <!doctype html..
2017-01-15T12:30:07.439Z	586059c3-db1e-11e6-ab0a-a37bca6277f6	=> FAIL (0/1) succeeded
END RequestId: 586059c3-db1e-11e6-ab0a-a37bca6277f6
```

AlertManager-Canary just posted this alarm to AlertManager for ingestion via SNS topic `AlertManager-ingest`.

You should've received the alert via email. Now if you hit `[ Test ]` again, Canary will submit the alarm again for ingestion,
but this time it will be discarded because the previous alarm for the same URL is not acknowledged yet.

You can now acknowledge the alert you just triggered (read the API gateway docs again if you are not sure how to do this),
and add actual websites to monitor to your Canary. Please don't leave the example.com check there, as it's not your website to hammer.


Add scheduled trigger
---------------------

We want this Canary to be ran automatically every minute (or any rate you want).

Go to `CloudWatch > Events > Rules > Create`:

- Event source = `Schedule`
- Fixed rate = `1 minutes`

In `Targets > Add target`:

- Lambda function = `AlertManager-Canary`

Hit `[ Configure details ]` ("next"):

- Name = `AlertManager-Canary`
- Description = leave empty
- State = `enabled`
- `[ Create rule ]`

Canary will not be run automatically - every minute. You can verify it works either by:

- Looking at the logs in `Lambda > AlertManager-Canary > Monitoring > Logs` or
- Tweaking the check definitions in a way that they'll trigger an alarm and wait a minute
  to receive the alarm so you know it's working. Just remember to tweak the check back to
  how it should be and acknowledge the alert!
