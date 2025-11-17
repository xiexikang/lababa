# 一键Mock数据脚本（PowerShell）
# 在server-go目录执行：powershell -File mock-data.ps1

$base = "http://localhost:8081"

function Invoke-Post($url, $body) {
    Invoke-RestMethod -Uri "$base$url" -Method Post -ContentType 'application/json' -Body ($body | ConvertTo-Json -Compress)
}

# 1. 创建3个用户
$users = @(
    @{code="u001";nickName="阿明";avatarUrl="https://example.com/a.jpg"},
    @{code="u002";nickName="小布";avatarUrl="https://example.com/b.jpg"},
    @{code="u003";nickName="陈晨";avatarUrl="https://example.com/c.jpg"}
)
$userIds = @()
foreach($u in $users){
    $r = Invoke-Post '/api/auth/weapp' $u
    $userIds += $r.user.id
    Write-Host "✅ 用户 $($u.nickName) => $($r.user.id)"
}

# 2. 为每个用户随机生成20条记录（最近30天内）
$colors = @("brown","green","yellow")
$shapes = @("banana","apple","grape")
$status = @("normal","good","bad")
$amount = @("little","moderate","lot")

$rand = [Random]::new()
1..60 | % {
    $uid = $userIds[$rand.Next(0,3)]
    $end = [DateTimeOffset]::UtcNow.AddMinutes(-$rand.Next(0,30*24*60)).ToUnixTimeMilliseconds()
    $dur = $rand.Next(60,1200)
    $rec = @{
        userId = $uid
        endTime = $end
        duration = $dur
        color = $colors[$rand.Next(0,3)]
        shape = $shapes[$rand.Next(0,3)]
        status = $status[$rand.Next(0,3)]
        amount = $amount[$rand.Next(0,3)]
        note = "Mock #$_"
    }
    Invoke-Post '/api/records/create' $rec | Out-Null
    if($_ % 10 -eq 0){ Write-Host "已写入 $_ /60 条记录" }
}

Write-Host "🎉 Mock数据写入完成！"
Write-Host "接下来可访问："
Write-Host "  - 首页列表  http://localhost:8081/api/index/list"
Write-Host "  - 统计汇总  http://localhost:8081/api/statistics/summary"
Write-Host "  - 排行榜    http://localhost:8081/api/ranking/list"