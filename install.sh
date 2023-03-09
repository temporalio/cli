#!/bin/sh
# shellcheck shell=dash

# The MIT License (MIT)

# Copyright (c) 2016 The Rust Project Developers

# Copyright (c) 2022 Temporal Technologies Inc.

# Permission is hereby granted, free of charge, to any
# person obtaining a copy of this software and associated
# documentation files (the "Software"), to deal in the
# Software without restriction, including without
# limitation the rights to use, copy, modify, merge,
# publish, distribute, sublicense, and/or sell copies of
# the Software, and to permit persons to whom the Software
# is furnished to do so, subject to the following
# conditions:

# The above copyright notice and this permission notice
# shall be included in all copies or substantial portions
# of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF
# ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
# TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
# PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT
# SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
# CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
# OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR
# IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
# DEALINGS IN THE SOFTWARE.

# This is just a little script that can be downloaded from the internet to
# install Temporal CLI. It just does platform detection, downloads the installer
# and runs it.

# It runs on Unix shells like {a,ba,da,k,z}sh. It uses the common `local`
# extension. Note: Most shells limit `local` to 1 var per line, contra bash.

if [ "$KSH_VERSION" = 'Version JM 93t+ 2010-03-05' ]; then
    # The version of ksh93 that ships with many illumos systems does not
    # support the "local" extension.  Print a message rather than fail in
    # subtle ways later on:
    echo 'temporal does not work with this ksh93 version; please try bash!' >&2
    exit 1
fi

set -u

#XXX: If you change anything here, please make the same changes in setup_mode.rs
usage() {
    cat <<EOF
install.sh 0.1.0
The installer for the Temporal CLI

USAGE:
    sh install.sh [FLAGS] [OPTIONS]

FLAGS:
    --help        Prints help information

OPTIONS:
    --dir         Installation directory (default: $HOME/.temporalio)
EOF
}

main() {
    downloader --check
    need_cmd uname
    need_cmd mktemp
    need_cmd chmod
    need_cmd mkdir
    need_cmd rm
    need_cmd tar

    get_architecture || return 1
    local _arch
    _arch="$RETVAL"
    assert_nz "$_arch" "arch"

    get_platform || return 1
    local _platform
    _platform="$RETVAL"

    assert_nz "$_platform" "platform"

    local _ansi_escapes_are_valid
    _ansi_escapes_are_valid=false
    if [ -t 2 ]; then
        if [ "${TERM+set}" = 'set' ]; then
            case "$TERM" in
            xterm* | rxvt* | urxvt* | linux* | vt*)
                _ansi_escapes_are_valid=true
                ;;
            esac
        fi
    fi

    for arg in "$@"; do
        case "$arg" in
        --help)
            usage
            exit 0
            ;;
        esac
    done

    say "Downloading Temporal CLI" >&2

    local _temp
    if ! _temp="$(ensure mktemp -d)"; then
        # Because the previous command ran in a subshell, we must manually
        # propagate exit status.
        exit 1
    fi

    local _ext
    _ext=".tar.gz"

    case "$_arch" in
    *windows*)
        _ext=".zip"
        ;;
    esac
    
    local _archive_path
    _archive_path="${_temp}/temporal_cli_latest${_ext}"
    local _url
    _url="https://temporal.download/cli/archive/latest?platform=${_platform}&arch=${_arch}"
    
    ensure downloader "$_url" "$_archive_path" "$_arch"
    ensure unzip "$_archive_path" "$_temp"
    local _dir
    _dir="$(ensure get_install_dir "$@")"
    if [ -z "$_dir" ]; then
        exit 1
    fi

    local _dirbin
    _dirbin="$_dir/bin"
    ensure mkdir -p "$_dirbin"

    local _bext
    _bext=""
    case "$_arch" in
    *windows*)
        _bext=".exe"
        ;;
    esac

    local _exe_name
    _exe_name="temporal$_bext"
    ensure mv "$_temp/${_exe_name}" "$_dirbin"
    ensure rm -rf "$_temp"
    ensure chmod u+x "$_dirbin/$_exe_name"

    ensure prompt_for_path "$_dirbin"

    local _retval
    _retval=$?
    return "$_retval"
}

