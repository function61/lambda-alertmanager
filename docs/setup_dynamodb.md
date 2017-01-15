Setting up DynamoDB
===================

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
