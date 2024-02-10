# Initialize an empty array to store all blob details
$all_blob_details = @()

# Initialize the marker
$marker = $null

# Loop until there are no more blobs to retrieve
do {
    # Run Azure CLI command to get blob list with marker
    $blob_list = az storage blob list --account-name <storage_account_name> --container-name <container_name> --query "[].{name:name, last_modified:lastModified, size:properties.contentLength}" --marker $marker -o json | ConvertFrom-Json

    # Append the batch of blobs to the array
    $all_blob_details += $blob_list

    # If the batch is not empty, update the marker for the next batch
    if ($blob_list) {
        $marker = $blob_list[-1].name
    }
} while ($blob_list)

# Convert the array to JSON format
$blob_json = $all_blob_details | ConvertTo-Json -Depth 100

# Append the JSON to blob_details.json
$blob_json | Out-File -FilePath "blob_details.json" -Append
