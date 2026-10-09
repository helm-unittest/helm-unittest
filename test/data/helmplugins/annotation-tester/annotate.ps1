$inputText = $input | Out-String
Write-Debug "Input text: $inputText"
Write-Debug "Arguments: $args"
$expression = $args |
  Where-Object { $_ -match '^\.metadata\.annotations\.[A-Za-z0-9._-]+="[^"]*"$' } |
  Select-Object -First 1

if (-not $expression) {
  Write-Output $inputText
  exit 0
}

if ($expression -notmatch '^\.metadata\.annotations\.([A-Za-z0-9._-]+)="([^"]*)"$') {
  Write-Output $inputText
  exit 0
}

$key = $Matches[1]
$value = $Matches[2]
$hasKey = [regex]::IsMatch($inputText, '(?m)^[\s]*' + [regex]::Escape($key) + ':\s*')
$lines = $inputText -split "`r?`n"
$outLines = New-Object System.Collections.Generic.List[string]

foreach ($line in $lines) {
  if ($line -match '^(?<prefix>\s*)' + [regex]::Escape($key) + ':\s*.*$') {
    $prefix = $Matches['prefix']
    $outLines.Add($prefix + $key + ': "' + $value + '"')
  }
  else {
    $outLines.Add($line)
  }

  if (-not $hasKey -and $line -match '^\s*annotations:\s*$') {
    $indent = [regex]::Match($line, '^\s*').Value
    $outLines.Add($indent + '  ' + $key + ': "' + $value + '"')
  }
}

$outputText = $outLines -join "`n"
Write-Debug "Output text: $outputText"
Write-Output $outputText
