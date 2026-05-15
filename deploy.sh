#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_error() {
    echo -e "${RED}ERROR: $1${NC}"
}

print_success() {
    echo -e "${GREEN}SUCCESS: $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}WARNING: $1${NC}"
}

print_info() {
    echo -e "${NC}$1${NC}"
}

show_usage() {
    cat << EOF
Arc Multi-Environment Deployment Script

Usage: $0 [ENVIRONMENT] [COMMAND]

Environments:
  dev       - Development environment
  staging   - Staging environment
  prod      - Production environment
  all       - All environments simultaneously

Commands:
  up        - Start the environment
  down      - Stop the environment
  restart   - Restart the environment
  logs      - View logs (follow mode)
  build     - Build images
  status    - Show container status
  shell     - Open shell in a service container
  migrate   - Copy data from one environment to another

Examples:
  $0 dev up           # Start development environment
  $0 staging logs     # View staging logs
  $0 prod restart     # Restart production environment
  $0 all up           # Start ALL environments simultaneously
  $0 all status       # Check status of all environments
  $0 dev shell auth   # Open shell in auth service (dev)
  $0 dev migrate prod # Copy dev data to production

Port Mappings (when running all):
  Development: 8081, 2525, 2151, 6070
  Staging:     8082, 2526, 2152, 6071
  Production:  8083, 2527, 2153, 6072

EOF
}

check_dependencies() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed"
        exit 1
    fi
}

validate_environment() {
    local env=$1
    if [[ ! "$env" =~ ^(dev|staging|prod|all)$ ]]; then
        print_error "Invalid environment: $env"
        show_usage
        exit 1
    fi
}

validate_command() {
    local cmd=$1
    if [[ ! "$cmd" =~ ^(up|down|restart|logs|build|status|shell|migrate)$ ]]; then
        print_error "Invalid command: $cmd"
        show_usage
        exit 1
    fi
}

get_data_dir() {
    local env=$1
    case $env in
        dev)     echo "./data/dev" ;;
        staging) echo "./data/staging" ;;
        prod)    echo "./data/production" ;;
        *)       echo "./data/$env" ;;
    esac
}

