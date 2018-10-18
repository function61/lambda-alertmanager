import { ActionInterface } from './actions';
import { Config, Monitor, MonitorCheckResult } from './types';

// wraps checkOneInternal() and adds retry capability
function checkOneWithRetries(
	monitor: Monitor,
	config: Config,
	actions: ActionInterface,
): Promise<MonitorCheckResult> {
	return new Promise((resolve, reject) => {
		checkOneInternal(monitor, config, actions, false).then((result) => {
			if (result.error !== undefined) {
				const retryInMs = 1000;
				actions.log(`   failed; re-trying once in ${retryInMs}ms`);

				setTimeout(() => {
					// this time, wire outcome directly to resolve
					checkOneInternal(monitor, config, actions, true).then(
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
	monitor: Monitor,
	config: Config,
	actions: ActionInterface,
	finalTry: boolean,
): Promise<MonitorCheckResult> {
	return new Promise((resolve, reject) => {
		const next = (result: MonitorCheckResult) => {
			const logMsgSucceededSign = result.error !== undefined ? '✗' : '✓';
			const logMsgDetails =
				result.error !== undefined
					? truncate(oneLinerize(result.error), 128)
					: 'OK';
			const logMsg = `${logMsgSucceededSign}  ${monitor.url} @ ${
				result.durationMs
			}ms => ${logMsgDetails}`;

			actions.log(logMsg);

			if (result.error !== undefined && finalTry) {
				actions
					.postSnsAlert(
						config.sns_topic_ingest,
						monitor.url,
						result.error,
					)
					.then(
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

		actions.httpGetBody(monitor.url).then(
			(body) => {
				const durationMs = actions.measureDuration(now(), timeStarted);

				let failure: string | undefined;

				if (body.indexOf(monitor.find) === -1) {
					failure = `find<${monitor.find}> NOT in body<${body}>`;
				}

				next({ monitor, durationMs, error: failure });
			},
			(err: Error) => {
				const durationMs = actions.measureDuration(now(), timeStarted);

				next({ monitor, durationMs, error: err.toString() });
			},
		);
	});
}

export function handleCanary(actions: ActionInterface): Promise<string> {
	return actions.getConfig().then((config) => {
		// runs all checks in parallel
		const allChecksPromises: Array<
			Promise<MonitorCheckResult>
		> = config.monitors
			.filter(isEnabled)
			.map((monitor) => checkOneWithRetries(monitor, config, actions));

		return Promise.all(allChecksPromises).then((allMonitorCheckResults) => {
			const numFailed = allMonitorCheckResults.filter(
				(check) => check.error !== undefined,
			).length;
			const numTotal = allMonitorCheckResults.length;
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

const isEnabled = (monitor: Monitor) => monitor.enabled;

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
