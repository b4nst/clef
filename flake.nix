{
  description = "Clef flake";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";

    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.utils.follows = "utils";
    };
  };
  
  
  outputs = { self, nixpkgs, utils, gomod2nix }:
    utils.lib.eachDefaultSystem
      (system:
        let
          overlays = [ gomod2nix.overlays.default ];
          pkgs = import nixpkgs {
            inherit system overlays;
          };
        in
        with pkgs;
        {
          devShells.default = mkShell {
            name = "b4nst/clef";
            buildInputs = [
              go
              golangci-lint
              gopls
              goreleaser
              gotools
            ];
         };
        }
      );
}
