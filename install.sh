#!/usr/bin/env bash

# The MIT License (MIT)

# Copyright (c) 2010 Tim Caswell

# Copyright (c) 2014 Jordan Harband

# Copyright (c) 2022 Temporal Technologies Inc.

# Permission is hereby granted, free of charge, to any person obtaining a copy of
# this software and associated documentation files (the "Software"), to deal in
# the Software without restriction, including without limitation the rights to
# use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
# the Software, and to permit persons to whom the Software is furnished to do so,
# subject to the following conditions:

# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
# FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
# COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
# IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
# CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

{ # this ensures the entire script is downloaded #

nvm_has() {
  type "$1" > /dev/null 2>&1
}

nvm_echo() {
  command printf %s\\n "$*" 2>/dev/null
}

if [ -z "${BASH_VERSION}" ] || [ -n "${ZSH_VERSION}" ]; then
  # shellcheck disable=SC2016
  nvm_echo >&2 'Error: Pipe the install script to `bash`'
  exit 1
fi

nvm_grep() {
  GREP_OPTIONS='' command grep "$@"
}

nvm_default_install_dir() {
  [ -z "${XDG_CONFIG_HOME-}" ] && printf %s "${HOME}/.temporalio" || printf %s "${XDG_CONFIG_HOME}/temporalio"
}

nvm_install_dir() {
  if [ -n "$TEMPORAL_DIR" ]; then
    printf %s "${TEMPORAL_DIR}"
  else
    nvm_default_install_dir
  fi
}

nvm_profile_is_bash_or_zsh() {
  local TEST_PROFILE
  TEST_PROFILE="${1-}"
  case "${TEST_PROFILE-}" in
    *"/.bashrc" | *"/.bash_profile" | *"/.zshrc" | *"/.zprofile")
      return
    ;;
    *)
      return 1
    ;;
  esac
}

#
# Fetches archive meta data from the Temporal CDN
#
temporal_fetch_meta() {
  local TEMPORAL_SOURCE_URL
  TEMPORAL_SOURCE_URL="$TEMPORAL_SOURCE"
  local TEMPORAL_ARCH
  TEMPORAL_ARCH="$(temporal_arch)"
  local TEMPORAL_PLATFORM
  TEMPORAL_PLATFORM="$(temporal_os)"
  TEMPORAL_SOURCE_META_URL="https://temporal.download/temporalite/latest?platform=${TEMPORAL_PLATFORM}&arch=${TEMPORAL_ARCH}"

  TEMPORAL_SOURCE_URL="$(temporal_fetch "$TEMPORAL_SOURCE_META_URL")"

  nvm_echo "$TEMPORAL_SOURCE_URL"
}

temporal_archive_url() {
  local TEMPORAL_ARCHIVE_URL
  TEMPORAL_ARCHIVE_URL="$(temporal_fetch_meta | jq '.archiveUrl' | tr -d '"')"

  nvm_echo "$TEMPORAL_ARCHIVE_URL"
}

temporal_binary_name() {
  local TEMPORAL_FILENAME
  TEMPORAL_FILENAME="$(temporal_fetch_meta | jq '.fileToExtract' | tr -d '"')"

  nvm_echo "$TEMPORAL_FILENAME"
}

temporal_arch() {
  local TEMPORAL_ARCH
  TEMPORAL_ARCH="$(uname -m)"
  case "$TEMPORAL_ARCH" in
    "x86_64" | "amd64")
      TEMPORAL_ARCH="amd64"
    ;;
    "arm64" | "aarch64")
      TEMPORAL_ARCH="arm64"
    ;;
    *)
      nvm_echo >&2 "Unsupported architecture $(uname -m)"
      return 1
    ;;
  esac
  nvm_echo "$TEMPORAL_ARCH"
}

temporal_os() {
  local TEMPORAL_OS
  TEMPORAL_OS="$(uname -s)"
  case "$TEMPORAL_OS" in
    "Linux")
      TEMPORAL_OS="linux"
    ;;
    "Darwin")
      TEMPORAL_OS="darwin"
    ;;
    *)
      nvm_echo >&2 "Unsupported OS $(uname -s)"
      return 1
    ;;
  esac
  nvm_echo "$TEMPORAL_OS"
}

# 
# Unarchives and deletes the archive
#
temporal_unzip_and_delete() {
  local TEMPORAL_INSTALL_DIR
  TEMPORAL_INSTALL_DIR="$1"
  local TEMPORAL_FILENAME
  TEMPORAL_FILENAME="$2"

  tar -xzvf "$TEMPORAL_INSTALL_DIR/$TEMPORAL_FILENAME" -C "$TEMPORAL_INSTALL_DIR"
  rm "$TEMPORAL_INSTALL_DIR/$TEMPORAL_FILENAME"
}

