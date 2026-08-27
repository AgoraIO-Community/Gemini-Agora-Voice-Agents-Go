"use client";

export type QuickstartAgentMetric = {
	type: string;
	module?: string;
	name: string;
	value: number;
	timestamp: number;
};

type QuickstartPipelineMetricsProps = {
	metrics: QuickstartAgentMetric[];
};

const PIPELINE = [
	{ key: "stt", label: "Gemini ASR", metricTypes: ["stt", "asr"] },
	{ key: "llm", label: "Gemini LLM", metricTypes: ["llm", "mllm"] },
	{ key: "tts", label: "MiniMax TTS", metricTypes: ["tts"] },
] as const;

function formatMetricName(name: string) {
	return name.replace(/[_-]+/g, " ");
}

export function QuickstartPipelineMetrics({
	metrics,
}: QuickstartPipelineMetricsProps) {
	const latestByType = new Map<string, QuickstartAgentMetric>();
	for (const metric of metrics) {
		const metricType = metric.type || metric.module;
		if (metricType) {
			latestByType.set(metricType.toLowerCase(), metric);
		}
	}

	return (
		<div className="flex min-w-0 flex-wrap items-center gap-2">
			<span className="text-sm font-medium leading-6 text-muted-foreground">
				Pipeline
			</span>
			{PIPELINE.map((step, index) => {
				const metric = step.metricTypes
					.map((type) => latestByType.get(type))
					.find(Boolean);

				return (
					<div key={step.key} className="flex items-center gap-2">
						{index > 0 ? (
							<span
								className="text-xs text-muted-foreground"
								aria-hidden="true"
							>
								/
							</span>
						) : null}
				<span className="inline-flex items-center gap-2 rounded-md border border-border bg-transparent px-2 py-0.5 text-xs font-semibold leading-4 text-foreground shadow-sm">
							{step.label}
							{metric ? (
								<span
									className="text-primary"
									title={new Date(metric.timestamp).toLocaleTimeString()}
								>
									{formatMetricName(metric.name)} {Math.round(metric.value)}ms
								</span>
							) : null}
						</span>
					</div>
				);
			})}
		</div>
	);
}
