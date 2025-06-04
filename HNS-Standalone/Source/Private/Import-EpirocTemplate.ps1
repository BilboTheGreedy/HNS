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
                ValidationValue = 'M,A,G,H,O,E'
            },
            @{
                Name = 'region'
                Length = 4
                Position = 4
                IsRequired = $true
                ValidationType = 'List'
                ValidationValue = 'WEEU,NOEU,SWNA,NCUS,SCUS,EAUS,WEUS,EAAS,SEAS,SOAS,JAPA,AUSE,BRSO,SAFE,GENG,MXCE,INCA,AUCE,AUCH,AUNW,WCUS,SWUS,CEUS,NEUS,NEEU,CEEU,UKSO,AZWE,AZNO,AZEA,AZSE,AZSC,AZNE,AZWC,AZUS,AZEU,AZAP,AZZA,AZAU,AZIN,AZCH,AZJP,AZBR,AZCA,AZUK,AZFR,AZGE,AZIT,AZKO,AZPO,AZSG,AZSF,AZUA,AZGA,AZVA,AZNJ,AZBE,AZBA,AZML,GOOG,USVG,USOH,USTX,USCA,USNJ,USNY,USIL,USFL,USWA,USMA,USPA,USMI,USGA,USNC,USVA,USMD,USCT,USAZ,USNV,USCO,USOR,USUT,USID,USMT,USWY,USNM,USND,USSD,USNE,USIA,USMN,USWI,USMO,USAR,USLA,USMS,USAL,USKY,USTN,USIN,USSC,USWV,USME,USVT,USNH,USRI,USDE,ALOH,OCQC,OCBC,OCON,OCMB,OCAB,OCSK,OCNB,OCNS,OCNL,OCPE,OCYK,OCNT,OCNU,HV01,HV02,HV03,HV04,HV05,ON00,ON01,ON02,ON03,ON04,ON05,ON06,ON07,ON08,ON09,ON10,ON11,ON12,ON13,ON14,ON15,ON16,ON17,ON18,ON19,ON20'
            },
            @{
                Name = 'environment'
                Length = 1
                Position = 5
                IsRequired = $true
                ValidationType = 'List'
                ValidationValue = 'P,Q,D,T,S,N,U,A,B,C,E,I'
            },
            @{
                Name = 'function'
                Length = 2
                Position = 6
                IsRequired = $true
                ValidationType = 'List'
                ValidationValue = 'AP,AS,BA,BE,BI,CA,CI,CM,CS,CT,DB,DC,DH,DI,DS,DV,EX,FS,FT,FW,GW,HV,IS,JS,JH,KM,LB,LI,MB,MD,MG,MQ,MS,MT,MW,NA,NS,PB,PC,PI,PK,PR,PS,PX,QA,RB,RD,RP,RH,RS,SC,SM,SQ,SR,ST,TE,TS,VC,VS,VD,VH,VM,VP,WA,WB,WF,WS,XX'
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