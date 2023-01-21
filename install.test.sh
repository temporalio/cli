#!/bin/sh
# shellcheck shell=dash

set -e

assert() {
  local _command="$1"
  local _string="$2"
  local _ansi_escapes_are_valid="$3"

  if ! eval "$_command" | grep -q "$_string"; then
    local _assertion_failed
    local _status="$(failure "Assertion failed:" $_ansi_escapes_are_valid)"
    printf "$_status '$_command' does not contain '$_string'\n"
    exit 1
  fi
}

failure() {
  local _string="$1"
  local _ansi_escapes_are_valid="$2"

  if $_ansi_escapes_are_valid; then
    _string="\33[1;31m$_string\33[0m"
  fi

  echo "$_string"
}

success() {
  local _string="$1"
  local _ansi_escapes_are_valid="$2"

  if $_ansi_escapes_are_valid; then
    _string="\33[1;32m$_string\33[0m"
  fi

  echo "$_string"
}

main() {
  sh ./install.sh
  . "$HOME"/.temporalio/env

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

  assert "temporal -v" "temporal version" $_ansi_escapes_are_valid
  assert "sh ./install.sh --help" "Temporal CLI" $_ansi_escapes_are_valid

  local _status="$(success "Tests passed" $_ansi_escapes_are_valid)"
  printf "$_status\n"
}

main "$@" || exit 1
