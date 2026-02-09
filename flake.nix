{
  description = "Cloudless dev environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        devShells.default = pkgs.mkShell {
          name = "cloudless-dev-shell";

          packages = with pkgs; [
            # Core
            # git
            # curl
            # wget
            # jq

            # Go ecosystem
            go
            gopls
            golangci-lint
            go-tools

            # Rust ecosystem
            # rustc
            # cargo
            # rust-analyzer

            # Web / tooling
            # nodejs_20
            # pnpm

            # Containers & DB tools
            # docker
            # docker-compose
            # postgresql
            # redis

            # Editor helpers
            # neovim
            # tree-sitter
          ];

          shellHook = ''
            echo "☁️  Entering Cloudless dev shell"
            echo "Go version: $(go version 2>/dev/null || echo not found)"
            echo "Rust version: $(rustc --version 2>/dev/null || echo not found)"
          '';
        };
      });
}
