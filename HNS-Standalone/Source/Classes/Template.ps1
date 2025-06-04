class Template {
    [int] $ID
    [string] $Name
    [string] $Description
    [int] $MaxLength
    [TemplateGroup[]] $Groups
    [int] $SequenceStart
    [int] $SequenceLength
    [bool] $SequencePadding
    [int] $SequenceIncrement
    [int] $SequencePosition
    [string] $CreatedBy
    [datetime] $CreatedAt
    [datetime] $UpdatedAt
    [bool] $IsActive
    [hashtable] $Metadata
    
    # Default constructor
    Template() {
        $this.Groups = @()
        $this.SequenceStart = 1
        $this.SequenceIncrement = 1
        $this.SequencePadding = $true
        $this.IsActive = $true
        $this.CreatedAt = Get-Date
        $this.UpdatedAt = Get-Date
        $this.Metadata = @{}
    }
    
    # Constructor with hashtable
    Template([hashtable]$Properties) {
        $this.Groups = @()
        $this.SequenceStart = 1
        $this.SequenceIncrement = 1
        $this.SequencePadding = $true
        $this.IsActive = $true
        $this.CreatedAt = Get-Date
        $this.UpdatedAt = Get-Date
        $this.Metadata = @{}
        
        foreach ($key in $Properties.Keys) {
            if ($key -eq 'Groups' -and $Properties[$key]) {
                $this.Groups = $Properties[$key] | ForEach-Object {
                    if ($_ -is [TemplateGroup]) { $_ }
                    else { [TemplateGroup]::new($_) }
                }
            }
            elseif ($null -ne $this.PSObject.Properties[$key]) {
                $this.$key = $Properties[$key]
            }
        }
    }
    
    # Validate the template configuration
    [bool] Validate() {
        # Check if groups cover all positions
        $positions = $this.Groups | ForEach-Object { $_.Position } | Sort-Object
        $expectedPositions = 1..$this.Groups.Count
        
        if (Compare-Object $positions $expectedPositions) {
            throw "Template groups must have continuous positions from 1 to $($this.Groups.Count)"
        }
        
        # Check total length
        $totalLength = ($this.Groups | Measure-Object -Property Length -Sum).Sum
        if ($totalLength -ne $this.MaxLength) {
            throw "Total group lengths ($totalLength) must equal MaxLength ($($this.MaxLength))"
        }
        
        # Check sequence position validity
        if ($this.SequencePosition -lt 1 -or $this.SequencePosition -gt $this.Groups.Count) {
            throw "SequencePosition must be between 1 and $($this.Groups.Count)"
        }
        
        # Check if sequence group exists and is correct type
        $sequenceGroup = $this.Groups | Where-Object { $_.Position -eq $this.SequencePosition }
        if ($sequenceGroup.ValidationType -ne [ValidationType]::Sequence) {
            throw "Group at SequencePosition must have ValidationType 'Sequence'"
        }
        
        if ($sequenceGroup.Length -ne $this.SequenceLength) {
            throw "Sequence group length must match SequenceLength"
        }
        
        return $true
    }
    
    # Get group by name
    [TemplateGroup] GetGroupByName([string]$Name) {
        return $this.Groups | Where-Object { $_.Name -eq $Name } | Select-Object -First 1
    }
    
    # Get group by position
    [TemplateGroup] GetGroupByPosition([int]$Position) {
        return $this.Groups | Where-Object { $_.Position -eq $Position } | Select-Object -First 1
    }
    
    # Generate example hostname
    [string] GenerateExample() {
        $example = ""
        foreach ($group in $this.Groups | Sort-Object Position) {
            switch ($group.ValidationType) {
                ([ValidationType]::Fixed) {
                    $example += $group.ValidationValue
                }
                ([ValidationType]::List) {
                    $values = $group.ValidationValue -split ','
                    $example += $values[0]
                }
                ([ValidationType]::Regex) {
                    $example += 'X' * $group.Length
                }
                ([ValidationType]::Sequence) {
                    $num = $this.SequenceStart.ToString()
                    if ($this.SequencePadding) {
                        $num = $num.PadLeft($this.SequenceLength, '0')
                    }
                    $example += $num
                }
            }
        }
        return $example
    }
    
    # Generate hostname with specific sequence number
    [string] GenerateHostname([int]$SequenceNumber) {
        $hostname = ""
        foreach ($group in $this.Groups | Sort-Object Position) {
            switch ($group.ValidationType) {
                ([ValidationType]::Fixed) {
                    $hostname += $group.ValidationValue
                }
                ([ValidationType]::Sequence) {
                    $num = $SequenceNumber.ToString()
                    if ($this.SequencePadding) {
                        $num = $num.PadLeft($this.SequenceLength, '0')
                    }
                    $hostname += $num
                }
                ([ValidationType]::List) {
                    # For list types, we need parameters - this method assumes fixed values
                    $values = $group.ValidationValue -split ','
                    if ($values.Count -gt 0) {
                        $hostname += $values[0].Trim()
                    }
                }
                ([ValidationType]::Regex) {
                    # For regex types, we need parameters - this method assumes pattern match
                    $hostname += 'X' * $group.Length
                }
            }
        }
        return $hostname
    }
    
    # Generate hostname with parameters and sequence number
    [string] GenerateHostname([int]$SequenceNumber, [hashtable]$Parameters) {
        $hostname = ""
        foreach ($group in $this.Groups | Sort-Object Position) {
            $value = ""
            switch ($group.ValidationType) {
                ([ValidationType]::Fixed) {
                    $value = $group.ValidationValue
                }
                ([ValidationType]::Sequence) {
                    $value = $SequenceNumber.ToString()
                    if ($this.SequencePadding) {
                        $value = $value.PadLeft($this.SequenceLength, '0')
                    }
                }
                ([ValidationType]::List) {
                    if ($Parameters.ContainsKey($group.Name)) {
                        $value = $Parameters[$group.Name]
                    } elseif ($group.IsRequired) {
                        throw "Required parameter '$($group.Name)' not provided. Valid values: $($group.ValidationValue)"
                    }
                }
                ([ValidationType]::Regex) {
                    if ($Parameters.ContainsKey($group.Name)) {
                        $value = $Parameters[$group.Name]
                    } elseif ($group.IsRequired) {
                        throw "Required parameter '$($group.Name)' not provided. Must match pattern: $($group.ValidationValue)"
                    }
                }
            }
            
            # Validate value if provided
            if ($value -and -not $group.Validate($value)) {
                throw "Value '$value' is not valid for group '$($group.Name)'. $($group.ToString())"
            }
            
            $hostname += $value
        }
        return $hostname
    }
    
    # ToString override
    [string] ToString() {
        return "$($this.Name) (MaxLength: $($this.MaxLength), Groups: $($this.Groups.Count), Active: $($this.IsActive))"
    }
    
    # Clone method
    [Template] Clone() {
        $clone = [Template]::new(@{
            Name = "$($this.Name)_Copy"
            Description = $this.Description
            MaxLength = $this.MaxLength
            SequenceStart = $this.SequenceStart
            SequenceLength = $this.SequenceLength
            SequencePadding = $this.SequencePadding
            SequenceIncrement = $this.SequenceIncrement
            SequencePosition = $this.SequencePosition
            IsActive = $this.IsActive
            Metadata = $this.Metadata.Clone()
        })
        
        $clone.Groups = $this.Groups | ForEach-Object { $_.Clone() }
        
        return $clone
    }
}