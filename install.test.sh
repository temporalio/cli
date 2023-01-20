#!/bin/sh

set -ex

sh ./install.sh
. "$HOME"/.temporalio/env
temporal -v
