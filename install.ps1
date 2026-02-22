# cftunnel Windows 安装脚本
$ErrorActionPreference = "Stop"
$repo = "qingchencloud/cftunnel"
$installDir = "$env:LOCALAPPDATA\cftunnel"

$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "amd64" }
$url = "https://github.com/$repo/releases/latest/download/cftunnel_windows_$arch.zip"

Write-Host "正在下载 cftunnel (windows/$arch)..."
$tmp = New-TemporaryFile | Rename-Item -NewName { $_.Name + ".zip" } -PassThru
Invoke-WebRequest -Uri $url -OutFile $tmp.FullName

New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Expand-Archive -Path $tmp.FullName -DestinationPath $installDir -Force
Remove-Item $tmp.FullName

# 添加到用户 PATH（持久化 + 当前会话立即生效）
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path += ";$installDir"
    Write-Host "已添加 $installDir 到 PATH（当前会话立即生效）"
}

Write-Host "cftunnel 已安装到 $installDir\cftunnel.exe"
Write-Host "运行 cftunnel quick <端口> 开始使用"