nvm_download() {
  if nvm_has "curl"; then
    curl --fail --compressed --show-error -q "$@"
  elif nvm_has "wget"; then
    # Emulate curl with wget
    ARGS=$(nvm_echo "$@" | command sed -e 's/--progress-bar /--progress=bar /' \
                            -e 's/--compressed //' \
                            -e 's/--fail //' \
                            -e 's/-L //' \
                            -e 's/-I /--server-response /' \
                            -e 's/-s /-q /' \
                            -e 's/-sS /-nv /' \
                            -e 's/-o /-O /' \
                            -e 's/-C - /-c /')
    # shellcheck disable=SC2086
    eval wget $ARGS
  fi
}

temporal_fetch() {
  local TEMPORAL_URL
  TEMPORAL_URL="$1"
  local TEMPORAL_RESPONSE

  if nvm_has "curl"; then
      TEMPORAL_RESPONSE="$(curl --fail -s -q "$TEMPORAL_URL")"
  elif nvm_has "wget"; then
    # Emulate curl with wget
    ARGS=$(nvm_echo "$TEMPORAL_URL" | command sed -e 's/--progress-bar /--progress=bar /' \
                            -e 's/--compressed //' \
                            -e 's/--fail //' \
                            -e 's/-L //' \
                            -e 's/-I /--server-response /' \
                            -e 's/-s /-q /' \
                            -e 's/-sS /-nv /' \
                            -e 's/-o /-O /' \
                            -e 's/-C - /-c /')
    # shellcheck disable=SC2086
    TEMPORAL_RESPONSE="$(eval wget $ARGS)"
  fi

  nvm_echo "$TEMPORAL_RESPONSE"
}

temporal_install_env() {
  local TEMPORAL_DIR
  TEMPORAL_DIR="$1"

cat > "$(nvm_install_dir)/env" << EOL
#!/bin/sh
case ":\${PATH}:" in
    *:"$(nvm_install_dir)":*)
        ;;
    *)
        export PATH="$(nvm_install_dir):\$PATH"
        ;;
esac
EOL

  nvm_echo . "$(nvm_install_dir)/env"
}

install_nvm_as_script() {
  local INSTALL_DIR
  INSTALL_DIR="$(nvm_install_dir)"
  local TEMPORAL_EXEC_SOURCE
  TEMPORAL_EXEC_SOURCE="$(temporal_archive_url)"
  local TEMPORAL_BINARY
  TEMPORAL_BINARY="$(temporal_binary_name)"

  # Downloading to $INSTALL_DIR
  mkdir -p "$INSTALL_DIR"
  if [ -f "$INSTALL_DIR/$TEMPORAL_BINARY" ]; then
    nvm_echo "=> $TEMPORAL_BINARY is already installed in $INSTALL_DIR, trying to update the script"
  else
    nvm_echo "=> Downloading $TEMPORAL_BINARY to '$INSTALL_DIR'"
  fi
  nvm_download -s "$TEMPORAL_EXEC_SOURCE" -o "$INSTALL_DIR/temporal.tar.gz" || {
    nvm_echo >&2 "Failed to download $TEMPORAL_EXEC_SOURCE"
    return 2
  }
  for job in $(jobs -p | command sort)
  do
    wait "$job" || return $?
  done

  temporal_unzip_and_delete "$INSTALL_DIR" temporal.tar.gz || {
    nvm_echo >&2 "Failed to unzip '$INSTALL_DIR/temporal.tar.gz'"
    return 2
  }

  chmod a+x "$INSTALL_DIR/$TEMPORAL_BINARY" || {
    nvm_echo >&2 "Failed to mark '$INSTALL_DIR/$TEMPORAL_BINARY' as executable"
    return 3
  }
}

nvm_try_profile() {
  if [ -z "${1-}" ] || [ ! -f "${1}" ]; then
    return 1
  fi
  nvm_echo "${1}"
}

#
# Detect profile file if not specified as environment variable
# (eg: PROFILE=~/.myprofile)
# The echo'ed path is guaranteed to be an existing file
# Otherwise, an empty string is returned
#
nvm_detect_profile() {
  if [ "${PROFILE-}" = '/dev/null' ]; then
    # the user has specifically requested NOT to have nvm touch their profile
    return
  fi

  if [ -n "${PROFILE}" ] && [ -f "${PROFILE}" ]; then
    nvm_echo "${PROFILE}"
    return
  fi

  local DETECTED_PROFILE
  DETECTED_PROFILE=''

  if [ "${SHELL#*bash}" != "$SHELL" ]; then
    if [ -f "$HOME/.bashrc" ]; then
      DETECTED_PROFILE="$HOME/.bashrc"
    elif [ -f "$HOME/.bash_profile" ]; then
      DETECTED_PROFILE="$HOME/.bash_profile"
    fi
  elif [ "${SHELL#*zsh}" != "$SHELL" ]; then
    if [ -f "$HOME/.zshrc" ]; then
      DETECTED_PROFILE="$HOME/.zshrc"
    elif [ -f "$HOME/.zprofile" ]; then
      DETECTED_PROFILE="$HOME/.zprofile"
    fi
  fi

  if [ -z "$DETECTED_PROFILE" ]; then
    for EACH_PROFILE in ".profile" ".bashrc" ".bash_profile" ".zprofile" ".zshrc"
    do
      if DETECTED_PROFILE="$(nvm_try_profile "${HOME}/${EACH_PROFILE}")"; then
        break
      fi
    done
  fi

  if [ -n "$DETECTED_PROFILE" ]; then
    nvm_echo "$DETECTED_PROFILE"
  fi
}

