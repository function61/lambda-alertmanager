export interface Monitor {
	id: string;
	enabled: boolean;
	url: string;
	find: string;
}

export interface Config {
	sns_topic_ingest: string;
	monitors: Monitor[];
}

export interface MonitorCheckResult {
	monitor: Monitor;
	error?: string;
	durationMs: number;
}
