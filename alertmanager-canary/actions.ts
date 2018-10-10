import * as AWS from 'aws-sdk';
import { httpGetBody } from './http';
import { Target, TargetCheckResult } from './types';

const sns = new AWS.SNS();

export interface ActionInterface {
	resolveTargets: () => Promise<Target[]>;
	httpGetBody: (url: string) => Promise<string>;
	log: (msg: string) => void;
	measureDuration: (started: number, ended: number) => number;
	postSnsAlert: (subject: string, details: string) => Promise<void>;
}

const INGEST_TOPIC = process.env.INGEST_TOPIC;

export class ProdActions implements ActionInterface {
	resolveTargets() {
		const targets: Target[] = [];

		for (const key in process.env) {
			if (/^CHECK[0-9]+$/.test(key)) {
				const target = JSON.parse(process.env[key]!);
				if (!target) {
					return Promise.reject(
						new Error(
							`failed to parse target ${key}: ${process.env[key]}`
						)
					);
				}

				targets.push(target);
			}
		}

		return Promise.resolve(targets);
	}

	log(msg: string) {
		console.log(msg);
	}

	measureDuration(started: number, ended: number) {
		return started - ended;
	}

	httpGetBody(url: string) {
		return httpGetBody(url);
	}

	postSnsAlert(subject: string, details: string) {
		return new Promise<void>((resolve, reject) => {
			if (!INGEST_TOPIC) {
				reject(new Error('INGEST_TOPIC not defined'));
				return;
			}

			const ingestTopic: string = INGEST_TOPIC;

			sns.publish(
				{
					TopicArn: ingestTopic,
					Subject: subject,
					Message: details,
				},
				(err: Error) => {
					if (err) {
						reject(err);
					} else {
						resolve();
					}
				}
			);
		});
	}
}
