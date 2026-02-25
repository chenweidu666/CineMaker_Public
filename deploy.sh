#!/bin/bash
# CineMaker 部署脚本
# 在 Mac 上构建 Docker 镜像，通过 SSH 部署到 NAS
#
# 用法:
#   ./deploy.sh                    # 完整构建 + 部署
#   ./deploy.sh --build-only       # 仅构建，不部署
#   ./deploy.sh --deploy-only      # 仅部署（使用已有镜像文件）

set -e

# ============ 配置 ============
NAS_HOST="${NAS_HOST:-cw@192.168.31.10}"
NAS_PROJECT_DIR="${NAS_PROJECT_DIR:-/home/cw/For-Work/2_项目经验/1_大模型/5_Agent/CineMaker}"
IMAGE_NAME="cinemaker-cinemaker"
IMAGE_TAG="latest"
TAR_FILE="cinemaker-image.tar.gz"
DESKTOP_PATH="$HOME/Desktop/${TAR_FILE}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${CYAN}[$(date +%H:%M:%S)]${NC} $1"; }
ok()   { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; exit 1; }

# ============ 参数解析 ============
BUILD=true
DEPLOY=true
case "${1:-}" in
  --build-only)  DEPLOY=false ;;
  --deploy-only) BUILD=false ;;
esac

# ============ 步骤1: 本地构建 ============
if [ "$BUILD" = true ]; then
  log "开始构建 Docker 镜像..."
  START=$(date +%s)

  docker build \
    --platform linux/arm64 \
    -t "${IMAGE_NAME}:${IMAGE_TAG}" \
    --build-arg GO_PROXY=https://goproxy.cn,direct \
    . 2>&1 | tail -5

  BUILD_TIME=$(( $(date +%s) - START ))
  ok "镜像构建完成 (${BUILD_TIME}s)"

  log "导出镜像到 ${DESKTOP_PATH} ..."
  docker save "${IMAGE_NAME}:${IMAGE_TAG}" | gzip > "${DESKTOP_PATH}"
  FILE_SIZE=$(du -h "${DESKTOP_PATH}" | cut -f1)
  ok "镜像导出完成 (${FILE_SIZE})"
fi

if [ "$DEPLOY" = false ]; then
  ok "仅构建模式，跳过部署"
  exit 0
fi

# ============ 步骤2: 传输到 NAS ============
if [ ! -f "${DESKTOP_PATH}" ]; then
  err "镜像文件不存在: ${DESKTOP_PATH}，请先运行构建"
fi

log "传输镜像到 NAS (${NAS_HOST}) ..."
START=$(date +%s)
scp "${DESKTOP_PATH}" "${NAS_HOST}:/tmp/${TAR_FILE}"
TRANSFER_TIME=$(( $(date +%s) - START ))
ok "传输完成 (${TRANSFER_TIME}s)"

# ============ 步骤3: 远程部署 ============
log "在 NAS 上加载镜像并重启服务..."
ssh "${NAS_HOST}" bash -s <<REMOTE_SCRIPT
  set -e
  echo ">>> 加载 Docker 镜像..."
  docker load < /tmp/${TAR_FILE}
  rm -f /tmp/${TAR_FILE}

  echo ">>> 重启容器..."
  cd "${NAS_PROJECT_DIR}"

  # 停止旧容器
  docker compose down 2>/dev/null || true

  # 用预构建镜像启动（不再在 NAS 上 build）
  docker compose up -d --no-build 2>&1

  echo ">>> 等待健康检查..."
  for i in \$(seq 1 20); do
    STATUS=\$(docker inspect --format='{{.State.Health.Status}}' cinemaker 2>/dev/null || echo "starting")
    if [ "\$STATUS" = "healthy" ]; then
      echo ">>> 容器已健康运行!"
      exit 0
    fi
    sleep 3
  done
  echo ">>> 警告: 容器未在60秒内通过健康检查"
REMOTE_SCRIPT

ok "部署完成!"
log "访问: http://192.168.31.10:7860"
