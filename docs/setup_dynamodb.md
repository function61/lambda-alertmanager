Setting up DynamoDB
===================

DynamoDB holds all the state that AlertManager keeps. The state is required to implement:

- Acknowledgements: mark an alert as having been handled.
- Alarm suppression: if alert is not yet acknowledged as handled, the same alert should not be sent twice.
- Rate limiting: if too many alarms are triggering, do not overwhelm operations' inboxes & phone with alerts.


Create table
------------

Create table:

- Name = `alertmanager_alerts`
- Primary key = `alert_key` (type=string)
- Use default settings
- `[ Create ]`


Enable stream
-------------

Now we need to enable stream for that table, so our Lambda function can listen to changes in this table.

From `alertmanager_alerts > Overview > Stream details > Manage stream`:

- View type = `New image`
- `[ Enable ]`
