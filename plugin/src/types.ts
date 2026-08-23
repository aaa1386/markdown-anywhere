export interface VaultConfig {
	name: string;
	path: string;
}

export interface GlobalConfig {
	version: number;
	defaultVault: VaultConfig;
	mirrorFolder: string;
	cleanupDays: number;
}

export interface CompanionStatus {
	version: string;
	configPath: string;
	configExists: boolean;
	defaultVault: VaultConfig;
	mirrorFolder: string;
	cleanupDays: number;
	registered: boolean;
	registeredCommand?: string;
}

export interface LinkInfo {
	path: string;
	target: string;
	valid: boolean;
}

export interface LinksReport {
	defaultVault: VaultConfig;
	mirrorFolder: string;
	mirrorPath: string;
	active: number;
	invalid: number;
	links: LinkInfo[];
}

export interface CleanupReport {
	removed: string[];
	count: number;
}
