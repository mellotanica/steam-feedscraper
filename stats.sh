#!/bin/bash

for f in "$(dirname "${BASH_SOURCE[0]}")"/feeds_cache_*; do
	echo "$(basename "$f"): $(jq ". | length" "$f")"
done
