package completionspec

#Spec: {
	name:         string
	description?: string
	helpers?: [string]: string
	values?: [string]: [...string]
	optionGroups?: [string]: [...#Option]
	globalOptionGroups?: [...string]
	globalOptions?: [...#Option]
	commands?: [...#Command]
}

#Command: {
	name:         string
	description?: string
	aliases?: [...string]
	optionGroups?: [...string]
	options?: [...#Option]
	arguments?: [...#Argument]
	commands?: [...#Command]
}

#Option: {
	flags: [...string]
	description?: string
	argument?:    string
	completion?:  string
	repeatable?:  bool
}

#Argument: {
	name:         string
	description?: string
	completion?:  string
	repeatable?:  bool
}
