import { execFile } from "child_process";
import { promisify } from "util";

import { CleanupReport, CompanionStatus, LinksReport } from "./types";

const execFileAsync = promisify(execFile) as (
	file: string,
	args: string[],
	options: { windowsHide: boolean; maxBuffer: number }
) => Promise<{ stdout: string; stderr: string }>;

export class CompanionClient {
	constructor(private readonly executablePath: string) {}

	async status(): Promise<CompanionStatus> {
		return this.runJSON<CompanionStatus>(["--status", "--json"]);
	}

	async links(): Promise<LinksReport> {
		return this.runJSON<LinksReport>(["--links", "--json"]);
	}

	async cleanup(): Promise<CleanupReport> {
		return this.runJSON<CleanupReport>(["--cleanup"]);
	}

	async register(): Promise<void> {
		await this.run(["--register"]);
	}

	async unregister(): Promise<void> {
		await this.run(["--unregister"]);
	}

	private async runJSON<T>(args: string[]): Promise<T> {
		const output = await this.run(args);
		try {
			return JSON.parse(output) as T;
		} catch (error) {
			throw new Error(`Companion returned invalid JSON: ${error instanceof Error ? error.message : String(error)}`);
		}
	}

	private async run(args: string[]): Promise<string> {
		if (!this.executablePath.trim()) {
			throw new Error("Companion executable path is not configured");
		}
		try {
			const result = await execFileAsync(this.executablePath, args, {
				windowsHide: true,
				maxBuffer: 1024 * 1024
			});
			if (result.stderr.trim()) {
				throw new Error(result.stderr.trim());
			}
			return result.stdout.trim();
		} catch (error) {
			if (error instanceof Error) {
				throw new Error(`Companion command failed: ${error.message}`);
			}
			throw new Error(`Companion command failed: ${String(error)}`);
		}
	}
}
