autoload -Uz compinit
_zcs_project_compdump="$PWD/.completions/zsh/.zcompdump"
compinit -d "$_zcs_project_compdump"
unset _zcs_project_compdump
