#!/bin/bash
set -e

REPO="qingchencloud/cftunnel"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr A-Z a-z)
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac

FILENAME="cftunnel_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$FILENAME"
MIRRORS=("https://ghfast.top/" "https://gh-proxy.com/" "https://ghproxy.cn/" "")

echo "正在下载 cftunnel ($OS/$ARCH)..."
TMP=$(mktemp -d)
download_ok=false
for mirror in "${MIRRORS[@]}"; do
  url="${mirror}${DOWNLOAD_URL}"
  src="${mirror:-GitHub}"
  echo "  尝试: ${src} ..."
  if curl -fsSL --connect-timeout 10 -o "$TMP/$FILENAME" "$url"; then
    download_ok=true; echo "  下载成功"; break
  fi
  echo "  失败，尝试下一个源..."
done
if [ "$download_ok" = false ]; then
  echo "所有下载源均失败，请检查网络后重试"; rm -rf "$TMP"; exit 1
fi
tar xzf "$TMP/$FILENAME" -C "$TMP"
sudo install -m 755 "$TMP/cftunnel" "$INSTALL_DIR/cftunnel"
rm -rf "$TMP"
echo "cftunnel 已安装到 $INSTALL_DIR/cftunnel"
echo "运行 cftunnel init 开始配置"
