import { ActionInterface } from './actions';
import { handlerInternal } from './index';

class TestMockActions implements ActionInterface {
	alerts: Array<{ subject: string; details: string }> = [];
	logMessages: string[] = [];

	private flakyTimeoutsTargetFirstRequest = true;

	resolveTargets() {
		return Promise.resolve([
			{
				url: 'https://this-one-succeeds.com/',
				find: 'but not the body you deserve',
			},
			{ url: 'https://this-one-fails.com/', find: 'will not be found' },
			{
				url: 'https://this-one-timeouts-only-the-first-try.net/',
				find: 'this is a response',
			},
			{ url: 'https://this-one-always-timeouts.net/', find: 'foo' },
		]);
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

	postSnsAlert(subject: string, details: string) {
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

process.on('uncaughtException', () => {
	process.exit(1);
});

handlerInternal(testMockActions)
	.then(() => {
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
	})
	.catch((err) => {
		// tslint:disable-next-line:no-console
		console.error(err);
		process.exit(1);
	});
