function Import-EpirocTemplate {
    <#
    .SYNOPSIS
        Imports the Epiroc VM Standard template
    .DESCRIPTION
        Creates the predefined Epiroc VM naming standard template with all validation rules
    #>
    [CmdletBinding()]
    param()
    
    try {
        $groups = @(
            @{
                Name = 'unit_code'
                Length = 3
                Position = 1
                IsRequired = $true
                ValidationType = 'List'
                ValidationValue = 'SFD,SFS,DAL,COP,JAC,RDT,AVT,RIG,FRD,EST,FAG,CER,ALL,HPI,NJB,OUL,ARE,CHR,CLK,FRB,GAR,GHA,KAL,LAN,NAC,OEB,SKE,UDD,ZED,GDE,JHA,NAN,ROS,ITS,LEG,CHI,BRA,USA,PER,MEX,CAN,AUS,IND,AFR,RUS,MON,KAZ,EUR,GLO,MAL,THA,KOR,JAP,NIG,GEO,FIN'
            },
            @{
                Name = 'type'
                Length = 1
                Position = 2
                IsRequired = $true
                ValidationType = 'Fixed'
                ValidationValue = 'S'
            },
            @{
                Name = 'provider'
                Length = 1
                Position = 3
                IsRequired = $true
                ValidationType = 'List'
                ValidationValue = 'E,M,A,G'
            },
            @{
                Name = 'region'
                Length = 4
                Position = 4
                IsRequired = $true
                ValidationType = 'List'
                ValidationValue = 'OCAU,ASCN,ASIN,ASSG,EUSE,AFZA,NAUS,ACL1,ACL2,AUEA,SEAU,INCE,SHA1,BJB1,EAAS,JPEA,JPWE,KRCE,KRSO,INSO,SEAS,INWE,FRCE,DEWC,ISCE,ITNO,NOEU,NWEA,PLCE,QACE,SANO,SAWE,SWCE,SZNO,UANO,UKSO,UKWE,WEEU,BRSO,CCAN,ECAN,CEUS,CLCE,EUS1,EUS2,NCUS,SCUS,WCUS,WUS1,WUS2,WUS3'
            },
            @{
                Name = 'environment'
                Length = 1
                Position = 5
                IsRequired = $true
                ValidationType = 'List'
                ValidationValue = 'P,T,D'
            },
            @{
                Name = 'function'
                Length = 2
                Position = 6
                IsRequired = $true
                ValidationType = 'List'
                ValidationValue = 'AS,BU,DB,DC,FS,IS,MG,PO,WS,PV,CC,UC,OT'
            },
            @{
                Name = 'sequence'
                Length = 3
                Position = 7
                IsRequired = $true
                ValidationType = 'Sequence'
                ValidationValue = ''
            }
        )
        
        $template = New-HNSTemplate -Name "Epiroc VM Standard" `
            -Description "Epiroc 15-character VM naming standard: [Unit(3)][Type(1)][Provider(1)][Region(4)][Env(1)][Function(2)][Seq(3)]" `
            -MaxLength 15 `
            -Groups $groups `
            -SequencePosition 7 `
            -SequenceStart 1 `
            -SequenceLength 3 `
            -SequencePadding $true `
            -Metadata @{
                Version = "1.0"
                Standard = "Epiroc VM Naming Standard"
                Documentation = "https://epiroc.sharepoint.com/sites/vm-naming-standard"
            }
        
        Write-Verbose "Imported Epiroc VM Standard template"
        return $template
    }
    catch {
        Write-Error "Failed to import Epiroc template: $_"
        throw
    }
}