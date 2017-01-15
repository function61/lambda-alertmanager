lambda-alertmanager installation
--------------------------------

Take note of your AWS region. These docs assume you are in the `us-west-2` region.
If not, substitute your region code everywhere in these docs!

Follow these steps precisely, and you've got yourself a working installation:

1. [Set up SNS topics](./setup_sns.md)
2. [Set up DynamoDB](./setup_dynamodb.md)
3. [Set up IAM](./setup_iam.md)
4. [Set up AlertManager](./setup_alertmanager.md)
5. [Set up API Gateway](./setup_apigwateway.md) (also includes: testing that this works)
6. (recommended) [Set up AlertManager-canary](./setup_alertmanager-canary.md)
7. (optional) Set up Prometheus integration
8. (optional) Set up custom integration
