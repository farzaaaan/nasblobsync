#!/bin/bash

# Run Azure CLI command to get blob list and store the output
blob_list=$(az storage blob list --account-name <storage_account_name> --container-name <container_name> --query "[].{name:name, last_modified:lastModified, size:properties.contentLength}" -o json)

# Extract blob names, last modified, and size from the JSON output and format as a map
blob_map="{"
while IFS= read -r line; do
    name=$(echo "$line" | jq -r '.name')
    last_modified=$(echo "$line" | jq -r '.last_modified')
    size=$(echo "$line" | jq -r '.size')
    blob_map="$blob_map \"$name\": {\"last_modified\": \"$last_modified\", \"size\": $size},"
done <<< "$blob_list"
blob_map="${blob_map%,} }" # Remove the trailing comma and close the map

# Print the blob map
echo "$blob_map"
