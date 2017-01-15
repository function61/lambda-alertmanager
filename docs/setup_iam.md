Setup IAM
=========

"IAM" takes care of access management.


Create role
-----------

Go to `IAM > Roles > Create new role`:

- Name = `AlertManager`.
- Role type = `AWS Service Roles > AWS Lambda`.
- Do no attach any policies (= just hit next) as we'll set up super restrictive custom policy.
- `[ Create role ]`


Attach policy to role
---------------------

Now go to `IAM > Roles > AlertManager > Inline policies > Create > Custom policy`:

- Policy name = `dynamodbAlertsPlusSnsAlertAndIngest`.

Content will be below, but you should copy it to a text editor first, and replace `__ACCOUNT_ID__` with your AWS account ID. It looks like `426466625513`.

```
{
    "Version": "2012-10-17",
    "Statement": [
    	{
            "Sid": "",
            "Effect": "Allow",
            "Action": [
                "dynamodb:PutItem",
                "dynamodb:DeleteItem",
                "dynamodb:Scan"
            ],
            "Resource": [
                "arn:aws:dynamodb:*:__ACCOUNT_ID__:table/alertmanager_alerts"
            ]
    	},
    	{
            "Sid": "",
            "Effect": "Allow",
            "Action": [
	            "dynamodb:GetRecords",
	            "dynamodb:GetShardIterator",
	            "dynamodb:DescribeStream",
	            "dynamodb:ListStreams"
            ],
            "Resource": [
                "arn:aws:dynamodb:*:__ACCOUNT_ID__:table/alertmanager_alerts/stream/*"
            ]
    	},
        {
            "Sid": "",
            "Effect": "Allow",
            "Action": [
                "sns:Publish"
            ],
            "Resource": [
                "arn:aws:sns:*:__ACCOUNT_ID__:AlertManager-alert",
                "arn:aws:sns:*:__ACCOUNT_ID__:AlertManager-ingest"
            ]
        },
        {
            "Sid": "",
            "Resource": "*",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Effect": "Allow"
        }
    ]
}
```
