function New-HNSTemplate {
    <#
    .SYNOPSIS
        Creates a new hostname template
    .DESCRIPTION
        Creates a new template for generating standardized hostnames with configurable groups and validation rules
    .PARAMETER Name
        The name of the template
    .PARAMETER Description
        A description of the template and its purpose
    .PARAMETER MaxLength
        Maximum length of hostnames generated from this template
    .PARAMETER Groups
        Array of template groups defining the hostname structure
    .PARAMETER SequenceStart
        Starting number for sequences (default: 1)
    .PARAMETER SequenceLength
        Length of sequence numbers (default: 3)
    .PARAMETER SequencePadding
        Whether to pad sequences with zeros (default: $true)
    .PARAMETER SequenceIncrement
        Increment value for sequences (default: 1)
    .PARAMETER SequencePosition
        Position of the sequence group in the hostname (1-based)
    .PARAMETER Metadata
        Additional metadata to store with the template
    .EXAMPLE
        $groups = @(
            @{Name='Unit'; Length=3; Position=1; ValidationType='List'; ValidationValue='ABC,DEF,GHI'},
            @{Name='Type'; Length=1; Position=2; ValidationType='Fixed'; ValidationValue='S'},
            @{Name='Sequence'; Length=3; Position=3; ValidationType='Sequence'; ValidationValue=''}
        )
        New-HNSTemplate -Name "Simple Server" -Description "Basic server naming" -MaxLength 7 -Groups $groups -SequencePosition 3
        
        Creates a template that generates hostnames like "ABCS001", "ABCS002", etc.
    .EXAMPLE
        # Create Epiroc-style template
        $groups = @(
            @{Name='unit_code'; Length=3; Position=1; ValidationType='List'; ValidationValue='SFD,SFS,DAL,COP,JAC,AVT'},
            @{Name='type'; Length=1; Position=2; ValidationType='Fixed'; ValidationValue='S'},
            @{Name='provider'; Length=1; Position=3; ValidationType='List'; ValidationValue='M,A,G,H,O,E'},
            @{Name='region'; Length=4; Position=4; ValidationType='List'; ValidationValue='WEEU,NOEU,EAAS,SEAS,CEUS,EAUS,WEUS'},
            @{Name='environment'; Length=1; Position=5; ValidationType='List'; ValidationValue='P,Q,D,T,S'},
            @{Name='function'; Length=2; Position=6; ValidationType='List'; ValidationValue='AS,DB,WS,FS,DC,TS'},
            @{Name='sequence'; Length=3; Position=7; ValidationType='Sequence'; ValidationValue=''}
        )
        New-HNSTemplate -Name "Epiroc VM Standard" -MaxLength 15 -Groups $groups -SequencePosition 7
    #>
    [CmdletBinding(SupportsShouldProcess)]
    param(
        [Parameter(Mandatory)]
        [ValidateNotNullOrEmpty()]
        [string]$Name,
        
        [Parameter()]
        [string]$Description = "",
        
        [Parameter(Mandatory)]
        [ValidateRange(1, 63)]
        [int]$MaxLength,
        
        [Parameter(Mandatory)]
        [ValidateNotNullOrEmpty()]
        [hashtable[]]$Groups,
        
        [Parameter()]
        [ValidateRange(0, 9999)]
        [int]$SequenceStart = 1,
        
        [Parameter()]
        [ValidateRange(1, 4)]
        [int]$SequenceLength = 3,
        
        [Parameter()]
        [bool]$SequencePadding = $true,
        
        [Parameter()]
        [ValidateRange(1, 100)]
        [int]$SequenceIncrement = 1,
        
        [Parameter(Mandatory)]
        [ValidateRange(1, 20)]
        [int]$SequencePosition,
        
        [Parameter()]
        [hashtable]$Metadata = @{}
    )
    
    begin {
        Test-HNSInitialized -Throw
        Write-Verbose "Creating new template: $Name"
    }
    
    process {
        try {
            # Check if template name already exists
            if ($script:Templates | Where-Object { $_.Name -eq $Name }) {
                throw "Template with name '$Name' already exists"
            }
            
            # Create new template
            $template = [Template]::new(@{
                ID = $script:Configuration.GetNextTemplateID()
                Name = $Name
                Description = $Description
                MaxLength = $MaxLength
                SequenceStart = $SequenceStart
                SequenceLength = $SequenceLength
                SequencePadding = $SequencePadding
                SequenceIncrement = $SequenceIncrement
                SequencePosition = $SequencePosition
                CreatedBy = $script:Configuration.DefaultUser
                Metadata = $Metadata
            })
            
            # Add groups
            $groupId = 1
            foreach ($groupDef in $Groups) {
                # Ensure required properties
                if (-not $groupDef.ContainsKey('IsRequired')) {
                    $groupDef['IsRequired'] = $true
                }
                
                $group = [TemplateGroup]::new($groupDef)
                $group.ID = $groupId++
                $group.TemplateID = $template.ID
                $template.Groups += $group
            }
            
            # Validate template
            $template.Validate()
            
            if ($PSCmdlet.ShouldProcess($Name, "Create template")) {
                # Add to collection and save
                $script:Templates += $template
                Save-HNSData -Type Templates
                
                # Audit log
                Write-AuditLog -Action "CreateTemplate" -Details "Created template: $Name (ID: $($template.ID))"
                
                Write-Verbose "Template created successfully: $Name"
                return $template
            }
        }
        catch {
            Write-Error "Failed to create template: $_"
            throw
        }
    }
}