@echo off
REM 运行 Go 工具 清除所所有数据库
echo Running Go tool...
cd sdk/test/tools
go run main.go

cd ../../../server/registersvr

REM 启动注册服务
set ports=20001 20002 20003
for %%p in (%ports%) do (
    echo Starting register service on port %%p...
    start go run main.go -port %%p
)

cd ../../server/healthsvr

REM 启动健康检查服务
set health_ports=30001 30002 30003
for %%p in (%health_ports%) do (
    echo Starting health check service on port %%p...
    start go run main.go -port %%p
)

cd ../../server/discoversvr

REM 启动发现服务
set discover_ports=40001 40002 40003
for %%p in (%discover_ports%) do (
    echo Starting discover service on port %%p...
    start go run main.go -port %%p
)

cd ../../sdk/test

REM 运行 Go 测试
echo Running Go tests...
go test -v TestForAll_test.go

echo All commands executed.
pause