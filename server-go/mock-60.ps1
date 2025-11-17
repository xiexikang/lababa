# 生成60条Mock数据（createdAt倒序，最新在前）
$base = "http://localhost:8081"
function Invoke-Post($url,$body){
    Invoke-RestMethod -Uri "$base$url" -Method Post -ContentType 'application/json' -Body ($body|ConvertTo-Json -Compress)
}
# 1. 创建3个用户
$uids = @()
@(
    @{code="u001";nickName="阿明";avatarUrl="https://example.com/a.jpg"},
    @{code="u002";nickName="小布";avatarUrl="https://example.com/b.jpg"},
    @{code="u003";nickName="陈晨";avatarUrl="https://example.com/c.jpg"}
)|%{
    $r = Invoke-Post '/api/auth/weapp' $_
    $uids += $r.user.id
    Write-Host "User $($_.nickName) => $($r.user.id)"
}
# 2. 生成60条记录，时间戳倒序（最新在前）
$colors = @("brown","green","yellow")
$shapes = @("banana","apple","grape")
$status = @("normal","good","bad")
$amount = @("little","moderate","lot")
$rand = [Random]::new()
$now = [DateTimeOffset]::UtcNow
0..59 | %{
    $end = $now.AddMinutes(-$_ * 30).ToUnixTimeMilliseconds()
    $dur = $rand.Next(60,1200)
    $rec = @{
        userId = $uids[$rand.Next(0,3)]
        endTime = $end
        duration = $dur
        color = $colors[$rand.Next(0,3)]
        shape = $shapes[$rand.Next(0,3)]
        status = $status[$rand.Next(0,3)]
        amount = $amount[$rand.Next(0,3)]
        note = "Mock-$_"
    }
    Invoke-Post '/api/records/create' $rec | Out-Null
    if($_ % 10 -eq 0){ Write-Host "Inserted $_ /60" }
}
Write-Host "Done: 60 records inserted in createdAt DESC order"