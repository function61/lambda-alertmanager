export interface Target {
	url: string;
	find: string;
}

export interface TargetCheckResult {
	target: Target;
	error?: string;
	durationMs: number;
}
