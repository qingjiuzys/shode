package stdlib

// FunctionMap defines the mapping between function names and their implementations
var FunctionMap = map[string]interface{}{
	// File System Operations
	"cat":      (*StdLib).ReadFile,
	"readfile": (*StdLib).ReadFile,
	"write":    (*StdLib).WriteFile,
	"writefile": (*StdLib).WriteFile,
	"ls":       (*StdLib).ListFiles,
	"list":     (*StdLib).ListFiles,
	"exists":   (*StdLib).FileExists,
	// CopyFile and MoveFile not yet implemented
	// "cp":       (*StdLib).CopyFile,
	// "copy":     (*StdLib).CopyFile,
	// "mv":       (*StdLib).MoveFile,
	// "move":     (*StdLib).MoveFile,
	"rm":       (*StdLib).DeleteFile,
	"delete":   (*StdLib).DeleteFile,
	"rmdir":    (*StdLib).DeleteDir,
	"mkdir":    (*StdLib).MakeDir,
	"size":     (*StdLib).FileSize,
	"mtime":    (*StdLib).FileModTime,
	"isdir":    (*StdLib).IsDir,
	"isfile":   (*StdLib).IsFile,
	"chmod":    (*StdLib).Chmod,
	"chown":    (*StdLib).Chown,
	"glob":     (*StdLib).Glob,
	"walk":     (*StdLib).Walk,

	// String Operations
	"contains": (*StdLib).Contains,
	"replace":  (*StdLib).Replace,
	"upper":    (*StdLib).ToUpper,
	"lower":    (*StdLib).ToLower,
	"trim":     (*StdLib).Trim,
	"split":    (*StdLib).Split,
	"join":     (*StdLib).Join,
	"hasprefix": (*StdLib).HasPrefix,
	"hassuffix": (*StdLib).HasSuffix,
	"index":    (*StdLib).Index,
	"lastindex": (*StdLib).LastIndex,
	"count":    (*StdLib).Count,
	"repeat":   (*StdLib).Repeat,
	"compare":  (*StdLib).Compare,

	// Regular Expression Operations
	"match":    (*StdLib).RegexMatch,
	"find":     (*StdLib).RegexFind,
	"findall":  (*StdLib).RegexFindAll,
	"regexreplace": (*StdLib).RegexReplace,

	// Environment Operations
	"getenv":   (*StdLib).GetEnv,
	"setenv":   (*StdLib).SetEnv,
	"pwd":      (*StdLib).WorkingDir,
	"cd":       (*StdLib).ChangeDir,

	// System Information
	"hostname": (*StdLib).Hostname,
	"whoami":   (*StdLib).GetUsername,
	"pid":      (*StdLib).GetPID,
	"ppid":     (*StdLib).GetPPID,
	"sleep":    (*StdLib).Sleep,
	"now":      (*StdLib).Now,

	// Network Operations
	"httpget":  (*StdLib).HTTPGet,
	"httppost": (*StdLib).HTTPPost,

	// Cryptographic Operations
	"md5":      (*StdLib).MD5Hash,
	"sha1":     (*StdLib).SHA1Hash,
	"sha256":   (*StdLib).SHA256Hash,
	"base64encode": (*StdLib).Base64Encode,
	"base64decode": (*StdLib).Base64Decode,

	// Data Processing
	"json":     (*StdLib).JSONStringify,
	"jsonparse": (*StdLib).JSONParse,

	// Process Execution
	"exec":     (*StdLib).Exec,
	"exectimeout": (*StdLib).ExecWithTimeout,

	// Utility Functions
	"print":    (*StdLib).Print,
	"println":  (*StdLib).Println,
	"error":    (*StdLib).Error,
	"errorln":  (*StdLib).Errorln,
}

// HasFunction checks if a function exists in the standard library
func (sl *StdLib) HasFunction(name string) bool {
	_, exists := FunctionMap[name]
	return exists
}

// GetFunction returns the function implementation by name
func (sl *StdLib) GetFunction(name string) (interface{}, bool) {
	fn, exists := FunctionMap[name]
	return fn, exists
}

// ListFunctions returns all available function names
func (sl *StdLib) ListFunctions() []string {
	names := make([]string, 0, len(FunctionMap))
	for name := range FunctionMap {
		names = append(names, name)
	}
	return names
}

// FunctionCategories returns functions grouped by category
func (sl *StdLib) FunctionCategories() map[string][]string {
	categories := map[string][]string{
		"File System": {
			"cat", "readfile", "write", "writefile", "ls", "list", 
			"exists", "cp", "copy", "mv", "move", "rm", "delete",
			"rmdir", "mkdir", "size", "mtime", "isdir", "isfile",
			"chmod", "chown", "glob", "walk",
		},
		"String": {
			"contains", "replace", "upper", "lower", "trim", "split",
			"join", "hasprefix", "hassuffix", "index", "lastindex",
			"count", "repeat", "compare",
		},
		"Regular Expressions": {
			"match", "find", "findall", "regexreplace",
		},
		"Environment": {
			"getenv", "setenv", "pwd", "cd",
		},
		"System Info": {
			"hostname", "whoami", "pid", "ppid", "sleep", "now",
		},
		"Network": {
			"httpget", "httppost",
		},
		"Crypto": {
			"md5", "sha1", "sha256", "base64encode", "base64decode",
		},
		"Data Processing": {
			"json", "jsonparse",
		},
		"Process": {
			"exec", "exectimeout",
		},
		"Utility": {
			"print", "println", "error", "errorln",
		},
	}
	return categories
}
