import { APIGatewayProxyEvent, APIGatewayProxyResult } from 'aws-lambda';
import { ActionInterface } from './actions';
import { Config } from './types';

async function handleGetConfig(
	actions: ActionInterface,
): Promise<APIGatewayProxyResult> {
	// TODO: try-catch
	const config = await actions.getConfig();

	return http200('application/json', JSON.stringify(config));
}

async function handlePutConfig(
	event: APIGatewayProxyEvent,
	actions: ActionInterface,
): Promise<APIGatewayProxyResult> {
	if (!event.body) {
		return http400('body not found');
	}

	const config: Config = JSON.parse(event.body);

	try {
		await actions.setConfig(config);
	} catch {
		return http500('setConfig failed');
	}

	return http200('text/plain', 'OK');
}

export async function handleRestApi(
	event: APIGatewayProxyEvent,
	actions: ActionInterface,
): Promise<APIGatewayProxyResult> {
	const endpoint = `${event.httpMethod} ${event.path}`;

	switch (endpoint) {
		case 'GET /config':
			return handleGetConfig(actions);
		case 'PUT /config':
			return handlePutConfig(event, actions);
		default:
			throw new Error(`Unknown endpoint: ${endpoint}`);
	}
}

function http200(contentType: string, body: string): APIGatewayProxyResult {
	return {
		statusCode: 200,
		headers: {
			'Content-Type': contentType,
		},
		body,
	};
}

function http400(body: string): APIGatewayProxyResult {
	return {
		statusCode: 400,
		headers: {
			'Content-Type': 'text/plain',
		},
		body,
	};
}

function http500(body: string): APIGatewayProxyResult {
	return {
		statusCode: 500,
		headers: {
			'Content-Type': 'text/plain',
		},
		body,
	};
}