migrate_data() {
    local source_env=$1
    local target_env=$2

    if [[ -z "$target_env" ]]; then
        print_error "Target environment required for migrate"
        print_info "Usage: $0 <source_env> migrate <target_env>"
        print_info "Example: $0 dev migrate prod"
        exit 1
    fi

    if [[ "$source_env" == "$target_env" ]]; then
        print_error "Source and target environments cannot be the same"
        exit 1
    fi

    if [[ "$source_env" == "all" || "$target_env" == "all" ]]; then
        print_error "Cannot migrate to/from 'all'"
        exit 1
    fi

    local source_dir
    source_dir=$(get_data_dir "$source_env")
    local target_dir
    target_dir=$(get_data_dir "$target_env")

    if [[ ! -d "$source_dir" ]]; then
        print_error "Source data directory does not exist: $source_dir"
        exit 1
    fi

    local source_size
    source_size=$(du -sh "$source_dir" 2>/dev/null | cut -f1)

    print_warning "Data Migration: $source_env → $target_env"
    print_info "  Source: $source_dir ($source_size)"
    print_info "  Target: $target_dir"
    print_info ""

    if [[ -d "$target_dir" ]]; then
        local target_size
        target_size=$(du -sh "$target_dir" 2>/dev/null | cut -f1)
        print_warning "Target directory already exists ($target_size)"
        print_warning "This will OVERWRITE all data in $target_dir"
    fi

    print_info ""
    read -p "Continue with migration? (yes/no): " confirm
    if [[ "$confirm" != "yes" ]]; then
        print_info "Migration cancelled"
        exit 0
    fi

    local backup_dir="${target_dir}.backup.$(date +%Y%m%d_%H%M%S)"
    if [[ -d "$target_dir" ]]; then
        print_info "Backing up existing target data to $backup_dir ..."
        cp -r "$target_dir" "$backup_dir"
        print_success "Backup created"
    fi

    print_info "Copying data from $source_dir to $target_dir ..."
    mkdir -p "$target_dir"

    if command -v rsync &> /dev/null; then
        rsync -a --delete "${source_dir}/" "${target_dir}/"
    else
        rm -rf "${target_dir:?}"/*
        cp -r "${source_dir}/"* "$target_dir/"
    fi

    print_success "Migration complete: $source_env → $target_env"
    print_info "Data copied to $target_dir"

    if [[ -d "$backup_dir" ]]; then
        print_info "Previous data backed up at $backup_dir"
    fi
}

get_files() {
    local env=$1

    if [[ "$env" == "all" ]]; then
        COMPOSE_FILE="docker-compose.all.yml"
        ENV_FILE=".env.dev"
    else
        COMPOSE_FILE="docker-compose.${env}.yml"
        ENV_FILE=".env.${env}"
    fi

    if [[ ! -f "$COMPOSE_FILE" ]]; then
        print_error "Compose file not found: $COMPOSE_FILE"
        exit 1
    fi

    if [[ ! -f "$ENV_FILE" ]] && [[ "$env" != "all" ]]; then
        print_warning "Environment file not found: $ENV_FILE"
        print_info "Using default environment variables"
    fi
}

execute_command() {
    local env=$1
    local cmd=$2
    local service=$3

    get_files "$env"

    case $cmd in
        up)
            if [[ "$env" == "all" ]]; then
                print_warning "⚠️  You are about to start ALL environments simultaneously!"
                print_warning "This will start dev, staging, AND production!"
                print_warning "This uses significant system resources."
                print_info ""
                print_info "Port mappings:"
                print_info "  Dev:     8081, 2525, 2151, 6070"
                print_info "  Staging: 8082, 2526, 2152, 6071"
                print_info "  Prod:    8083, 2527, 2153, 6072"
                print_info ""
                read -p "Continue? (yes/no): " confirm
                if [[ "$confirm" != "yes" ]]; then
                    print_info "Deployment cancelled"
                    exit 0
                fi
            elif [[ "$env" == "prod" ]]; then
                print_warning "You are about to start the PRODUCTION environment!"
                print_warning "Please ensure:"
                print_warning "  1. You have updated .env.prod with secure credentials"
                print_warning "  2. Backups are configured"
                print_warning "  3. Monitoring is in place"
                read -p "Continue? (yes/no): " confirm
                if [[ "$confirm" != "yes" ]]; then
                    print_info "Deployment cancelled"
                    exit 0
                fi
            fi

            print_info "Starting $env environment..."
            docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d
            print_success "$env environment started!"
            print_info "Run '$0 $env status' to check container status"
            ;;

        down)
            print_info "Stopping $env environment..."
            docker-compose -f "$COMPOSE_FILE" down
            print_success "$env environment stopped!"
            ;;

        restart)
            print_info "Restarting $env environment..."
            docker-compose -f "$COMPOSE_FILE" down
            docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d
            print_success "$env environment restarted!"
            ;;

        logs)
            print_info "Showing logs for $env environment (Ctrl+C to exit)..."
            docker-compose -f "$COMPOSE_FILE" logs -f
            ;;

        build)
            print_info "Building images for $env environment..."
            docker-compose -f "$COMPOSE_FILE" build
            print_success "Build complete!"
            ;;

        status)
            print_info "Status for $env environment:"
            docker-compose -f "$COMPOSE_FILE" ps
            ;;

        shell)
            if [[ "$env" == "all" ]]; then
                print_error "Cannot use shell command with 'all' environment"
                print_info "Please specify a specific environment: dev, staging, or prod"
                print_info "Example: $0 dev shell auth"
                exit 1
            fi

            if [[ -z "$service" ]]; then
                print_error "Service name required for shell command"
                print_info "Available services: auth, account, anticheat, internal"
                exit 1
            fi

            container_name="arc-${service}-${env}"
            print_info "Opening shell in $container_name..."
            docker exec -it "$container_name" /bin/sh
            ;;

        migrate)
            migrate_data "$env" "$service"
            ;;

        *)
            print_error "Unknown command: $cmd"
            exit 1
            ;;
    esac
}

main() {
    check_dependencies

    if [[ $# -lt 2 ]]; then
        show_usage
        exit 1
    fi

    local env=$1
    local cmd=$2
    local service=$3

    validate_environment "$env"
    validate_command "$cmd"

    execute_command "$env" "$cmd" "$service"
}

main "$@"