import { App, Notice, Plugin, PluginSettingTab, Setting } from "obsidian";
import * as path from "path";

import { CompanionClient } from "./companion";
import {
	defaultGlobalConfig,
	loadGlobalConfig,
	saveGlobalConfig
} from "./global-config";
import { DEFAULT_SETTINGS, PluginSettings } from "./settings";
import { CompanionStatus, GlobalConfig, LinksReport } from "./types";

export default class MarkownAnywherePlugin extends Plugin {
	settings: PluginSettings = { ...DEFAULT_SETTINGS };
	globalConfig: GlobalConfig = defaultGlobalConfig();
	status: CompanionStatus | null = null;
	links: LinksReport | null = null;
	statusError = "";
	settingTab: MarkownAnywhereSettingTab | null = null;

	async onload(): Promise<void> {
		await this.loadSettings();
		this.globalConfig = await loadGlobalConfig();
		this.settingTab = new MarkownAnywhereSettingTab(this.app, this);
		this.addSettingTab(this.settingTab);
		void this.refreshStatus(false);
		this.addCommand({
			id: "refresh-companion-status",
			name: "Refresh Companion status",
			callback: async () => {
				await this.refreshStatus(true);
			}
		});
		this.addCommand({
			id: "cleanup-invalid-links",
			name: "Clean invalid external links",
			callback: async () => {
				await this.cleanupInvalidLinks();
			}
		});
	}

	async loadSettings(): Promise<void> {
		this.settings = Object.assign({}, DEFAULT_SETTINGS, await this.loadData());
	}

	async saveSettings(): Promise<void> {
		await this.saveData(this.settings);
	}

	getCompanionPath(): string {
		if (this.settings.companionPath.trim()) {
			return this.settings.companionPath.trim();
		}
		const vaultPath = this.getVaultPath();
		if (!vaultPath) {
			return "";
		}
		return path.join(vaultPath, this.app.vault.configDir, "plugins", this.manifest.id, "MarkownAnywhere.exe");
	}

	async setCurrentVault(): Promise<void> {
		const vaultPath = this.getVaultPath();
		if (!vaultPath) {
			throw new Error("The current Vault path is unavailable");
		}
		this.globalConfig.defaultVault = {
			name: this.app.vault.getName(),
			path: vaultPath
		};
		await saveGlobalConfig(this.globalConfig);
		new Notice("Markown Anywhere: current Vault saved as default");
		await this.refreshStatus(false);
	}

	async saveGlobalSettings(mirrorFolder: string, cleanupDays: number): Promise<void> {
		this.globalConfig.mirrorFolder = mirrorFolder;
		this.globalConfig.cleanupDays = cleanupDays;
		await saveGlobalConfig(this.globalConfig);
		new Notice("Markown Anywhere settings saved");
	}

	async refreshStatus(showNotice: boolean): Promise<void> {
		try {
			const client = new CompanionClient(this.getCompanionPath());
			this.status = await client.status();
			this.links = await client.links();
			this.statusError = "";
			if (showNotice) {
				new Notice("Markown Anywhere status refreshed");
			}
		} catch (error) {
			this.status = null;
			this.links = null;
			this.statusError = error instanceof Error ? error.message : String(error);
			if (showNotice) {
				new Notice(this.statusError);
			}
		}
		this.refreshSettingTab();
	}

	async registerCompanion(): Promise<void> {
		await new CompanionClient(this.getCompanionPath()).register();
		new Notice("Markown Anywhere file association registered");
		await this.refreshStatus(false);
	}

	async unregisterCompanion(): Promise<void> {
		await new CompanionClient(this.getCompanionPath()).unregister();
		new Notice("Markown Anywhere file association unregistered");
		await this.refreshStatus(false);
	}

	async cleanupInvalidLinks(): Promise<void> {
		const report = await new CompanionClient(this.getCompanionPath()).cleanup();
		new Notice(`Removed ${report.count} invalid link${report.count === 1 ? "" : "s"}`);
		await this.refreshStatus(false);
	}

	private getVaultPath(): string {
		const adapter = this.app.vault.adapter as { getBasePath?: () => string };
		return adapter.getBasePath?.() ?? "";
	}

	private refreshSettingTab(): void {
		this.settingTab?.display();
	}
}