get_architecture() {
    local _arch
    _arch="$(uname -m)"

    case "$_arch" in
    "x86_64" | "amd64")
        _arch="amd64"
        ;;
    "arm64" | "aarch64")
        _arch="arm64"
        ;;
    *)
        err "Unsupported architecture $_arch"
        # shellcheck disable=SC2317
        return 1
        ;;
    esac

    RETVAL="$_arch"
}

get_platform() {
    local _platform
    _platform="$(uname -s)"

    case "$_platform" in
    "Linux")
        _platform="linux"
        ;;
    "Darwin")
        _platform="darwin"
        ;;
    *)
        err "Unsupported OS $_platform"
        # shellcheck disable=SC2317
        return 1
        ;;
    esac

    RETVAL="$_platform"
}

say() {
    if $_ansi_escapes_are_valid; then
        printf "\33[1mtemporal:\33[0m %s\n" "$1"
    else
        printf 'temporal: %s\n' "$1"
    fi
}

err() {
    say "$1" >&2
    exit 1
}

need_cmd() {
    if ! check_cmd "$1"; then
        err "need '$1' (command not found)"
    fi
}

check_cmd() {
    command -v "$1" >/dev/null 2>&1
}

assert_nz() {
    if [ -z "$1" ]; then err "assert_nz $2"; fi
}

# Run a command that should never fail. If the command fails execution
# will immediately terminate with an error showing the failing
# command.
ensure() {
    if ! "$@"; then err "command failed: $*"; fi
}

# This wraps curl or wget. Try curl first, if not installed,
# use wget instead.
downloader() {
    local _dld
    local _ciphersuites
    local _err
    local _status
    if check_cmd curl; then
        _dld=curl
    elif check_cmd wget; then
        _dld=wget
    else
        _dld='curl or wget' # to be used in error message of need_cmd
    fi

    if [ "$1" = --check ]; then
        need_cmd "$_dld"
    elif [ "$_dld" = curl ]; then
        get_ciphersuites_for_curl
        _ciphersuites="$RETVAL"
        if [ -n "$_ciphersuites" ]; then
            _err=$(curl --proto '=https' --tlsv1.2 --ciphers "$_ciphersuites" --silent --show-error --fail --location "$1" --output "$2" 2>&1)
            _status=$?
        else
            echo "Warning: Not enforcing strong cipher suites for TLS, this is potentially less secure"
            if ! check_help_for "$3" curl --proto --tlsv1.2; then
                echo "Warning: Not enforcing TLS v1.2, this is potentially less secure"
                _err=$(curl --silent --show-error --fail --location "$1" --output "$2" 2>&1)
                _status=$?
            else
                _err=$(curl --proto '=https' --tlsv1.2 --silent --show-error --fail --location "$1" --output "$2" 2>&1)
                _status=$?
            fi
        fi
        if [ -n "$_err" ]; then
            echo "$_err" >&2
            if echo "$_err" | grep -q 404$; then
                err "installer for platform '$3' not found, this may be unsupported"
            fi
        fi
        return $_status
    elif [ "$_dld" = wget ]; then
        if [ "$(wget -V 2>&1 | head -2 | tail -1 | cut -f1 -d" ")" = "BusyBox" ]; then
            echo "Warning: using the BusyBox version of wget.  Not enforcing strong cipher suites for TLS or TLS v1.2, this is potentially less secure"
            _err=$(wget "$1" -O "$2" 2>&1)
            _status=$?
        else
            get_ciphersuites_for_wget
            _ciphersuites="$RETVAL"
            if [ -n "$_ciphersuites" ]; then
                _err=$(wget --https-only --secure-protocol=TLSv1_2 --ciphers "$_ciphersuites" "$1" -O "$2" 2>&1)
                _status=$?
            else
                echo "Warning: Not enforcing strong cipher suites for TLS, this is potentially less secure"
                if ! check_help_for "$3" wget --https-only --secure-protocol; then
                    echo "Warning: Not enforcing TLS v1.2, this is potentially less secure"
                    _err=$(wget "$1" -O "$2" 2>&1)
                    _status=$?
                else
                    _err=$(wget --https-only --secure-protocol=TLSv1_2 "$1" -O "$2" 2>&1)
                    _status=$?
                fi
            fi
        fi
        if [ -n "$_err" ]; then
            echo "$_err" >&2
            if echo "$_err" | grep -q ' 404 Not Found$'; then
                err "installer for platform '$3' not found, this may be unsupported"
            fi
        fi
        return $_status
    else
        err "Unknown downloader" # should not reach here
    fi
}

