#!/usr/bin/env bash
set -euo pipefail

# cftunnel 中继服务端一键安装脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/qingchencloud/cftunnel/main/install-relay.sh | bash

FRP_VERSION="0.66.0"
BIND_PORT="${RELAY_PORT:-7000}"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/frps"
SERVICE_NAME="frps"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

# 检查 root
if [ "$(id -u)" -ne 0 ]; then
    error "请使用 root 用户运行，或 sudo bash"
fi

# 检查系统
if [ "$(uname -s)" != "Linux" ]; then
    error "此脚本仅支持 Linux"
fi

# 检测架构
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  FRP_ARCH="amd64" ;;
    aarch64) FRP_ARCH="arm64" ;;
    armv7l)  FRP_ARCH="arm" ;;
    *)       error "不支持的架构: $ARCH" ;;
esac

FILENAME="frp_${FRP_VERSION}_linux_${FRP_ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/fatedier/frp/releases/download/v${FRP_VERSION}/${FILENAME}"

# 镜像源列表
MIRRORS=(
    "https://ghfast.top/"
    "https://gh-proxy.com/"
    "https://ghproxy.cn/"
    ""
)

# 下载 frps
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

download_ok=false
for mirror in "${MIRRORS[@]}"; do
    url="${mirror}${DOWNLOAD_URL}"
    src="${mirror:-GitHub}"
    info "尝试下载: ${src} ..."
    if curl -fsSL --connect-timeout 10 -o "$TMP_DIR/$FILENAME" "$url"; then
        download_ok=true
        info "下载成功"
        break
    fi
    warn "下载失败，尝试下一个源..."
done

if [ "$download_ok" = false ]; then
    error "所有下载源均失败"
fi

# 解压
info "正在解压..."
tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"
EXTRACT_DIR="$TMP_DIR/frp_${FRP_VERSION}_linux_${FRP_ARCH}"

# 安装 frps
install -m 755 "$EXTRACT_DIR/frps" "$INSTALL_DIR/frps"
info "frps 已安装到 $INSTALL_DIR/frps"

# 生成随机 token
TOKEN=$(head -c 16 /dev/urandom | xxd -p)

# 写配置文件
mkdir -p "$CONFIG_DIR"
cat > "$CONFIG_DIR/frps.toml" <<EOF
bindPort = ${BIND_PORT}
auth.token = "${TOKEN}"
EOF
chmod 600 "$CONFIG_DIR/frps.toml"
info "配置文件: $CONFIG_DIR/frps.toml"

# 注册 systemd 服务
cat > /etc/systemd/system/${SERVICE_NAME}.service <<EOF
[Unit]
Description=frps relay server (cftunnel)
After=network.target

[Service]
ExecStart=${INSTALL_DIR}/frps -c ${CONFIG_DIR}/frps.toml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now "$SERVICE_NAME"
info "systemd 服务已启动"

# 获取服务器 IP
SERVER_IP=$(curl -s --connect-timeout 5 https://api.ipify.org 2>/dev/null || hostname -I | awk '{print $1}')

echo ""
echo "╔════════════════════════════════════════════╗"
echo "║       ✅ frps 中继服务端安装完成!           ║"
echo "╚════════════════════════════════════════════╝"
echo ""
echo "  二进制: $INSTALL_DIR/frps"
echo "  配置:   $CONFIG_DIR/frps.toml"
echo "  端口:   $BIND_PORT"
echo ""
echo "在客户端执行以下命令连接:"
echo ""
echo "  cftunnel relay init --server ${SERVER_IP}:${BIND_PORT} --token ${TOKEN}"
echo ""
echo "也可以使用 Docker 部署: https://github.com/qingchencloud/cftunnel#relay-server"
echo ""
