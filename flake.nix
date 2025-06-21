{
  description = "Temporal Server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
  };

  outputs =
    inputs:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forEachSupportedSystem =
        f:
        inputs.nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            pkgs = import inputs.nixpkgs {
              inherit system;
              overlays = [
                (import ./nix/go-override.nix) # https://github.com/NixOS/nixpkgs/pull/414434/files
              ];
            };
          }
        );
    in
    {
      packages = forEachSupportedSystem (
        { pkgs }:
        let
          inherit (pkgs) lib;
        in
        {
          default = pkgs.buildGoModule (finalAttrs: {
            pname = "temporalcli";
            version = "1.3.0";

            src = builtins.path {
              path = ./.;
              filter = lib.cleanSourceFilter;
              name = "temporalcli-source";
            };

            vendorHash = "sha256-nHLN8VTD4Zlc8kjjv4XLxgDe+/wN339nukl/VbhWchU=";

            doCheck = false;

            meta = {
              description = "Command-line interface for running Temporal Server and interacting with Workflows, Activities, Namespaces, and other parts of Temporal";
              homepage = "https://github.com/temporalio/cli";
              license = lib.licenses.mit;
              mainProgram = "temporal";
            };
          });
        }
      );

      devShells = forEachSupportedSystem (
        { pkgs }:
        {
          default = pkgs.mkShellNoCC {
            packages = with pkgs; [
              go
              gotools
              golangci-lint
            ];
          };
        }
      );
    };
}
