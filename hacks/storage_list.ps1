# Run Azure CLI command to get blob list and store the output
$blob_list = az storage blob list --account-name <storage_account_name> --container-name <container_name> --query "[].{name:name, last_modified:lastModified, size:properties.contentLength}" -o json | ConvertFrom-Json

# Initialize an empty hashtable to store blob details
$blob_map = @{}

# Iterate through each blob and construct the map
foreach ($blob in $blob_list) {
    $blob_map[$blob.name] = @{
        last_modified = $blob.last_modified
        size = $blob.size
    }
}

# Convert the hashtable to JSON format
$blob_json = $blob_map | ConvertTo-Json -Depth 100

# Write the JSON to blob_details.json
$blob_json | Out-File -FilePath "blob_details.json"
