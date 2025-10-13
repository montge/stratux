#!/bin/bash
VER=$(git describe --tags --abbrev=0 2>/dev/null)
if [ -z "$VER" ]; then
	# No tags found, use commit hash and date
	VER="dev-$(git rev-parse --short HEAD)-$(date +%Y%m%d)"
fi
if [ "${VER:0:1}" == "v" ]; then
	VER=${VER:1}
fi
echo $VER
