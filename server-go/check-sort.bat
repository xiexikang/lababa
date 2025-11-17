# 验证排序：取最新3条
cd /d "%~dp0"
curl -s "http://localhost:8081/api/index/list?limit=3" > resp.json & type resp.json