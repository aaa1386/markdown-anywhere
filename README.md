# Markown Anywhere

Open Markdown files from anywhere with Obsidian on Windows.

Markown Anywhere has two parts:

- `companion/` contains the Windows Companion written in Go. It handles file associations, Vault routing, symlink creation, Obsidian URI launching, and link cleanup.
- `plugin/` contains the desktop-only Obsidian plugin. It manages settings and calls the Companion CLI; it is not part of the Markdown opening hot path.

The project does not copy Markdown contents, upload files, or collect telemetry. External files are represented inside the Vault by symlinks under `_external-open/` by default.

## Local development

Build and test the Companion:

```powershell
cd companion
go test ./...
go vet ./...
go build -buildvcs=false -ldflags="-H=windowsgui" -o .\dist\MarkownAnywhere.exe .\cmd\markown-anywhere
```

Build the Obsidian plugin:

```powershell
cd plugin
npm ci
npm run build
```

For a local install, copy `plugin/main.js`, `plugin/manifest.json`, and `plugin/versions.json` to `<Vault>/.obsidian/plugins/markown-anywhere/`. Copy `companion/dist/MarkownAnywhere.exe` to the same folder, or set its absolute path in the plugin settings.

## Configuration

The Companion and plugin share this file:

```text
%APPDATA%\MarkownAnywhere\config.json
```

The plugin can set the current Vault as the default, register or unregister the Windows Markdown handler, inspect active/invalid symlinks, and clean invalid links.

## BRAT beta testing

Install BRAT from Obsidian Community plugins, then add this repository:

```text
zhangcongke/markdown-anywhere
```

BRAT installs `main.js` and `manifest.json` from GitHub Release assets. The Windows Companion is an additional executable and must be downloaded and configured separately.

## Releases

Update `plugin/manifest.json` and `plugin/versions.json`, commit the change, then push a semantic-version tag such as `0.1.0` or `0.1.1-beta.1`:

```powershell
git tag 0.1.0
git push origin 0.1.0
```

GitHub Actions validates the version, builds both components, and publishes the plugin files, Windows executable, and a complete Windows ZIP to the GitHub Release.
