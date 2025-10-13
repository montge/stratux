#!/bin/bash
VER=$(git describe --tags --abbrev=0 2>/dev/null)
if [ -z "$VER" ]; then
	# No tags found, use date and commit hash (must start with digit for Debian)
	VER="0.0.$(date +%Y%m%d)-$(git rev-parse --short HEAD)"
fi
if [ "${VER:0:1}" == "v" ]; then
	VER=${VER:1}
fi
echo $VER
