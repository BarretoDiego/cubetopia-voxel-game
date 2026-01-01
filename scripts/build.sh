#!/bin/bash

# ============================================================
# Voxel Engine - Cross-Platform Build Script
# ============================================================
# Este script compila o jogo para todas as plataformas suportadas
# com nomes corretos e metadados embutidos no binário.
#
# Plataformas suportadas:
#   - Windows (amd64)
#   - macOS (amd64, arm64)
#   - Linux (amd64)
#
# Uso:
#   ./scripts/build.sh [versão] [--all|--windows|--macos|--linux]
#
# Exemplos:
#   ./scripts/build.sh                    # Build para plataforma atual
#   ./scripts/build.sh 1.0.0 --all        # Build para todas as plataformas
#   ./scripts/build.sh 1.0.0 --windows    # Build apenas para Windows
# ============================================================

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Diretórios
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"
GO_DIR="$PROJECT_ROOT/go"
BUILD_DIR="$PROJECT_ROOT/build"
ASSETS_DIR="$GO_DIR/assets"

# Configurações do jogo
GAME_NAME="VoxelEngine"
GAME_DISPLAY_NAME="Voxel Engine"
GAME_DESCRIPTION="A high-performance voxel engine written in Go with OpenGL rendering"
GAME_AUTHOR="Diego da Cunha"
GAME_BUNDLE_ID="com.diegodacunha.voxelengine"

# Versão padrão
VERSION="${1:-1.0.0}"
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git -C "$PROJECT_ROOT" rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Flags de build
BUILD_FLAGS="-trimpath"
LD_FLAGS="-s -w \
    -X 'main.Version=$VERSION' \
    -X 'main.BuildDate=$BUILD_DATE' \
    -X 'main.GitCommit=$GIT_COMMIT' \
    -X 'main.GameName=$GAME_DISPLAY_NAME'"

# ============================================================
# Funções auxiliares
# ============================================================

print_header() {
    echo -e "${CYAN}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║              VOXEL ENGINE - BUILD SYSTEM                 ║"
    echo "╠═══════════════════════════════════════════════════════════╣"
    echo "║  Version: $VERSION                                            ║"
    echo "║  Commit:  $GIT_COMMIT                                       ║"
    echo "║  Date:    $BUILD_DATE                       ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Verifica dependências necessárias
check_dependencies() {
    log_info "Verificando dependências..."

    # Verifica Go
    if ! command -v go &> /dev/null; then
        log_error "Go não encontrado! Instale Go 1.21+ primeiro."
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $GO_VERSION"

    # Verifica Git (opcional, para commit hash)
    if ! command -v git &> /dev/null; then
        log_warning "Git não encontrado. Commit hash será 'unknown'."
    fi

    log_success "Dependências verificadas!"
}

# Prepara diretórios de build
prepare_build_dirs() {
    log_info "Preparando diretórios de build..."

    # Limpa builds anteriores
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"

    log_success "Diretórios preparados!"
}

# Build para uma plataforma específica
build_platform() {
    local os="$1"
    local arch="$2"
    local name_suffix="$3"
    
    local output_name="${GAME_NAME}-${VERSION}-${os}-${arch}"
    local output_dir="$BUILD_DIR/$output_name"
    local binary_name="${GAME_NAME}${name_suffix}"
    
    echo ""
    log_info "═══════════════════════════════════════════════════"
    log_info "Building para $os/$arch..."
    log_info "═══════════════════════════════════════════════════"
    
    mkdir -p "$output_dir"
    
    # Configura ambiente de cross-compilation
    export CGO_ENABLED=1
    export GOOS="$os"
    export GOARCH="$arch"
    
    # Configura CC para cross-compilation
    case "$os-$arch" in
        "windows-amd64")
            if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
                export CC=x86_64-w64-mingw32-gcc
            else
                log_warning "mingw-w64 não encontrado. Build para Windows pode falhar."
                log_warning "Instale com: brew install mingw-w64"
            fi
            ;;
        "linux-amd64")
            if command -v x86_64-linux-gnu-gcc &> /dev/null; then
                export CC=x86_64-linux-gnu-gcc
            elif [[ "$(uname)" == "Linux" ]]; then
                export CC=gcc
            else
                log_warning "Cross-compiler para Linux não encontrado."
                log_warning "Instale com: brew install FiloSottile/musl-cross/musl-cross"
            fi
            ;;
        "darwin-amd64"|"darwin-arm64")
            if [[ "$(uname)" == "Darwin" ]]; then
                export CC=clang
            else
                log_warning "Compilação para macOS requer macOS."
            fi
            ;;
    esac
    
    cd "$GO_DIR"
    
    # Baixa dependências
    log_info "Baixando dependências..."
    go mod tidy
    go mod download
    
    # Compila
    log_info "Compilando $binary_name (assets embutidos)..."
    
    if go build $BUILD_FLAGS -ldflags "$LD_FLAGS" -o "$output_dir/$binary_name" ./cmd/voxelgame 2>&1; then
        log_success "Binário compilado: $output_dir/$binary_name"
        log_info "✓ Assets (shaders + texturas) embutidos no executável!"
        
        # Cria arquivo de informações
        create_info_file "$output_dir"
        
        # Cria arquivo README
        create_readme "$output_dir" "$os" "$arch"
        
        # Cria pacote ZIP (exceto para a plataforma atual durante dev)
        if [[ "$2" != "--current" ]]; then
            create_zip "$output_dir" "$output_name"
        fi
        
        echo ""
        log_success "✓ Build para $os/$arch concluído!"
        echo "  Output: $output_dir"
        echo "  Executável standalone: $binary_name"
        
        return 0
    else
        log_error "Falha ao compilar para $os/$arch"
        return 1
    fi
}

