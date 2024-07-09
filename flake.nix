{
  description = "zbpack aims to automatically analyze the language, version, and framework used based on the source code and package the service into the most suitable deployment form, such as static resources, cloud functions, containers, or multiple types by one click.";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";
  inputs.gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
  inputs.gomod2nix.inputs.flake-utils.follows = "flake-utils";

  outputs = { self, nixpkgs, gomod2nix, flake-utils }:
    let
      # to work with older version of flakes
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";

      # Generate a user-friendly version number.
      version = builtins.substring 0 8 lastModifiedDate;
    in
    # Provide some binary packages for selected system types.
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          buildGoApplication = gomod2nix.legacyPackages.${system}.buildGoApplication;
          pkgs = nixpkgs.legacyPackages.${system}.extend gomod2nix.overlays.default;
        in
        {
          checks.zbpack = pkgs.runCommand "zbpack-command" {
            src = ./.;
            buildInputs = [ self.packages.${system}.default ];
          } ''
            zbpack -h
            mkdir "$out"
          '';

          packages.default = buildGoApplication {
            pname = "zbpack";
            inherit version;
            pwd = ./cmd/zbpack;
            src = ./.;
            modules = ./gomod2nix.toml;
          };
        }
      );
}
