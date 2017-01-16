Setup IAM
=========

"IAM" takes care of access management. For security we'll restrict AlertManager's access to
the bare minimum it needs to operate under.


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
                "arn:aws:dynamodb:*:*:table/alertmanager_alerts"
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
                "arn:aws:dynamodb:*:*:table/alertmanager_alerts/stream/*"
            ]
    	},
        {
            "Sid": "",
            "Effect": "Allow",
            "Action": [
                "sns:Publish"
            ],
            "Resource": [
                "arn:aws:sns:*:*:AlertManager-alert",
                "arn:aws:sns:*:*:AlertManager-ingest"
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

Are the wildcards safe?
-----------------------

Yes. I used wildcards so you can just copy-paste the policy from above without needing to do region and
account id replacements (the `*:*` parts). It is acceptable to have wildcards for:

- Region component: gives additional access only to table with same name (alertmanager_alerts)
  in other regions (you won't have same table name in other regions) or SNS topics with same
  names in other regions (you won't have same topic names in other regions).
- Account id component: gives AlertManager additional access to resources in other accounts you have access to: **none**,
  as how could you give yourself access to other accounts' resources?

If you're unsure of this in any capacity, feel free to plug in your region and account IDs in the resource constraints.
