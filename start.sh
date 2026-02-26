#!/bin/bash

# CineMaker 启动管理脚本
# 全部基于 Docker，不依赖宿主机的 Go / Node 环境
#
# 开发模式（dev）— 源码挂载 + air 热加载，改 .go 文件 ~5秒自动重编
# 部署模式（deploy）— 多阶段构建生产镜像，数据在 Docker Volume

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

COMPOSE_CMD=""
if docker compose version &> /dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
elif command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
else
    echo -e "${RED}错误: 未找到 docker compose 命令${NC}"
    exit 1
fi

BACKEND_PORT=${PORT:-5678}
FRONTEND_PORT=${FRONTEND_PORT:-3012}

# ==================== 工具函数 ====================

kill_host_port() {
    local port=$1
    local pids
    pids=$(lsof -ti :"$port" 2>/dev/null || true)
    if [ -n "$pids" ]; then
        echo -e "${YELLOW}停止占用端口 $port 的宿主机进程: $pids${NC}"
        echo "$pids" | xargs kill 2>/dev/null || true
        sleep 1
    fi
}

wait_healthy() {
    local name=$1
    local max_wait=${2:-90}
    local elapsed=0
    echo -ne "  等待 $name 就绪..."
    while [ $elapsed -lt $max_wait ]; do
        if curl -sf http://localhost:$BACKEND_PORT/health > /dev/null 2>&1; then
            echo -e " ${GREEN}✓${NC}"
            return 0
        fi
        sleep 2
        elapsed=$((elapsed + 2))
        echo -n "."
    done
    echo -e " ${RED}超时${NC}"
    return 1
}

show_access_info() {
    echo ""
    local ip
    ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    ip=${ip:-localhost}
    echo -e "  前端页面:  ${CYAN}http://${ip}:${FRONTEND_PORT}/${NC}"
    echo -e "  后端 API:  ${CYAN}http://${ip}:${BACKEND_PORT}/api/v1${NC}"
    echo -e "  健康检查:  ${CYAN}http://${ip}:${BACKEND_PORT}/health${NC}"
    echo ""
}

# ==================== 开发模式（热加载） ====================

dev_compose() {
    $COMPOSE_CMD -f docker-compose.dev.yml "$@"
}

do_dev_start() {
    echo ""
    echo -e "${CYAN}======= CineMaker 开发模式 (热加载) =======${NC}"
    echo ""

    if [ ! -f "configs/config.yaml" ]; then
        echo -e "${YELLOW}未检测到 configs/config.yaml，从模板创建...${NC}"
        cp configs/config.example.yaml configs/config.yaml
        echo -e "${GREEN}✓${NC} 已创建，请编辑 configs/config.yaml 配置 AI 密钥"
    fi
    mkdir -p data/storage tmp

    kill_host_port $BACKEND_PORT
    kill_host_port $FRONTEND_PORT

    local build_flag=""
    if [ "$1" = "--build" ] || [ "$1" = "-b" ]; then
        build_flag="--build"
        echo -e "${YELLOW}强制重新构建开发镜像...${NC}"
    fi

    echo -e "${GREEN}启动后端服务 (air 热加载)...${NC}"
    dev_compose up -d $build_flag cinemaker

    wait_healthy "后端" 90

    echo ""
    if lsof -ti :$FRONTEND_PORT &>/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} 前端已在运行 (端口 $FRONTEND_PORT)，跳过"
    else
        echo -e "${GREEN}启动前端 Dev Server (Vite HMR)...${NC}"
        dev_compose --profile with-frontend up -d frontend
        sleep 5
        if dev_compose ps frontend 2>/dev/null | grep -q "running"; then
            echo -e "${GREEN}✓${NC} 前端已启动"
        else
            echo -e "${YELLOW}提示: 前端容器启动中，稍等几秒即可访问${NC}"
        fi
    fi

    echo ""
    echo -e "${CYAN}======= 启动完成 =======${NC}"
    echo -e "${GREEN}  改 .go 文件后 air 会自动重编（~5秒），无需手动操作${NC}"
    echo -e "${GREEN}  改前端代码 Vite 会自动热更新${NC}"
    do_dev_status
}

do_dev_stop() {
    echo -e "${YELLOW}停止开发模式...${NC}"
    dev_compose --profile with-frontend down
    echo -e "${GREEN}✓${NC} 已停止"
}

do_dev_restart() {
    echo -e "${YELLOW}重启开发模式...${NC}"
    do_dev_stop
    sleep 1
    do_dev_start "$1"
}

