_zcs_global_completion_dir=${ZCS_GLOBAL_OUTPUT_DIR:-${ZCS_OUTPUT_DIR:-$HOME/.zsh/completions}}

for dir in "$_zcs_global_completion_dir"; do
  if [[ -d "$dir" && ${fpath[(Ie)$dir]} -eq 0 ]]; then
    fpath=("$dir" $fpath)
  fi
done

unset _zcs_global_completion_dir
