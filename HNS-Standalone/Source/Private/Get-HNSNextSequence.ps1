function Get-HNSNextSequence {
    <#
    .SYNOPSIS
        Gets the next available sequence number for a template
    .DESCRIPTION
        Finds the next sequence number, considering gaps in the sequence
    .PARAMETER TemplateID
        The ID of the template
    .PARAMETER Start
        Starting number to search from
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [int]$TemplateID,
        
        [Parameter()]
        [int]$Start
    )
    
    try {
        $template = $script:Templates | Where-Object { $_.ID -eq $TemplateID }
        if (-not $template) {
            throw "Template with ID $TemplateID not found"
        }
        
        if (-not $PSBoundParameters.ContainsKey('Start')) {
            $Start = $template.SequenceStart
        }
        
        # Get all non-released hostnames for this template
        $hostnames = $script:Hostnames | Where-Object { 
            $_.TemplateID -eq $TemplateID -and 
            $_.Status -ne [HostnameStatus]::Released 
        }
        
        $usedSequences = $hostnames | ForEach-Object { $_.SequenceNum } | Sort-Object
        
        # Find first gap or next number
        $current = $Start
        foreach ($used in $usedSequences) {
            if ($used -gt $current) {
                return $current
            }
            $current = $used + $template.SequenceIncrement
        }
        
        return $current
    }
    catch {
        throw "Failed to get next sequence number: $_"
    }
}