unzip() {
    local _file
    _file="$1"
    local _dir
    _dir="$2"

    tar -xzf "$_file" -C "$_dir"
}

get_default_install_dir() {
    printf %s "${HOME}/.temporalio"
}

get_install_dir() {
    local _dir
    _dir="$(get_default_install_dir)"

    while [ $# -gt 0 ]; do
        case "$1" in
        --dir)
            _dir="$2"
            shift
            ;;
        *) ;;

        esac
        shift
    done

    printf %s "$_dir"
}

prompt_for_path() {
    local _dirbin
    _dirbin="$1"
    local _source
    _source="export PATH=\"\\\$PATH:$_dirbin\" >> ~/.bashrc"

    say "Temporal CLI installed at $_dirbin/temporal"

    if echo ":$PATH:" | grep -q ":$_dirbin:"; then
        say "Start the server with: temporal server start-dev"
        say "Or start a workflow with: temporal workflow start"
        say "For usage, run: temporal --help"
    else
        say "For convenience, we recommend adding it to your PATH"
        say "If using bash, run echo $_source"
    fi
}

check_help_for() {
    local _arch
    local _cmd
    local _arg
    _arch="$1"
    shift
    _cmd="$1"
    shift

    local _category
    if "$_cmd" --help | grep -q 'For all options use the manual or "--help all".'; then
        _category="all"
    else
        _category=""
    fi

    case "$_arch" in

    *darwin*)
        if check_cmd sw_vers; then
            case $(sw_vers -productVersion) in
            10.*)
                # If we're running on macOS, older than 10.13, then we always
                # fail to find these options to force fallback
                if [ "$(sw_vers -productVersion | cut -d. -f2)" -lt 13 ]; then
                    # Older than 10.13
                    echo "Warning: Detected macOS platform older than 10.13"
                    return 1
                fi
                ;;
            11.*)
                # We assume Big Sur will be OK for now
                ;;
            *)
                # Unknown product version, warn and continue
                echo "Warning: Detected unknown macOS major version: $(sw_vers -productVersion)"
                echo "Warning TLS capabilities detection may fail"
                ;;
            esac
        fi
        ;;

    esac

    for _arg in "$@"; do
        if ! "$_cmd" --help $_category | grep -q -- "$_arg"; then
            return 1
        fi
    done

    true # not strictly needed
}

