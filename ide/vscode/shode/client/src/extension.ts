import * as path from "path";
import * as vscode from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind,
} from "vscode-languageclient/node";

let client: LanguageClient | undefined;

class ShodeDebugConfigurationProvider
  implements vscode.DebugConfigurationProvider
{
  resolveDebugConfiguration(
    _folder: vscode.WorkspaceFolder | undefined,
    config: vscode.DebugConfiguration,
  ): vscode.DebugConfiguration | undefined {
    if (!config.type) {
      config.type = "shode";
    }
    if (!config.request) {
      config.request = "launch";
    }
    if (!config.name) {
      config.name = "Shode: Launch Script";
    }
    if (!config.program && vscode.window.activeTextEditor) {
      config.program = vscode.window.activeTextEditor.document.fileName;
    }

    if (!config.program) {
      void vscode.window.showErrorMessage("Missing Shode program to debug.");
      return undefined;
    }

    return config;
  }
}

class ShodeDebugAdapterFactory
  implements vscode.DebugAdapterDescriptorFactory
{
  createDebugAdapterDescriptor(
    _session: vscode.DebugSession,
  ): vscode.ProviderResult<vscode.DebugAdapterDescriptor> {
    return new vscode.DebugAdapterExecutable("shode", ["debug-adapter"]);
  }

  dispose(): void {
    // noop
  }
}

export async function activate(context: vscode.ExtensionContext) {
  const serverModule = context.asAbsolutePath(
    path.join("dist", "server", "server.js"),
  );

  const serverOptions: ServerOptions = {
    run: { module: serverModule, transport: TransportKind.ipc },
    debug: {
      module: serverModule,
      transport: TransportKind.ipc,
      options: { execArgv: ["--nolazy", "--inspect=6009"] },
    },
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ language: "shode" }],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher("**/*.shode"),
    },
  };

  client = new LanguageClient(
    "shodeLanguageServer",
    "Shode Language Server",
    serverOptions,
    clientOptions,
  );

  context.subscriptions.push(client.start());

  const debugProvider = new ShodeDebugConfigurationProvider();
  context.subscriptions.push(
    vscode.debug.registerDebugConfigurationProvider("shode", debugProvider),
  );

  const adapterFactory = new ShodeDebugAdapterFactory();
  context.subscriptions.push(
    vscode.debug.registerDebugAdapterDescriptorFactory("shode", adapterFactory),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("shode.runScript", async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor) {
        return;
      }
      await editor.document.save();
      await runShode(["run", editor.document.fileName]);
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("shode.execSelection", async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor) {
        return;
      }
      const selection = editor.selection;
      const text = selection.isEmpty
        ? editor.document.lineAt(selection.start.line).text
        : editor.document.getText(selection);

      if (!text.trim()) {
        vscode.window.showInformationMessage("No Shode code selected.");
        return;
      }

      await runShode(["exec", text]);
    }),
  );
}

export async function deactivate(): Promise<void> {
  if (!client) {
    return;
  }
  await client.stop();
}

async function runShode(args: string[]) {
  const terminal =
    vscode.window.terminals.find((term) => term.name === "Shode") ??
    vscode.window.createTerminal("Shode");

  terminal.show();
  const command = ["shode", ...args]
    .map((segment) => {
      if (segment.includes(" ")) {
        return `"${segment.replace(/"/g, '\\"')}"`;
      }
      return segment;
    })
    .join(" ");

  terminal.sendText(command, true);
}
