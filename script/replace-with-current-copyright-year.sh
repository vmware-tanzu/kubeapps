#!/usr/bin/env bash

# Copyright 2023 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# Get the current year
current_year=$(date +%Y)

# Get the list of unique changed files since the specified year
changed_files=$(git log --name-only --pretty=format: --since="${current_year}-01-01" --no-merges | sort -u)

# Set the suffix to be used in the replacement
copyright_suffix="the Kubeapps contributors"

# Function to perform the text replacement in a file
function replace_text_in_file() {
    local file="$1"
    local errors_log="errors.log"

    if [ -n "$file" ]; then
        # Replace "Copyright YYYY-YYYY the Kubeapps contributors" with "Copyright YYYY-CURRENT_YEAR the Kubeapps contributors"
        sed -E -i "s/(Copyright\s+)([0-9]{4})-([0-9]{4})\s+$copyright_suffix/\1\2-$current_year $copyright_suffix/g" "$file" || echo error in "$file" >>$errors_log

        # Replace "Copyright YYYY the Kubeapps contributors" with "Copyright YYYY-CURRENT_YEAR the Kubeapps contributors"
        sed -E -i "s/(Copyright\s+[0-9]{4})\s+$copyright_suffix/\1-$current_year $copyright_suffix/g" "$file" || echo error in "$file" >>$errors_log

        # Replace "Copyright CURRENT_YEAR-CURRENT_YEAR the Kubeapps contributors" with "Copyright CURRENT_YEAR the Kubeapps contributors"
        sed -E -i "s/(Copyright\s+)($current_year)-($current_year)\s+$copyright_suffix/\1$current_year $copyright_suffix/g" "$file" || echo error in "$file" >>$errors_log

    else
        echo "No file passed to 'replace_text_in_file' function"
    fi
}

# Iterate over the list of changed files and perform text replacement
echo "Replacing copyright year in files changed since $current_year:"
while IFS= read -r file; do
    replace_text_in_file "$file"
done <<<"$changed_files"

# Display a message indicating the replacements are done
echo "Copyright year replacement completed."
