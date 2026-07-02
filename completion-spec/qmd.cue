package qmd

let outputFormats = [
	"cli:colorized command-line output",
	"json:JSON output",
	"csv:CSV output",
	"md:Markdown output",
	"xml:XML output",
	"files:docid, score, file path, and context",
]

let chunkStrategies = [
	"regex:regex-based chunking",
	"auto:AST-aware chunking for supported code files",
]

name:        "qmd"
description: "Quick Markdown Search"

helpers: {
	collections: """
		_qmd_collections() {
		  local -a collections
		  collections=(${(f)"$(_call_program qmd-collections qmd collection list 2>/dev/null | awk 'NF { print $1 }')"})
		  _describe -t collections 'collection' collections
		}
		"""
}

values: {
	"output-formats":   outputFormats
	"chunk-strategies": chunkStrategies
}

optionGroups: {
	global: [
		{flags: ["-h", "--help"], description: "show help"},
		{flags: ["-v", "--version"], description: "show version"},
		{
			flags: ["--index"]
			description: "use named index"
			argument:    "index name"
		},
		{flags: ["--no-gpu"], description: "force CPU mode"},
		{flags: ["--skill"], description: "print the QMD skill"},
	]

	search: [
		{
			flags: ["-n", "--number"]
			description: "number of results"
			argument:    "number"
		},
		{
			flags: ["-c", "--collection"]
			description: "restrict to collection"
			argument:    "collection"
			completion:  "_qmd_collections"
		},
		{flags: ["--all"], description: "return all matches"},
		{
			flags: ["--min-score"]
			description: "minimum score threshold"
			argument:    "score"
		},
		{flags: ["--full"], description: "show full document content"},
		{flags: ["--line-numbers"], description: "add line numbers to output"},
		{flags: ["--explain"], description: "include retrieval score traces"},
		{
			flags: ["--intent"]
			description: "disambiguation context"
			argument:    "intent"
		},
		{flags: ["--no-rerank"], description: "skip LLM reranking"},
		{
			flags: ["-C", "--candidate-limit"]
			description: "maximum candidates to rerank"
			argument:    "number"
		},
		{
			flags: ["--chunk-strategy"]
			description: "chunk selection strategy"
			argument:    "strategy"
			completion:  "value:chunk-strategies"
		},
		{
			flags: ["--full-path"]
			description: "emit filesystem paths instead of qmd URIs"
		},
		{
			flags: ["--format"]
			description: "output format"
			argument:    "format"
			completion:  "value:output-formats"
		},
		{flags: ["--json"], description: "JSON output"},
		{flags: ["--csv"], description: "CSV output"},
		{flags: ["--md"], description: "Markdown output"},
		{flags: ["--xml"], description: "XML output"},
		{flags: ["--files"], description: "files output"},
	]

	get: [
		{
			flags: ["-l", "--lines"]
			description: "maximum lines to return"
			argument:    "lines"
		},
		{flags: ["--from"], description: "start line", argument: "line"},
		{flags: ["--no-line-numbers"], description: "disable line numbers"},
		{
			flags: ["--full-path"]
			description: "emit filesystem paths instead of qmd URIs"
		},
	]

	"multi-get": [
		{
			flags: ["-l", "--lines"]
			description: "maximum lines per file"
			argument:    "lines"
		},
		{
			flags: ["--max-bytes"]
			description: "skip files larger than bytes"
			argument:    "bytes"
		},
		{flags: ["--no-line-numbers"], description: "disable line numbers"},
		{
			flags: ["--full-path"]
			description: "emit filesystem paths instead of qmd URIs"
		},
		{
			flags: ["--format"]
			description: "output format"
			argument:    "format"
			completion:  "value:output-formats"
		},
		{flags: ["--json"], description: "JSON output"},
		{flags: ["--csv"], description: "CSV output"},
		{flags: ["--md"], description: "Markdown output"},
		{flags: ["--xml"], description: "XML output"},
		{flags: ["--files"], description: "files output"},
	]
}

globalOptionGroups: ["global"]

