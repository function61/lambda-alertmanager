import {
	APIGatewayProxyEvent,
	APIGatewayProxyResult,
	Handler,
	ScheduledEvent,
} from 'aws-lambda';
import { ActionInterface, ProdActions } from './actions';
import { handleCanary } from './canary';
import { handleRestApi } from './restapi';

// exported for testing purposes
export function handlerWithActions(
	event: ScheduledEvent | APIGatewayProxyEvent,
	actions: ActionInterface,
): Promise<string | APIGatewayProxyResult> {
	// "multiplexed" handler => recognize format of incoming event. this is really ugly.
	if (isScheduledEvent(event)) {
		return handleCanary(actions);
	} else if (isProxyEvent(event)) {
		return handleRestApi(event, actions);
	} else {
		return Promise.reject('unknown event');
	}
}

const prodActions = new ProdActions();

export const handler: Handler<
	ScheduledEvent | APIGatewayProxyEvent,
	string | APIGatewayProxyResult
> = (event) => {
	return handlerWithActions(event, prodActions);
};

function isScheduledEvent(input: any): input is ScheduledEvent {
	if (!('source' in input) || input.source !== 'aws.events') {
		return false;
	}

	if (
		!('detail-type' in input) ||
		input['detail-type'] !== 'Scheduled Event'
	) {
		return false;
	}

	return true;
}

function isProxyEvent(input: any): input is APIGatewayProxyEvent {
	return 'httpMethod' in input && 'path' in input;
}

export function isAPIGatewayProxyResult(
	input: any,
): input is APIGatewayProxyResult {
	return 'statusCode' in input && 'headers' in input;
}
