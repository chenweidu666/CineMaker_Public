.PHONY: test test-coverage test-verbose tidy build

# 运行所有测试
test:
	go test ./... -v

# 运行测试并生成覆盖率报告
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# 仅显示覆盖率百分比
test-coverage-percent:
	go test ./... -cover

# 整理依赖
tidy:
	go mod tidy

# 构建项目
build:
	go build -o main .

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
vet:
	go vet ./...

# 完整检查流程
check: fmt vet test
