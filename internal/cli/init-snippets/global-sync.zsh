_zcs_global_completion_dir=${ZCS_GLOBAL_OUTPUT_DIR:-${ZCS_OUTPUT_DIR:-$HOME/.zsh/completions}}
_zcs_global_sync=0

for _zcs_global_completion in "$_zcs_global_completion_dir"/_*(N); do
  _zcs_global_tool=${_zcs_global_completion:t}
  _zcs_global_tool=${_zcs_global_tool#_}
  _zcs_global_executable=${commands[$_zcs_global_tool]}
  if [[ -n "$_zcs_global_executable" && "${_zcs_global_executable:A}" -nt "$_zcs_global_completion" ]]; then
    _zcs_global_sync=1
    break
  fi
done

if (( _zcs_global_sync )); then
  ZCS_OUTPUT_DIR="$_zcs_global_completion_dir" zcs global >/dev/null 2>&1
fi

unset _zcs_global_completion_dir _zcs_global_sync _zcs_global_completion _zcs_global_tool _zcs_global_executable
