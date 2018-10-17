import * as AWS from 'aws-sdk';
import { httpGetBody } from './http';
import { Config } from './types';

const sns = new AWS.SNS({ apiVersion: '2010-03-31' });
const s3 = new AWS.S3({ apiVersion: '2006-03-01' });

export interface ActionInterface {
	getConfig: () => Promise<Config>;
	setConfig: (config: Config) => Promise<void>;
	httpGetBody: (url: string) => Promise<string>;
	log: (msg: string) => void;
	measureDuration: (started: number, ended: number) => number;
	postSnsAlert: (
		ingestTopic: string,
		subject: string,
		details: string,
	) => Promise<void>;
}

const S3_BUCKET = process.env.S3_BUCKET;

const targetsJsonKey = 'targets.json';

export class ProdActions implements ActionInterface {
	getConfig() {
		return new Promise<Config>((resolve, reject) => {
			if (!S3_BUCKET) {
				reject(new Error('S3_BUCKET not set'));
				return;
			}

			s3.getObject(
				{
					Bucket: S3_BUCKET,
					Key: targetsJsonKey,
				},
				(err, resp) => {
					if (err) {
						reject(err);
						return;
					}

					// can't believe the API can't explicitly specify what the hell it returns
					if (!(resp.Body instanceof Buffer)) {
						reject(new Error('Unexpected S3 response body type'));
						return;
					}

					const config: Config = JSON.parse(resp.Body.toString());

					resolve(config);
				},
			);
		});
	}

	setConfig(config: Config) {
		return new Promise<void>((resolve, reject) => {
			if (!S3_BUCKET) {
				reject(new Error('S3_BUCKET not set'));
				return;
			}

			s3.putObject(
				{
					Body: JSON.stringify(config),
					Bucket: S3_BUCKET,
					Key: targetsJsonKey,
					ContentType: 'application/json',
				},
				(err, data) => {
					if (err) {
						reject(err);
						return;
					}

					resolve();
				},
			);
		});
	}

	log(msg: string) {
		// tslint:disable-next-line:no-console
		console.log(msg);
	}

	measureDuration(started: number, ended: number) {
		return started - ended;
	}

	httpGetBody(url: string) {
		return httpGetBody(url);
	}

	postSnsAlert(ingestTopic: string, subject: string, details: string) {
		return new Promise<void>((resolve, reject) => {
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
				},
			);
		});
	}
}
