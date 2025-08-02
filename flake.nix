{
  description = "Go development environment with Nix Flakes";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      
      in {
        devShell = pkgs.mkShell {
          buildInputs = [
            pkgs.go_1_24
            pkgs.gopls
            # pkgs.sqlite
            # pkgs.nodejs
            # pkgs.graphql-codegen
            pkgs.gotools
          ];

          shellHook = ''
            go mod tidy
            echo "Go development environment is ready!"
          '';
        };
      });
}