class MarkownAnywhereSettingTab extends PluginSettingTab {
	constructor(app: App, private readonly plugin: MarkownAnywherePlugin) {
		super(app, plugin);
	}

		display(): void {
		const { containerEl } = this;
		containerEl.empty();
		new Setting(containerEl).setName("Markown Anywhere").setHeading();
		let mirrorFolder = this.plugin.globalConfig.mirrorFolder;
		let cleanupDays = this.plugin.globalConfig.cleanupDays;

		new Setting(containerEl)
			.setName("Default Vault")
			.setDesc(this.plugin.globalConfig.defaultVault.path || "Not configured")
			.addButton((button) =>
				button.setButtonText("Set current Vault as default").onClick(async () => {
					await this.runAction(() => this.plugin.setCurrentVault());
				})
			);

		new Setting(containerEl)
			.setName("Companion executable")
			.setDesc(this.plugin.statusError || "Path to MarkownAnywhere.exe")
			.addText((text) =>
				text
					.setPlaceholder("C:\\Path\\to\\MarkownAnywhere.exe")
					.setValue(this.plugin.settings.companionPath || this.plugin.getCompanionPath())
					.onChange(async (value) => {
						this.plugin.settings.companionPath = value.trim();
						await this.plugin.saveSettings();
					})
			)
			.addButton((button) =>
				button.setButtonText("Refresh").onClick(async () => {
					await this.runAction(() => this.plugin.refreshStatus(true));
				})
			);

		new Setting(containerEl)
			.setName("Companion status")
			.setDesc(this.plugin.status ? `Connected | ${this.plugin.status.version}` : "Not connected")
			.addButton((button) =>
				button.setButtonText("Check status").onClick(async () => {
					await this.runAction(() => this.plugin.refreshStatus(true));
				})
			);

		new Setting(containerEl)
			.setName("Windows Integration")
			.setDesc(this.plugin.status?.registered ? "Registered" : "Not registered")
			.addButton((button) =>
				button.setButtonText("Register").onClick(async () => {
					await this.runAction(() => this.plugin.registerCompanion());
				})
			)
			.addButton((button) =>
				button.setButtonText("Unregister").onClick(async () => {
					await this.runAction(() => this.plugin.unregisterCompanion());
				})
			);

		new Setting(containerEl)
			.setName("Mirror folder")
			.setDesc("Relative to the default Vault")
			.addText((text) =>
				text.setValue(mirrorFolder).onChange((value) => {
					mirrorFolder = value.trim();
				})
			)
			.addButton((button) =>
				button.setButtonText("Save").onClick(async () => {
					await this.runAction(() => this.plugin.saveGlobalSettings(mirrorFolder, cleanupDays));
				})
			);

		new Setting(containerEl)
			.setName("Cleanup interval (days)")
			.setDesc("Used by the future automatic cleanup task")
			.addText((text) =>
				text
					.setValue(String(cleanupDays))
					.onChange((value) => {
						const days = Number.parseInt(value, 10);
						if (Number.isInteger(days) && days >= 0) {
							cleanupDays = days;
						}
					})
			)
			.addButton((button) =>
				button.setButtonText("Save").onClick(async () => {
					await this.runAction(() => this.plugin.saveGlobalSettings(mirrorFolder, cleanupDays));
				})
			);

		const active = this.plugin.links?.active ?? 0;
		const invalid = this.plugin.links?.invalid ?? 0;
		new Setting(containerEl)
			.setName("Symlinks")
			.setDesc(`Active: ${active} | Invalid: ${invalid}`)
			.addButton((button) =>
				button.setButtonText("Refresh").onClick(async () => {
					await this.runAction(() => this.plugin.refreshStatus(true));
				})
			)
			.addButton((button) =>
				button.setButtonText("Clean invalid links").onClick(async () => {
					await this.runAction(() => this.plugin.cleanupInvalidLinks());
				})
			);
	}

	private async saveGlobalSettings(mirrorFolder: string, cleanupDays: number): Promise<void> {
		try {
			await this.plugin.saveGlobalSettings(mirrorFolder, cleanupDays);
		} catch (error) {
			new Notice(error instanceof Error ? error.message : String(error));
			this.display();
		}
	}

	private async runAction(action: () => Promise<void>): Promise<void> {
		try {
			await action();
			this.display();
		} catch (error) {
			new Notice(error instanceof Error ? error.message : String(error));
		}
	}
}
