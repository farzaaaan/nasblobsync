# Initialize an empty array to store all blob details
$all_blob_details = @()

# Initialize the marker
$marker = $null

# Run Azure CLI command to get blob list with marker
do {
    # Run Azure CLI command to get blob list with marker and nextMarker
    $blob_list = az storage blob list --account-name <storage_account_name> --container-name <container_name> --query "[].{name:name, last_modified:lastModified, size:properties.contentLength, nextMarker:nextMarker}" --marker $marker --show-next-marker -o json | ConvertFrom-Json

    # If the blob list is not empty, append it to the array
    if ($blob_list) {
        # Remove the nextMarker property from the last item in the blob list
        $last_index = $blob_list.Count - 1
        $blob_list[$last_index].PSObject.Properties.Remove('nextMarker') | Out-Null
        $all_blob_details += $blob_list
        $marker = $blob_list[-1].name
    }
} while ($blob_list)

# Convert the array to JSON format
$blob_json = $all_blob_details | ConvertTo-Json -Depth 100

# Write the JSON to blob_details.json
$blob_json | Out-File -FilePath "blob_details.json"