do_dev_rebuild() {
    echo -e "${YELLOW}重新构建开发镜像并启动...${NC}"
    echo -e "${YELLOW}（仅在修改了 go.mod / Dockerfile.dev 时需要）${NC}"
    dev_compose --profile with-frontend down
    sleep 1
    do_dev_start "--build"
}

do_dev_status() {
    echo ""
    echo "========== 开发模式状态 =========="
    dev_compose --profile with-frontend ps 2>/dev/null || echo "(无运行容器)"
    echo "=================================="
    show_access_info
}

do_dev_logs() {
    local target=${1:-"cinemaker"}
    local follow_flag=""
    if [ "$2" = "-f" ] || [ "$2" = "--follow" ]; then
        follow_flag="-f"
    fi

    case $target in
        backend|b|cinemaker)
            dev_compose logs $follow_flag --tail 200 cinemaker
            ;;
        frontend|f)
            dev_compose logs $follow_flag --tail 200 frontend
            ;;
        all|a)
            dev_compose --profile with-frontend logs $follow_flag --tail 100
            ;;
        *)
            echo -e "${RED}未知目标: $target${NC}"
            echo "用法: $0 logs [backend|frontend|all] [-f]"
            ;;
    esac
}

# ==================== 部署模式（生产构建） ====================

deploy_compose() {
    $COMPOSE_CMD -f docker-compose.yml "$@"
}

do_deploy_start() {
    if [ ! -f ".env" ]; then
        echo -e "${YELLOW}未检测到 .env 文件，从模板创建...${NC}"
        cp .env.example .env
    fi

    kill_host_port $BACKEND_PORT

    echo -e "${GREEN}正在启动 CineMaker (部署模式)...${NC}"
    deploy_compose up -d --build
    echo ""
    deploy_compose ps
    echo ""
    local ip
    ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    echo -e "访问地址: ${CYAN}http://${ip:-localhost}:${BACKEND_PORT}${NC}"
}

do_deploy_stop() {
    echo -e "${YELLOW}停止部署模式...${NC}"
    deploy_compose down
    echo -e "${GREEN}✓${NC} 已停止"
}

do_deploy_rebuild() {
    echo -e "${YELLOW}重新构建部署镜像...${NC}"
    deploy_compose down
    deploy_compose build --no-cache
    deploy_compose up -d
    echo -e "${GREEN}✓${NC} 已重建并启动"
    deploy_compose ps
}

# ==================== 帮助 ====================

show_help() {
    echo -e "${CYAN}CineMaker 启动管理脚本${NC}"
    echo ""
    echo "用法: $0 <命令> [选项]"
    echo ""
    echo -e "${GREEN}开发模式（源码挂载 + 热加载，改代码自动重编）:${NC}"
    echo "  dev [--build]    启动开发环境（后端 air 热加载 + 前端 Vite HMR）"
    echo "                   --build/-b  重新构建开发镜像（改了 go.mod 时用）"
    echo "  stop             停止所有服务"
    echo "  restart          重启服务"
    echo "  rebuild          重新构建开发镜像并启动"
    echo "  status           查看运行状态"
    echo "  logs [b|f|a] [-f] 查看日志（b=后端 f=前端 a=全部 -f=实时）"
    echo ""
    echo -e "${GREEN}部署模式（多阶段构建生产镜像）:${NC}"
    echo "  deploy start     启动"
    echo "  deploy stop      停止"
    echo "  deploy rebuild   重新构建并启动"
    echo ""
    echo -e "${GREEN}日常开发流程:${NC}"
    echo "  $0 dev              # 首次启动开发环境"
    echo "  # 编辑 .go 文件 → air 自动检测 → ~5秒重编 → 服务自动重启"
    echo "  # 编辑前端代码 → Vite 自动热更新 → 浏览器即时刷新"
    echo "  $0 logs b -f        # 实时看后端日志（含 air 重编信息）"
    echo "  $0 stop             # 下班收工"
    echo ""
}

# ==================== 入口 ====================

main() {
    local command=${1:-help}

    case $command in
        dev)
            do_dev_start "$2"
            ;;
        stop)
            do_dev_stop
            ;;
        restart)
            do_dev_restart "$2"
            ;;
        rebuild)
            do_dev_rebuild
            ;;
        status)
            do_dev_status
            ;;
        logs)
            do_dev_logs "$2" "$3"
            ;;
        deploy)
            local sub=${2:-start}
            case $sub in
                start)   do_deploy_start ;;
                stop)    do_deploy_stop ;;
                rebuild) do_deploy_rebuild ;;
                *)       echo -e "${RED}未知子命令: $sub${NC}"; show_help ;;
            esac
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${RED}未知命令: $command${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