# Cria arquivo de informações
create_info_file() {
    local dir="$1"
    
    cat > "$dir/version.txt" << EOF
$GAME_DISPLAY_NAME
================================
Version:     $VERSION
Build Date:  $BUILD_DATE
Git Commit:  $GIT_COMMIT
Description: $GAME_DESCRIPTION
Author:      $GAME_AUTHOR
EOF
}

# Cria arquivo README
create_readme() {
    local dir="$1"
    local os="$2"
    local arch="$3"
    
    local run_instructions=""
    case "$os" in
        "windows")
            run_instructions="Duplo-clique em VoxelEngine.exe ou execute via terminal."
            ;;
        "darwin")
            run_instructions="Abra o Terminal e execute: ./VoxelEngine"
            ;;
        "linux")
            run_instructions="Abra o Terminal e execute: ./VoxelEngine"
            ;;
    esac
    
    cat > "$dir/README.txt" << EOF
╔═══════════════════════════════════════════════════════════════════╗
║                      $GAME_DISPLAY_NAME                              ║
╚═══════════════════════════════════════════════════════════════════╝

Versão: $VERSION
Plataforma: $os ($arch)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

COMO EXECUTAR
━━━━━━━━━━━━━
$run_instructions

Este é um executável STANDALONE - todos os assets (shaders, texturas)
estão embutidos no binário. Você pode mover o executável para qualquer
lugar do seu sistema.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

CONTROLES
━━━━━━━━━
  WASD       - Mover
  Mouse      - Olhar ao redor
  Shift      - Correr
  Space      - Pular
  Ctrl       - Agachar
  F          - Modo voar
  C          - Alternar câmera (1ª/3ª pessoa)
  R          - Raytracing
  1-9        - Selecionar slot da hotbar
  Scroll     - Navegar hotbar
  LMB        - Quebrar bloco
  RMB        - Colocar bloco
  F3         - Debug
  F5         - Salvar rápido
  F9         - Carregar rápido
  ESC/P      - Pausar/Menu

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

REQUISITOS
━━━━━━━━━━
  • OpenGL 4.1 ou superior
  • GPU com suporte a shaders modernos
  • 4GB RAM recomendado

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Desenvolvido por $GAME_AUTHOR
Licença: MIT

EOF
}

# Cria arquivo ZIP
create_zip() {
    local dir="$1"
    local name="$2"
    
    log_info "Criando arquivo ZIP..."
    
    cd "$BUILD_DIR"
    zip -r "${name}.zip" "$(basename "$dir")" -x "*.DS_Store" -x "*/.DS_Store"
    
    log_success "ZIP criado: $BUILD_DIR/${name}.zip"
}

# ============================================================
# Funções de build por plataforma
# ============================================================

build_current() {
    local os=$(go env GOOS)
    local arch=$(go env GOARCH)
    local suffix=""
    
    if [[ "$os" == "windows" ]]; then
        suffix=".exe"
    fi
    
    build_platform "$os" "$arch" "$suffix"
}

build_windows() {
    build_platform "windows" "amd64" ".exe"
}

build_macos_amd64() {
    build_platform "darwin" "amd64" ""
}

build_macos_arm64() {
    build_platform "darwin" "arm64" ""
}

build_linux() {
    build_platform "linux" "amd64" ""
}

build_all() {
    local failed=0
    
    log_info "Iniciando build para TODAS as plataformas..."
    
    # macOS (nativo - sempre funciona no macOS)
    if [[ "$(uname)" == "Darwin" ]]; then
        build_macos_amd64 || ((failed++))
        build_macos_arm64 || ((failed++))
    fi
    
    # Windows (requer mingw-w64)
    if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        build_windows || ((failed++))
    else
        log_warning "Pulando Windows (mingw-w64 não instalado)"
        log_warning "Instale com: brew install mingw-w64"
    fi
    
    # Linux (requer cross-compiler ou estar no Linux)
    if [[ "$(uname)" == "Linux" ]] || command -v x86_64-linux-gnu-gcc &> /dev/null; then
        build_linux || ((failed++))
    else
        log_warning "Pulando Linux (cross-compiler não instalado)"
    fi
    
    return $failed
}

# ============================================================
# Sumário final
# ============================================================

print_summary() {
    echo ""
    echo -e "${GREEN}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║                    BUILD COMPLETO!                       ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    log_info "Arquivos gerados em: $BUILD_DIR"
    echo ""
    
    if [[ -d "$BUILD_DIR" ]]; then
        ls -la "$BUILD_DIR"
    fi
}

# ============================================================
# Main
# ============================================================

main() {
    print_header
    check_dependencies
    prepare_build_dirs
    
    # Parse argumentos
    local build_target="${2:---current}"
    
    case "$build_target" in
        "--all"|"-a")
            build_all
            ;;
        "--windows"|"-w")
            build_windows
            ;;
        "--macos"|"-m")
            build_macos_amd64
            build_macos_arm64
            ;;
        "--macos-amd64")
            build_macos_amd64
            ;;
        "--macos-arm64")
            build_macos_arm64
            ;;
        "--linux"|"-l")
            build_linux
            ;;
        "--current"|"-c"|*)
            build_current
            ;;
    esac
    
    print_summary
}

# Executa
main "$@"
