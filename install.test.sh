#!/bin/sh
# shellcheck shell=dash

set -e

assert() {
  local _command="$1"
  local _string="$2"
  local _colored="$3"

  if ! eval "$_command" | grep -q "$_string"; then
    local _assertion_failed
    local _status="$(failure "Assertion failed:" $_colored)"
    printf "$_status '$_command' does not contain '$_string'\n"
    exit 1
  fi
}

failure() {
  local _string="$1"
  local _colored="$2"

  if $_colored; then
    _string="\33[1;31m$_string\33[0m"
  fi

  echo "$_string"
}

success() {
  local _string="$1"
  local _colored="$2"

  if $_colored; then
    _string="\33[1;32m$_string\33[0m"
  fi

  echo "$_string"
}

main() {
  sh ./install.sh
  . "$HOME"/.temporalio/env

  local _colored=false
  if [ -t 2 ]; then
    if [ "${TERM+set}" = 'set' ]; then
      case "$TERM" in
      xterm* | rxvt* | urxvt* | linux* | vt*)
        # ansi escapes are valid
        _colored=true
        ;;
      esac
    fi
  fi

  assert "temporal -v" "temporal version" $_colored
  assert "sh ./install.sh --help" "Temporal CLI" $_colored

  local _status="$(success "Tests passed" $_colored)"
  printf "$_status\n"
}

main "$@" || exit 1