nvm_do_install() {
  local TEMPORAL_BINARY
  TEMPORAL_BINARY="$(temporal_binary_name)"

  if [ -n "${TEMPORAL_DIR-}" ] && ! [ -d "${TEMPORAL_DIR}" ]; then
    if [ -e "${TEMPORAL_DIR}" ]; then
      nvm_echo >&2 "File \"${TEMPORAL_DIR}\" has the same name as installation directory."
      exit 1
    fi

    if [ "${TEMPORAL_DIR}" = "$(nvm_default_install_dir)" ]; then
      mkdir "${TEMPORAL_DIR}"
    else
      nvm_echo >&2 "You have \$TEMPORAL_DIR set to \"${TEMPORAL_DIR}\", but that directory does not exist. Check your profile files and environment."
      exit 1
    fi
  fi
  # Disable the optional which check, https://www.shellcheck.net/wiki/SC2230
  # shellcheck disable=SC2230
  if nvm_has xcode-select && [ "$(xcode-select -p >/dev/null 2>/dev/null ; echo $?)" = '2' ] && [ "$(which git)" = '/usr/bin/git' ] && [ "$(which curl)" = '/usr/bin/curl' ]; then
    nvm_echo >&2 'You may be on a Mac, and need to install the Xcode Command Line Developer Tools.'
    # shellcheck disable=SC2016
    nvm_echo >&2 'If so, run `xcode-select --install` and try again. If not, please report this!'
    exit 1
  fi

  if nvm_has curl || nvm_has wget; then
    install_nvm_as_script  || {
    exit 1
  }
  else
    nvm_echo >&2 'You need curl or wget to install $TEMPORAL_BINARY'
    exit 1
  fi

  nvm_echo

  local NVM_PROFILE
  NVM_PROFILE="$(nvm_detect_profile)"
  local PROFILE_INSTALL_DIR
  PROFILE_INSTALL_DIR="$(nvm_install_dir | command sed "s:^$HOME:\$HOME:")"

  SOURCE_STR="$(temporal_install_env)"

  BASH_OR_ZSH=false

  if [ -z "${NVM_PROFILE-}" ] ; then
    local TRIED_PROFILE
    if [ -n "${PROFILE}" ]; then
      TRIED_PROFILE="${NVM_PROFILE} (as defined in \$PROFILE), "
    fi
    nvm_echo "=> Profile not found. Tried ${TRIED_PROFILE-}~/.bashrc, ~/.bash_profile, ~/.zprofile, ~/.zshrc, and ~/.profile."
    nvm_echo "=> Create one of them and run this script again"
    nvm_echo "   OR"
    nvm_echo "=> Append the following lines to the correct file yourself:"
    command printf "${SOURCE_STR}"
    nvm_echo
  else
    if nvm_profile_is_bash_or_zsh "${NVM_PROFILE-}"; then
      BASH_OR_ZSH=true
    fi
    if ! command grep -qc '/.temporalio' "$NVM_PROFILE"; then
      nvm_echo "=> Appending $TEMPORAL_BINARY source string to $NVM_PROFILE"
      command printf "${SOURCE_STR}\n" >> "$NVM_PROFILE"
    else
      nvm_echo "=> $TEMPORAL_BINARY source string already in ${NVM_PROFILE}"
    fi
  fi

  # Source temporal
  # shellcheck source=/dev/null
  $SOURCE_STR

  nvm_reset

  nvm_echo "=> Close and reopen your terminal to start using $TEMPORAL_BINARY or run the following to use it now:"
  command printf "${SOURCE_STR}"
}

#
# Unsets the various functions defined
# during the execution of the install script
#
nvm_reset() {
  unset -f nvm_has nvm_install_dir nvm_profile_is_bash_or_zsh \
    nvm_download \
    install_nvm_as_script nvm_try_profile nvm_detect_profile \
    nvm_do_install nvm_reset nvm_default_install_dir nvm_grep \
    temporal_arch temporal_archive_url temporal_os temporal_unzip_and_delete \
    temporal_fetch temporal_binary_name
}

[ "_$TEMPORAL_ENV" = "_testing" ] || nvm_do_install

} # this ensures the entire script is downloaded #