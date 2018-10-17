import { APIGatewayProxyEvent, APIGatewayProxyResult } from 'aws-lambda';
import { ActionInterface } from './actions';

async function handleGetTargets(
	actions: ActionInterface,
): Promise<APIGatewayProxyResult> {
	const targets = await actions.getTargets();

	return {
		statusCode: 200,
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(targets),
	};
}

export async function handleRestApi(
	event: APIGatewayProxyEvent,
	actions: ActionInterface,
): Promise<APIGatewayProxyResult> {
	const endpoint = `${event.httpMethod} ${event.path}`;

	switch (endpoint) {
		case 'GET /targets':
			return handleGetTargets(actions);
		default:
			throw new Error(`Unknown endpoint: ${endpoint}`);
	}
}
