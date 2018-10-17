export interface Target {
	id: string;
	enabled: boolean;
	url: string;
	find: string;
}

export interface Config {
	ingestSnsTopic: string;
	targets: Target[];
}

export interface TargetCheckResult {
	target: Target;
	error?: string;
	durationMs: number;
}
