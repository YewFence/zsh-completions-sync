package waydroid

name:        "waydroid"
description: "Waydroid"

optionGroups: {
	global: [
		{flags: ["-h", "--help"], description: "show help message and exit"},
		{flags: ["-V", "--version"], description: "show program version and exit"},
		{
			flags: ["-l", "--log"]
			description: "path to log file"
			argument:    "log file"
			completion:  "_files"
		},
		{flags: ["--details-to-stdout"], description: "print details to stdout"},
		{flags: ["-v", "--verbose"], description: "write more to log files"},
		{flags: ["-q", "--quiet"], description: "do not output log messages"},
	]
}

globalOptionGroups: ["global"]

commands: [
	{name: "status", description: "quick check for the waydroid"},
	{
		name:        "log"
		description: "follow the waydroid logfile"
		options: [
			{
				flags: ["-n", "--lines"]
				description: "count of initial output lines"
				argument:    "lines"
			},
			{flags: ["-c", "--clear"], description: "clear the log"},
		]
	},
	{
		name:        "init"
		description: "set up waydroid specific configs and install images"
		options: [
			{
				flags: ["-i", "--images_path"]
				description: "custom path to waydroid images"
				argument:    "images path"
				completion:  "_files -/"
			},
			{
				flags: ["-f", "--force"]
				description: "re-initialize configs and images"
			},
			{
				flags: ["-c", "--system_channel"]
				description: "custom system channel"
				argument:    "system channel"
			},
			{
				flags: ["-v", "--vendor_channel"]
				description: "custom vendor channel"
				argument:    "vendor channel"
			},
			{
				flags: ["-r", "--rom_type"]
				description: "rom type"
				argument:    "rom type"
				completion:  "values:rom-types"
			},
			{
				flags: ["-s", "--system_type"]
				description: "system type"
				argument:    "system type"
				completion:  "values:system-types"
			},
			{flags: ["--client"], description: "run as user mode"},
		]
	},
	{
		name:        "upgrade"
		description: "upgrade images"
		options: [
			{flags: ["-o", "--offline"], description: "just for updating configs"},
		]
	},
	{
		name:        "session"
		description: "session controller"
		commands: [
			{name: "start", description: "start session"},
			{name: "stop", description: "stop session"},
		]
	},
	{
		name:        "container"
		description: "container controller"
		commands: [
			{name: "start", description: "start container"},
			{name: "stop", description: "stop container"},
			{name: "restart", description: "restart container"},
			{name: "freeze", description: "freeze container"},
			{name: "unfreeze", description: "unfreeze container"},
		]
	},
	{
		name:        "app"
		description: "applications controller"
		commands: [
			{
				name:        "install"
				description: "push a single package to the container and install it"
				arguments: [
					{
						name: "package"
						completion: """
							_files -g "*.apk(-.)"
							"""
					},
				]
			},
			{
				name:        "remove"
				description: "remove single app package from the container"
				arguments: [{name: "package"}]
			},
			{
				name:        "launch"
				description: "start single application"
				arguments: [{name: "package"}]
			},
			{
				name:        "intent"
				description: "start single application"
				arguments: [{name: "action"}, {name: "uri"}]
			},
			{name: "list", description: "list installed applications"},
		]
	},
	{
		name:        "prop"
		description: "android properties controller"
		commands: [
			{
				name:        "get"
				description: "get value of property from container"
				arguments: [{name: "key"}]
			},
			{
				name:        "set"
				description: "set value to property on container"
				arguments: [{name: "key"}, {name: "value"}]
			},
		]
	},
	{name: "show-full-ui", description: "show android full screen in window"},
	{
		name:        "first-launch"
		description: "start waydroid, prompting to initialize first if necessary"
	},
	{
		name:        "shell"
		description: "run remote shell command"
		options: [
			{flags: ["-u", "--uid"], description: "UID to run as", argument: "uid"},
			{flags: ["-g", "--gid"], description: "GID to run as", argument: "gid"},
			{
				flags: ["-s", "--context"]
				description: "security context"
				argument:    "context"
			},
			{
				flags: ["-L", "--nolsm"]
				description: "do not perform security domain transition"
			},
			{flags: ["-C", "--allcaps"], description: "do not drop capabilities"},
			{
				flags: ["-G", "--nocgroup"]
				description: "do not switch to the container cgroup"
			},
		]
		arguments: [{name: "command", repeatable: true}]
	},
	{
		name:        "logcat"
		description: "show android logcat"
		arguments: [{name: "args", repeatable: true}]
	},
	{
		name:        "adb"
		description: "manage adb connection"
		commands: [
			{name: "connect", description: "connect adb to the Android container"},
			{name: "disconnect", description: "disconnect adb from the Android container"},
		]
	},
	{name: "bugreport", description: "create a bugreport archive interactively"},
]

values: {
	"rom-types": [
		"lineage:LineageOS",
		"bliss:Bliss OS",
	]
	"system-types": [
		"VANILLA:vanilla Android system image",
		"FOSS:FOSS Android system image",
		"GAPPS:Google Apps Android system image",
	]
}
