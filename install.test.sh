#!/bin/sh
# shellcheck shell=dash

set -e

assert() {
  local _command="$1"
  local _string="$2"
  local _ansi_escapes_are_valid="$3"

  if ! eval "$_command" | grep -q "$_string"; then
    local _assertion_failed
    if $_ansi_escapes_are_valid; then
      _assertion_failed="\33[1;31mAssertion failed:\33[0m"
    else
      _assertion_failed="Assertion failed:"
    fi

    printf "$_assertion_failed '$_command' does not contain $_string\n"
  fi
}

main() {
  sh ./install.sh
  . "$HOME"/.temporalio/env
  temporal -v

  local _ansi_escapes_are_valid=false
  if [ -t 2 ]; then
    if [ "${TERM+set}" = 'set' ]; then
      case "$TERM" in
      xterm* | rxvt* | urxvt* | linux* | vt*)
        _ansi_escapes_are_valid=true
        ;;
      esac
    fi
  fi

  assert "temporal -v" "temporal" $_ansi_escapes_are_valid
  assert "sh ./install.sh --help" "Temporal CLI" $_ansi_escapes_are_valid
}

main "$@" || exit 1
