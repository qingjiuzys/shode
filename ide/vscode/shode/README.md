# Shode VSCode Extension

This directory contains the VSCode integration for the Shode scripting language. The extension provides:

- Syntax highlighting & language configuration
- Language Server (LSP) with completion, hover & diagnostics for stdlib/builtins
- Commands for running `shode` scripts directly from VSCode

## Structure

```
ide/vscode/shode
├── package.json              # Extension manifest
├── tsconfig.json             # Root TS config
├── client/                   # VSCode client (activates extension)
│   ├── src/extension.ts
│   └── tsconfig.json
├── server/                   # LSP server implementation
│   ├── src/server.ts
│   └── tsconfig.json
├── syntaxes/shode.tmLanguage.json
└── language-configuration.json
```

## Development

```bash
cd ide/vscode/shode
npm install
npm run compile
code .
```

Press `F5` in VSCode to launch the extension host.

## Commands

- `shode.runScript`: runs the current file via `shode run`
- `shode.execSelection`: executes selected text via `shode exec`

Both commands rely on the `shode` binary being available on your `PATH`.
