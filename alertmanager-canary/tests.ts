import { APIGatewayProxyEvent, ScheduledEvent } from 'aws-lambda';
import { ActionInterface } from './actions';
import { handlerWithActions, isAPIGatewayProxyResult } from './index';
import { Config } from './types';

class TestMockActions implements ActionInterface {
	alerts: Array<{ subject: string; details: string }> = [];
	logMessages: string[] = [];

	private flakyTimeoutsTargetFirstRequest = true;

	getConfig() {
		return Promise.resolve({
			ingestSnsTopic:
				'arn:aws:sns:us-east-1:123456789123:AlertManager-ingest',
			targets: [
				{
					id: '1',
					enabled: true,
					url: 'https://this-one-succeeds.com/',
					find: 'but not the body you deserve',
				},
				{
					id: '2',
					enabled: true,
					url: 'https://this-one-fails.com/',
					find: 'will not be found',
				},
				{
					id: '3',
					enabled: true,
					url: 'https://this-one-timeouts-only-the-first-try.net/',
					find: 'this is a response',
				},
				{
					id: '4',
					enabled: true,
					url: 'https://this-one-always-timeouts.net/',
					find: 'foo',
				},
				{
					id: '5',
					enabled: false, // => won't show up in tests
					url: 'https://this-one-is-not-enabled.org/',
					find: 'doesntmatter',
				},
			],
		});
	}

	setConfig(config: Config) {
		return Promise.reject(new Error('not implemented yet'));
	}

	log(msg: string) {
		this.logMessages.push(msg);
	}

	measureDuration(started: number, ended: number) {
		// our log outputs would be nondeterministic if we allowed time to affect the
		// output. sure, we could use regex to match outputs but this feels simpler
		return 0;
	}

	httpGetBody(url: string): Promise<string> {
		switch (url) {
			case 'https://this-one-succeeds.com/':
			case 'https://this-one-fails.com/':
				return Promise.resolve(
					'the body you need, but not the body you deserve',
				);
			case 'https://this-one-always-timeouts.net/':
				return Promise.reject(new Error('Faking timeout'));
			case 'https://this-one-timeouts-only-the-first-try.net/':
				if (this.flakyTimeoutsTargetFirstRequest) {
					this.flakyTimeoutsTargetFirstRequest = false;
					return Promise.reject(new Error('Faking timeout'));
				}

				// second request for this URL, let it go through
				return Promise.resolve(
					'woop woop this is a response from a website that only works sometimes',
				);
			default:
				throw new Error(`unknown url: ${url}`);
		}
	}

	postSnsAlert(ingestTopic: string, subject: string, details: string) {
		this.alerts.push({ subject, details });

		return Promise.resolve();
	}
}

const testMockActions = new TestMockActions();

function assertEqual(actual: any, expected: any): void {
	if (actual !== expected) {
		throw new Error(`expecting<${expected}> actual<${actual}>`);
	}
}

export function mockScheduledEvent(): ScheduledEvent {
	return ({
		source: 'aws.events',
		'detail-type': 'Scheduled Event',
	} as any) as ScheduledEvent;
}

function mockProxyEvent(
	httpMethod: string,
	path: string,
	body: string,
): APIGatewayProxyEvent {
	return ({
		httpMethod,
		path,
		body,
	} as any) as APIGatewayProxyEvent;
}

async function testRestApiGetConfig() {
	const resp = await handlerWithActions(
		mockProxyEvent('GET', '/config', ''),
		testMockActions,
	);
	if (!isAPIGatewayProxyResult(resp)) {
		throw new Error('unexpected response');
	}

	assertEqual(resp.statusCode, 200);
	assertEqual(resp.headers!['Content-Type'], 'application/json');

	const config: Config = JSON.parse(resp.body);

	assertEqual(
		config.ingestSnsTopic,
		'arn:aws:sns:us-east-1:123456789123:AlertManager-ingest',
	);
	assertEqual(config.targets.length, 5);
	assertEqual(config.targets[0].url, 'https://this-one-succeeds.com/');
}

async function testRestApiPutConfig() {
	const resp = await handlerWithActions(
		mockProxyEvent('PUT', '/config', '{}'),
		testMockActions,
	);
	if (!isAPIGatewayProxyResult(resp)) {
		throw new Error('unexpected response');
	}

	assertEqual(resp.statusCode, 500);
	assertEqual(resp.body, 'setConfig failed');
}

async function testCanary() {
	await handlerWithActions(mockScheduledEvent(), testMockActions);

	const logs = testMockActions.logMessages;

	assertEqual(logs.length, 11);
	assertEqual(logs[0], '✓  https://this-one-succeeds.com/ @ 0ms => OK');
	assertEqual(
		logs[1],
		'✗  https://this-one-fails.com/ @ 0ms => find<will not be found> NOT in body<the body you need, but not the body you deserve>',
	);
	assertEqual(
		logs[2],
		'✗  https://this-one-timeouts-only-the-first-try.net/ @ 0ms => Error: Faking timeout',
	);
	assertEqual(
		logs[3],
		'✗  https://this-one-always-timeouts.net/ @ 0ms => Error: Faking timeout',
	);
	assertEqual(logs[4], '   failed; re-trying once in 1000ms');
	assertEqual(logs[5], '   failed; re-trying once in 1000ms');
	assertEqual(logs[6], '   failed; re-trying once in 1000ms');
	assertEqual(
		logs[7],
		'✗  https://this-one-fails.com/ @ 0ms => find<will not be found> NOT in body<the body you need, but not the body you deserve>',
	);
	assertEqual(
		logs[8],
		'✓  https://this-one-timeouts-only-the-first-try.net/ @ 0ms => OK',
	);
	assertEqual(
		logs[9],
		'✗  https://this-one-always-timeouts.net/ @ 0ms => Error: Faking timeout',
	);
	assertEqual(logs[10], '=> FAIL (2/4) succeeded');

	const alerts = testMockActions.alerts;

	assertEqual(alerts.length, 2);

	assertEqual(alerts[0].subject, 'https://this-one-fails.com/');
	assertEqual(
		alerts[0].details,
		'find<will not be found> NOT in body<the body you need, but not the body you deserve>',
	);

	assertEqual(alerts[1].subject, 'https://this-one-always-timeouts.net/');
	assertEqual(alerts[1].details, 'Error: Faking timeout');
}

async function runAllTests() {
	try {
		await testCanary();
		await testRestApiGetConfig();
		await testRestApiPutConfig();
	} catch (err) {
		// tslint:disable-next-line:no-console
		console.error(err);
		process.exit(1);
	}
}

// why ignore? only way to catch error is use of "await", but that can be only used from
// "async" fn, which brings a chicken-egg type problem. this is OK because this fn "cant" fail
//
// tslint:disable-next-line:no-floating-promises
runAllTests();
