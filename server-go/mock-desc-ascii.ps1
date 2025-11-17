# 重新生成Mock数据（按createdAt倒序）
$base = "http://localhost:8081"
function Invoke-Post($url,$body){
    Invoke-RestMethod -Uri "$base$url" -Method Post -ContentType 'application/json' -Body ($body|ConvertTo-Json -Compress)
}
# 创建3个用户
$uids = @()
@(
    @{code="u001";nickName="阿明";avatarUrl="https://example.com/a.jpg"},
    @{code="u002";nickName="小布";avatarUrl="https://example.com/b.jpg"},
    @{code="u003";nickName="陈晨";avatarUrl="https://example.com/c.jpg"}
)|%{
    $r = Invoke-Post '/api/auth/weapp' $_
    $uids += $r.user.id
    Write-Host "用户 $($_.nickName) => $($r.user.id)"
}
# 按时间倒序写入60条记录（最新在前）
$colors = @("brown","green","yellow")
$shapes = @("banana","apple","grape")
$status = @("normal","good","bad")
$amount = @("little","moderate","lot")
$rand = [Random]::new()
# 先生成时间戳数组（倒序）
$now = [DateTimeOffset]::UtcNow
$tsList = @(0..59 | % { $now.AddMinutes(-$_ * 30).ToUnixTimeMilliseconds() })
$tsList | %{
    $end = $_
    $dur = $rand.Next(60,1200)
    $rec = @{
        userId    = $uids[$rand.Next(0,3)]
        endTime   = $end
        duration  = $dur
        color     = $colors[$rand.Next(0,3)]
        shape     = $shapes[$rand.Next(0,3)]
        status    = $status[$rand.Next(0,3)]
        amount    = $amount[$rand.Next(0,3)]
        note      = "Mock"
    }
    Invoke-Post '/api/records/create' $rec | Out-Null
}
Write-Host "60条记录已按createdAt倒序写入完成！"