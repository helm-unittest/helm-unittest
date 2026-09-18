param(
  [Parameter(ValueFromRemainingArguments = $true)]
  [string[]]$ExtraArgs
)

$inputText = [Console]::In.ReadToEnd()
$expression = $ExtraArgs | Where-Object { $_ -match '^\.metadata\.annotations\.[A-Za-z0-9._-]+=\".*\"$' } | Select-Object -First 1
if (-not $expression) {
  [Console]::Out.Write($inputText)
  exit 0
}

if ($expression -notmatch '^\.metadata\.annotations\.([A-Za-z0-9._-]+)=\"(.*)\"$') {
  [Console]::Out.Write($inputText)
  exit 0
}

$key = $Matches[1]
$value = $Matches[2]
$hasKey = [regex]::IsMatch($inputText, "(?m)^\\s*$([regex]::Escape($key)):\\s*")
$lines = $inputText -split "`n", -1
$outLines = New-Object System.Collections.Generic.List[string]

foreach ($line in $lines) {
  $replacedLine = [regex]::Replace($line, "^(\\s*)$([regex]::Escape($key)):\\s*.*$", "`$1$key: `\"$value`\"")
  $outLines.Add($replacedLine)

  if (-not $hasKey -and $line -match '^\s*annotations:\s*$') {
    $indent = ([regex]::Match($line, '^\s*')).Value
    $outLines.Add("$indent  $key: `\"$value`\"")
  }
}

[Console]::Out.Write(($outLines -join "`n"))
