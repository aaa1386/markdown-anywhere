import { promises as fs } from "fs";
import * as os from "os";
import * as path from "path";

import { GlobalConfig } from "./types";

export const CURRENT_VERSION = 1;
export const DEFAULT_MIRROR_FOLDER = "_external-open";
export const DEFAULT_CLEANUP_DAYS = 7;

export function defaultGlobalConfig(): GlobalConfig {
	return {
		version: CURRENT_VERSION,
		defaultVault: { name: "", path: "" },
		mirrorFolder: DEFAULT_MIRROR_FOLDER,
		cleanupDays: DEFAULT_CLEANUP_DAYS
	};
}

export function globalConfigPath(): string {
	const appData = process.env.APPDATA || path.join(os.homedir(), "AppData", "Roaming");
	return path.join(appData, "MarkownAnywhere", "config.json");
}

export async function loadGlobalConfig(): Promise<GlobalConfig> {
	try {
		const raw = await fs.readFile(globalConfigPath(), "utf8");
		const parsed = JSON.parse(raw) as Partial<GlobalConfig>;
		return normalizeConfig(parsed);
	} catch (error) {
		if (isMissingFile(error)) {
			return defaultGlobalConfig();
		}
		throw error;
	}
}

export async function saveGlobalConfig(config: GlobalConfig): Promise<void> {
	const normalized = normalizeConfig(config);
	validateConfig(normalized);
	const filePath = globalConfigPath();
	await fs.mkdir(path.dirname(filePath), { recursive: true });
	await fs.writeFile(filePath, `${JSON.stringify(normalized, null, 2)}\n`, "utf8");
}

function normalizeConfig(config: Partial<GlobalConfig>): GlobalConfig {
	const defaults = defaultGlobalConfig();
	return {
		version: config.version || defaults.version,
		defaultVault: {
			name: config.defaultVault?.name || "",
			path: config.defaultVault?.path || ""
		},
		mirrorFolder: config.mirrorFolder || defaults.mirrorFolder,
		cleanupDays: typeof config.cleanupDays === "number" ? config.cleanupDays : defaults.cleanupDays
	};
}

function validateConfig(config: GlobalConfig): void {
	if (config.version !== CURRENT_VERSION) {
		throw new Error(`Unsupported config version: ${config.version}`);
	}
	if (!config.mirrorFolder || path.isAbsolute(config.mirrorFolder)) {
		throw new Error("Mirror folder must be a non-empty relative path");
	}
	const clean = path.normalize(config.mirrorFolder);
	if (clean === "." || clean === ".." || clean.startsWith(`..${path.sep}`)) {
		throw new Error("Mirror folder must stay inside the Vault");
	}
	if (!Number.isInteger(config.cleanupDays) || config.cleanupDays < 0) {
		throw new Error("Cleanup days must be a non-negative integer");
	}
}

function isMissingFile(error: unknown): boolean {
	return typeof error === "object" && error !== null && "code" in error && error.code === "ENOENT";
}
