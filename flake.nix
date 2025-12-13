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
          default = pkgs.temporal-cli.overrideAttrs {
            version = "1.3.0";
            src = lib.cleanSource ./.;
            vendorHash = "sha256-AO6djBGm4cUZ1p1h3AMskNnJSxV0OSyOkvTLrpiVw8g=";
          };
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