commands: [
	{name: "init", description: "initialize a project-local index"},
	{
		name:        "collection"
		description: "manage collections"
		commands: [
			{name: "list", description: "list all collections"},
			{
				name:        "add"
				description: "add a collection"
				options: [
					{
						flags: ["--name"]
						description: "collection name"
						argument:    "name"
					},
					{
						flags: ["--mask"]
						description: "glob pattern"
						argument:    "glob pattern"
					},
				]
				arguments: [
					{name: "directory", completion: "_files -/"},
				]
			},
			{
				name: "remove"
				aliases: ["rm"]
				description: "remove a collection"
				arguments: [
					{name: "collection", completion: "_qmd_collections"},
					{name: "command word", repeatable: true},
				]
			},
			{
				name: "rename"
				aliases: ["mv"]
				description: "rename a collection"
				arguments: [
					{name: "old collection", completion: "_qmd_collections"},
					{name: "new collection name"},
				]
			},
			{
				name: "show"
				aliases: ["info"]
				description: "show collection details"
				arguments: [
					{name: "collection", completion: "_qmd_collections"},
					{name: "command word", repeatable: true},
				]
			},
			{
				name: "update-cmd"
				aliases: ["set-update"]
				description: "set pre-update command"
				arguments: [
					{name: "collection", completion: "_qmd_collections"},
					{name: "command word", repeatable: true},
				]
			},
			{
				name:        "include"
				description: "include collection in default queries"
				arguments: [
					{name: "collection", completion: "_qmd_collections"},
					{name: "command word", repeatable: true},
				]
			},
			{
				name:        "exclude"
				description: "exclude collection from default queries"
				arguments: [
					{name: "collection", completion: "_qmd_collections"},
					{name: "command word", repeatable: true},
				]
			},
			{name: "help", description: "show collection help"},
		]
	},
	{
		name:        "context"
		description: "manage search context"
		commands: [
			{
				name:        "add"
				description: "add context"
				arguments: [
					{
						name: "path"
						completion: """
							alternative:qmd-paths:qmd path:_guard "qmd://*" "qmd path"|files:file:_files
							"""
						repeatable: true
					},
				]
			},
			{name: "list", description: "list contexts"},
			{
				name: "rm"
				aliases: ["remove"]
				description: "remove context"
				arguments: [
					{
						name: "path"
						completion: """
							alternative:qmd-paths:qmd path:_guard "qmd://*" "qmd path"|files:file:_files
							"""
						repeatable: true
					},
				]
			},
			{name: "check", description: "check paths missing context"},
		]
	},
	{
		name:        "ls"
		description: "list collections or files"
		arguments: [
			{
				name: "collection or qmd path"
				completion: """
					alternative:collections:collection:_qmd_collections|qmd-paths:qmd path:_guard "qmd://*" "qmd path"
					"""
			},
		]
	},
	{
		name:        "get"
		description: "get a document by path or docid"
		optionGroups: ["get"]
		arguments: [{name: "file or docid", completion: "_files"}]
	},
	{
		name:        "multi-get"
		description: "get multiple documents by glob or list"
		optionGroups: ["multi-get"]
		arguments: [{name: "glob, list, or docid", completion: "_files"}]
	},
	{name: "status", description: "show index status"},
	{name: "doctor", description: "diagnose the install"},
	{
		name:        "update"
		description: "re-index collections"
		options: [
			{
				flags: ["--pull"]
				description: "run collection pre-update commands before indexing"
			},
		]
	},
	{
		name:        "embed"
		description: "generate vector embeddings"
		options: [
			{flags: ["-f", "--force"], description: "re-embed everything"},
			{
				flags: ["-c", "--collection"]
				description: "embed collection"
				argument:    "collection"
				completion:  "_qmd_collections"
			},
			{
				flags: ["--chunk-strategy"]
				description: "chunk strategy"
				argument:    "strategy"
				completion:  "value:chunk-strategies"
			},
			{
				flags: ["--max-docs-per-batch"]
				description: "maximum docs per batch"
				argument:    "number"
			},
			{
				flags: ["--max-batch-mb"]
				description: "maximum batch size in MB"
				argument:    "megabytes"
			},
			{
				flags: ["--timeout"]
				description: "maximum duration"
				argument:    "duration"
			},
		]
	},
	{
		name:        "pull"
		description: "pull local models"
		options: [
			{flags: ["--refresh"], description: "refresh cached model metadata"},
		]
	},
	{
		name:        "search"
		description: "full-text keyword search"
		optionGroups: ["search"]
		arguments: [{name: "query", repeatable: true}]
	},
	{
		name:        "vsearch"
		description: "vector similarity search"
		optionGroups: ["search"]
		arguments: [{name: "query", repeatable: true}]
	},
	{
		name:        "vector-search"
		description: "vector similarity search"
		optionGroups: ["search"]
		arguments: [{name: "query", repeatable: true}]
	},
	{
		name:        "query"
		description: "hybrid search with expansion and reranking"
		optionGroups: ["search"]
		arguments: [{name: "query", repeatable: true}]
	},
	{
		name:        "deep-search"
		description: "hybrid search with expansion and reranking"
		optionGroups: ["search"]
		arguments: [{name: "query", repeatable: true}]
	},
	{
		name:        "bench"
		description: "run search quality benchmarks"
		options: [
			{
				flags: ["-c", "--collection"]
				description: "benchmark collection"
				argument:    "collection"
				completion:  "_qmd_collections"
			},
			{flags: ["--json"], description: "JSON output"},
		]
		arguments: [
			{
				name: "fixture JSON"
				completion: """
					_files -g "*.json(-.)"
					"""
			},
		]
	},
	{
		name:        "mcp"
		description: "start MCP server"
		options: [
			{flags: ["--http"], description: "start MCP server with HTTP transport"},
			{flags: ["--daemon"], description: "start as background daemon"},
			{flags: ["--port"], description: "HTTP port", argument: "port"},
			{
				flags: ["--host"]
				description: "HTTP bind host"
				argument:    "host"
				completion:  "_hosts"
			},
		]
		commands: [
			{name: "stop", description: "stop background MCP daemon"},
			{name: "status", description: "show MCP daemon status"},
		]
	},
	{
		name:        "skills"
		description: "list QMD skills"
		options: [
			{flags: ["--json"], description: "JSON output"},
			{flags: ["--full"], description: "include full skill content"},
			{flags: ["--all"], description: "include all skills"},
		]
		arguments: [{name: "skill name", repeatable: true}]
	},
	{
		name:        "skill"
		description: "show or install the QMD skill"
		commands: [
			{name: "show", description: "print the QMD skill"},
			{
				name:        "install"
				description: "install QMD skill"
				options: [
					{
						flags: ["--global"]
						description: "install into global skills directory"
					},
					{flags: ["--yes"], description: "also create Claude skill symlink"},
					{
						flags: ["-f", "--force"]
						description: "replace existing install or symlink"
					},
				]
			},
			{name: "help", description: "show skill help"},
		]
	},
	{name: "cleanup", description: "clean cache and orphaned data"},
]
