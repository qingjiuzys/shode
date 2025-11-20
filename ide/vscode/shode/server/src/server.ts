import {
  createConnection,
  TextDocuments,
  ProposedFeatures,
  CompletionItem,
  CompletionItemKind,
  InitializeParams,
  TextDocumentSyncKind,
  Diagnostic,
  DiagnosticSeverity,
  Hover,
  Position,
} from "vscode-languageserver/node";
import { TextDocument } from "vscode-languageserver-textdocument";

const connection = createConnection(ProposedFeatures.all);
const documents: TextDocuments<TextDocument> = new TextDocuments(TextDocument);

const stdlibDocs: Record<string, string> = {
  ReadFile: "Read the contents of a file",
  WriteFile: "Write text to a file",
  ListFiles: "List entries inside a directory",
  FileExists: "Return true if a file exists",
  CopyFile: "Copy a file to a new location",
  Move: "Move files or directories",
  MkdirAll: "Create directory tree",
  Remove: "Delete files or directories recursively",
  Glob: "Return paths matching a glob pattern",
  TempFile: "Create a temporary file",
  Contains: "Check substring existence",
  Replace: "Replace substrings",
  ToUpper: "Convert to uppercase",
  ToLower: "Convert to lowercase",
  Trim: "Trim whitespace",
  Split: "Split a string by separator",
  Join: "Join multiple strings",
  MatchRegex: "Test regex match",
  ReplaceRegex: "Replace via regex",
  JSONEncodeMap: "Encode map as JSON string",
  JSONDecodeToMap: "Parse JSON string into a map",
  JSONPretty: "Pretty-print JSON",
  SleepSeconds: "Pause execution",
  TimeNowRFC3339: "Return current timestamp string",
  GenerateUUID: "Create a random UUID",
  HTTPGet: "Perform HTTP GET request",
  HTTPPostJSON: "POST JSON payload",
  Println: "Print with newline",
  Print: "Print without newline",
  Errorln: "Write to stderr with newline",
  Error: "Write to stderr",
};

const stdlibCompletions: CompletionItem[] = Object.keys(stdlibDocs).map(
  (name) => ({
    label: name,
    kind: CompletionItemKind.Function,
    data: name,
    detail: "Shode Stdlib",
    documentation: stdlibDocs[name],
  }),
);

connection.onInitialize((_params: InitializeParams) => {
  return {
    capabilities: {
      textDocumentSync: TextDocumentSyncKind.Incremental,
      completionProvider: {
        resolveProvider: false,
        triggerCharacters: ["."],
      },
      hoverProvider: true,
    },
  };
});

documents.onDidChangeContent((change) => {
  validateTextDocument(change.document);
});

connection.onCompletion((_textDocumentPosition) => {
  return stdlibCompletions;
});

connection.onHover((params): Hover | null => {
  const document = documents.get(params.textDocument.uri);
  if (!document) {
    return null;
  }

  const word = getWordAtPosition(document, params.position);
  if (word && stdlibDocs[word]) {
    return {
      contents: {
        kind: "markdown",
        value: `**${word}**\n\n${stdlibDocs[word]}`,
      },
    };
  }

  return null;
});

documents.listen(connection);
connection.listen();

function validateTextDocument(textDocument: TextDocument): void {
  const diagnostics: Diagnostic[] = [];
  const text = textDocument.getText();

  const lines = text.split(/\r?\n/);
  lines.forEach((line, index) => {
    const trimmed = line.trim();
    if (trimmed.startsWith("TODO")) {
      diagnostics.push({
        severity: DiagnosticSeverity.Information,
        range: {
          start: { line: index, character: 0 },
          end: { line: index, character: line.length },
        },
        message: "TODO found in script",
        source: "shode-lsp",
      });
    }

    if (trimmed.includes("rm -rf /")) {
      diagnostics.push({
        severity: DiagnosticSeverity.Warning,
        range: {
          start: { line: index, character: 0 },
          end: { line: index, character: line.length },
        },
        message: "Potentially dangerous command detected",
        source: "shode-lsp",
      });
    }
  });

  connection.sendDiagnostics({ uri: textDocument.uri, diagnostics });
}

function getWordAtPosition(document: TextDocument, position: Position): string {
  const line = document.getText({
    start: { line: position.line, character: 0 },
    end: { line: position.line + 1, character: 0 },
  });

  const regex = /[A-Za-z_][A-Za-z0-9_]*/g;
  let match: RegExpExecArray | null;
  while ((match = regex.exec(line))) {
    const start = match.index;
    const end = start + match[0].length;
    if (position.character >= start && position.character <= end) {
      return match[0];
    }
  }
  return "";
}
