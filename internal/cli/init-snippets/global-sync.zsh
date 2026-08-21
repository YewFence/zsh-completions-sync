_zcs_global_completion_dir=${ZCS_GLOBAL_OUTPUT_DIR:-${ZCS_OUTPUT_DIR:-$HOME/.zsh/completions}}
_zcs_global_outdated=0

for _zcs_global_completion in "$_zcs_global_completion_dir"/_*(N); do
  _zcs_global_tool=${_zcs_global_completion:t}
  _zcs_global_tool=${_zcs_global_tool#_}
  _zcs_global_executable=${commands[$_zcs_global_tool]}
  if [[ -n "$_zcs_global_executable" && "${_zcs_global_executable:A}" -nt "$_zcs_global_completion" ]]; then
    _zcs_global_outdated=1
    break
  fi
done

if (( _zcs_global_outdated )); then
  print -u2 -r -- "zcs: Some zsh completion scripts are out of date. Run 'zcs generate' to update them."
fi

unset _zcs_global_completion_dir _zcs_global_outdated _zcs_global_completion _zcs_global_tool _zcs_global_executable
