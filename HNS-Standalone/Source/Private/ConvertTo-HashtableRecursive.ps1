function ConvertTo-HashtableRecursive {
    <#
    .SYNOPSIS
        Converts PSCustomObject to hashtable recursively
    .DESCRIPTION
        Recursively converts PSCustomObject instances to hashtables,
        handling nested objects and arrays
    .PARAMETER InputObject
        The object to convert
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        $InputObject
    )
    
    if ($null -eq $InputObject) {
        return $null
    }
    
    if ($InputObject -is [System.Collections.IEnumerable] -and $InputObject -isnot [string]) {
        # Handle arrays
        $array = @()
        foreach ($item in $InputObject) {
            $array += ConvertTo-HashtableRecursive $item
        }
        return $array
    }
    
    if ($InputObject.PSObject.TypeNames -contains 'System.Management.Automation.PSCustomObject') {
        # Convert PSCustomObject to hashtable
        $hashtable = @{}
        $InputObject.PSObject.Properties | ForEach-Object {
            $hashtable[$_.Name] = ConvertTo-HashtableRecursive $_.Value
        }
        return $hashtable
    }
    
    # Return primitive types as-is
    return $InputObject
}