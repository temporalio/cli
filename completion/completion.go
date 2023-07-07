package completion

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// taken from https://github.com/urfave/cli/blob/v2-maint/autocomplete/zsh_autocomplete
var zsh_script = `#compdef temporal
_cli_zsh_autocomplete() {
  local -a opts
  local cur
  cur=${words[-1]}
  if [[ "$cur" == "-"* ]]; then
    opts=("${(@f)$(${words[@]:0:#words[@]-1} ${cur} --generate-bash-completion)}")
  else
    opts=("${(@f)$(${words[@]:0:#words[@]-1} --generate-bash-completion)}")
  fi
  if [[ "${opts[1]}" != "" ]]; then
    _describe 'values' opts
  else
    _files
  fi
}
compdef _cli_zsh_autocomplete temporal
`

// taken from https://github.com/urfave/cli/blob/v2-maint/autocomplete/bash_autocomplete
var bash_script = `#! /bin/bash
# Macs have bash3 for which the bash-completion package doesn't include
# _init_completion. This is a minimal version of that function.
_cli_init_completion() {
  COMPREPLY=()
  _get_comp_words_by_ref "$@" cur prev words cword
}
_cli_bash_autocomplete() {
  if [[ "${COMP_WORDS[0]}" != "source" ]]; then
    local cur opts base words
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    if declare -F _init_completion >/dev/null 2>&1; then
      _init_completion -n "=:" || return
    else
      _cli_init_completion -n "=:" || return
    fi
    words=("${words[@]:0:$cword}")
    if [[ "$cur" == "-"* ]]; then
      requestComp="${words[*]} ${cur} --generate-bash-completion"
    else
      requestComp="${words[*]} --generate-bash-completion"
    fi
    opts=$(eval "${requestComp}" 2>/dev/null)
    COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
    return 0
  fi
}
complete -o bashdefault -o default -F _cli_bash_autocomplete temporal
`

func NewCompletionCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "bash",
			Usage:     "bash completion output",
			UsageText: "source <(temporal completion bash)",
			Action: func(c *cli.Context) error {
				fmt.Fprintln(os.Stdout, bash_script)
				return nil
			},
		},
		{
			Name:      "zsh",
			Usage:     "zsh completion output",
			UsageText: "source <(temporal completion zsh)",
			Action: func(c *cli.Context) error {
				fmt.Fprintln(os.Stdout, zsh_script)
				return nil
			},
		},
	}
}
