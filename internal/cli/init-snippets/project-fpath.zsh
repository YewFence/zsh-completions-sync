for dir in "$PWD/.completions/zsh" "$HOME/.zsh/completions"; do
  if [[ -d "$dir" && ${fpath[(Ie)$dir]} -eq 0 ]]; then
    fpath=("$dir" $fpath)
  fi
done
