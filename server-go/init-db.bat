@echo off
REM 一键初始化MySQL数据库脚本（Windows）
REM 请根据实际账号密码调整下方变量
set MYSQL_HOST=localhost
set MYSQL_PORT=3306
set MYSQL_USER=root
set MYSQL_PWD=password
set DB_NAME=lababa

echo 创建数据库 %DB_NAME%（若不存在）...
mysql -h%MYSQL_HOST% -P%MYSQL_PORT% -u%MYSQL_USER% -p%MYSQL_PWD% -e "CREATE DATABASE IF NOT EXISTS %DB_NAME% CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
if %errorlevel% neq 0 (
  echo 连接失败，请检查MySQL服务是否启动或账号密码是否正确
  pause
  exit /b 1
)
echo 数据库准备完成，请在.env中确认DSN后启动Go服务
pause