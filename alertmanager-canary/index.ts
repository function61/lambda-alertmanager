import {
	APIGatewayProxyEvent,
	APIGatewayProxyResult,
	Handler,
	ScheduledEvent,
} from 'aws-lambda';
import { ActionInterface, ProdActions } from './actions';
import { Target, TargetCheckResult } from './types';

// wraps checkOneInternal() and adds retry capability
function checkOneWithRetries(
	target: Target,
	actions: ActionInterface,
): Promise<TargetCheckResult> {
	return new Promise((resolve, reject) => {
		checkOneInternal(target, actions, false).then((result) => {
			if (result.error !== undefined) {
				const retryInMs = 1000;
				actions.log(`   failed; re-trying once in ${retryInMs}ms`);

				setTimeout(() => {
					// this time, wire outcome directly to resolve
					checkOneInternal(target, actions, true).then(
						resolve,
						reject,
					);
				}, retryInMs);
			} else {
				resolve(result);
			}
		}, reject);
	});
}

function checkOneInternal(
	target: Target,
	actions: ActionInterface,
	finalTry: boolean,
): Promise<TargetCheckResult> {
	return new Promise((resolve, reject) => {
		const next = (result: TargetCheckResult) => {
			const logMsgSucceededSign = result.error !== undefined ? '✗' : '✓';
			const logMsgDetails =
				result.error !== undefined
					? truncate(oneLinerize(result.error), 128)
					: 'OK';
			const logMsg = `${logMsgSucceededSign}  ${target.url} @ ${
				result.durationMs
			}ms => ${logMsgDetails}`;

			actions.log(logMsg);

			if (result.error !== undefined && finalTry) {
				actions.postSnsAlert(target.url, result.error).then(
					() => {
						resolve(result);
					},
					(err: Error) => {
						// failure posting alert
						reject(err);
					},
				);
			} else {
				resolve(result);
			}
		};

		const timeStarted = now();

		actions.httpGetBody(target.url).then(
			(body) => {
				const durationMs = actions.measureDuration(now(), timeStarted);

				let failure: string | undefined;

				if (body.indexOf(target.find) === -1) {
					failure = `find<${target.find}> NOT in body<${body}>`;
				}

				next({ target, durationMs, error: failure });
			},
			(err: Error) => {
				const durationMs = actions.measureDuration(now(), timeStarted);

				next({ target, durationMs, error: err.toString() });
			},
		);
	});
}

function handleCanary(actions: ActionInterface): Promise<string> {
	return actions.resolveTargets().then((targets) => {
		// runs all checks in parallel
		const allChecksPromises: Array<
			Promise<TargetCheckResult>
		> = targets.map((target) => checkOneWithRetries(target, actions));

		return Promise.all(allChecksPromises).then((allTargetCheckResults) => {
			const numFailed = allTargetCheckResults.filter(
				(check) => check.error !== undefined,
			).length;
			const numTotal = allTargetCheckResults.length;
			const numSucceeded = numTotal - numFailed;

			if (numFailed > 0) {
				actions.log(
					'=> FAIL (' + numSucceeded + '/' + numTotal + ') succeeded',
				);
			} else {
				actions.log('=> All passed. Awesome!');
			}

			// always resolve even if any check fails, because semantically, Canary has
			// done its job succesfully
			return 'Canary finished';
		});
	});
}

function handleProxyEvent(
	event: APIGatewayProxyEvent,
): Promise<APIGatewayProxyResult> {
	return Promise.resolve({
		statusCode: 200,
		headers: {
			'Content-Type': 'application/json',
		},
		body: `you requested ${event.path}`,
	});
}

// exported for testing purposes
export function handlerWithActions(
	event: ScheduledEvent | APIGatewayProxyEvent,
	actions: ActionInterface,
): Promise<string | APIGatewayProxyResult> {
	// "multiplexed" handler => recognize format of incoming event. this is really ugly.
	if (isScheduledEvent(event)) {
		return handleCanary(actions);
	} else if (isProxyEvent(event)) {
		return handleProxyEvent(event);
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

function oneLinerize(input: string): string {
	return input.replace(/\n/g, '\\n');
}

function truncate(input: string, to: number): string {
	return to >= input.length ? input : input.substr(0, to - 2) + '..';
}

function now(): number {
	// we could use performance.now() for sub-millisecond measurements,
	// but for network I/O this precision is sufficient
	return new Date().getTime();
}

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