# Return cipher suite string specified by user, otherwise return strong TLS 1.2-1.3 cipher suites
# if support by local tools is detected. Detection currently supports these curl backends:
# GnuTLS and OpenSSL (possibly also LibreSSL and BoringSSL). Return value can be empty.
get_ciphersuites_for_curl() {
    if [ -n "${TEMPORAL_TLS_CIPHERSUITES-}" ]; then
        # user specified custom cipher suites, assume they know what they're doing
        RETVAL="$TEMPORAL_TLS_CIPHERSUITES"
        return
    fi

    local _openssl_syntax
    _openssl_syntax="no"
    local _gnutls_syntax
    _gnutls_syntax="no"
    local _backend_supported
    _backend_supported="yes"
    if curl -V | grep -q ' OpenSSL/'; then
        _openssl_syntax="yes"
    elif curl -V | grep -iq ' LibreSSL/'; then
        _openssl_syntax="yes"
    elif curl -V | grep -iq ' BoringSSL/'; then
        _openssl_syntax="yes"
    elif curl -V | grep -iq ' GnuTLS/'; then
        _gnutls_syntax="yes"
    else
        _backend_supported="no"
    fi

    local _args_supported
    _args_supported="no"
    if [ "$_backend_supported" = "yes" ]; then
        # "unspecified" is for arch, allows for possibility old OS using macports, homebrew, etc.
        if check_help_for "notspecified" "curl" "--tlsv1.2" "--ciphers" "--proto"; then
            _args_supported="yes"
        fi
    fi

    local _cs
    _cs=""
    if [ "$_args_supported" = "yes" ]; then
        if [ "$_openssl_syntax" = "yes" ]; then
            _cs=$(get_strong_ciphersuites_for "openssl")
        elif [ "$_gnutls_syntax" = "yes" ]; then
            _cs=$(get_strong_ciphersuites_for "gnutls")
        fi
    fi

    RETVAL="$_cs"
}

# Return cipher suite string specified by user, otherwise return strong TLS 1.2-1.3 cipher suites
# if support by local tools is detected. Detection currently supports these wget backends:
# GnuTLS and OpenSSL (possibly also LibreSSL and BoringSSL). Return value can be empty.
get_ciphersuites_for_wget() {
    if [ -n "${TEMPORAL_TLS_CIPHERSUITES-}" ]; then
        # user specified custom cipher suites, assume they know what they're doing
        RETVAL="$TEMPORAL_TLS_CIPHERSUITES"
        return
    fi

    local _cs
    _cs=""
    if wget -V | grep -q '\-DHAVE_LIBSSL'; then
        # "unspecified" is for arch, allows for possibility old OS using macports, homebrew, etc.
        if check_help_for "notspecified" "wget" "TLSv1_2" "--ciphers" "--https-only" "--secure-protocol"; then
            _cs=$(get_strong_ciphersuites_for "openssl")
        fi
    elif wget -V | grep -q '\-DHAVE_LIBGNUTLS'; then
        # "unspecified" is for arch, allows for possibility old OS using macports, homebrew, etc.
        if check_help_for "notspecified" "wget" "TLSv1_2" "--ciphers" "--https-only" "--secure-protocol"; then
            _cs=$(get_strong_ciphersuites_for "gnutls")
        fi
    fi

    RETVAL="$_cs"
}

# Return strong TLS 1.2-1.3 cipher suites in OpenSSL or GnuTLS syntax. TLS 1.2
# excludes non-ECDHE and non-AEAD cipher suites. DHE is excluded due to bad
# DH params often found on servers (see RFC 7919). Sequence matches or is
# similar to Firefox 68 ESR with weak cipher suites disabled via about:config.
# $1 must be openssl or gnutls.
get_strong_ciphersuites_for() {
    if [ "$1" = "openssl" ]; then
        # OpenSSL is forgiving of unknown values, no problems with TLS 1.3 values on versions that don't support it yet.
        echo "TLS_AES_128_GCM_SHA256:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_256_GCM_SHA384:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384"
    elif [ "$1" = "gnutls" ]; then
        # GnuTLS isn't forgiving of unknown values, so this may require a GnuTLS version that supports TLS 1.3 even if wget doesn't.
        # Begin with SECURE128 (and higher) then remove/add to build cipher suites. Produces same 9 cipher suites as OpenSSL but in slightly different order.
        echo "SECURE128:-VERS-SSL3.0:-VERS-TLS1.0:-VERS-TLS1.1:-VERS-DTLS-ALL:-CIPHER-ALL:-MAC-ALL:-KX-ALL:+AEAD:+ECDHE-ECDSA:+ECDHE-RSA:+AES-128-GCM:+CHACHA20-POLY1305:+AES-256-GCM"
    fi
}

main "$@" || exit 1
