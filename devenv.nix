{ pkgs, ... }:

{
  # https://devenv.sh/basics/

  # https://devenv.sh/packages/
  packages = [ pkgs.buildkit pkgs.skopeo pkgs.gomod2nix pkgs.golangci-lint ];

  # https://devenv.sh/scripts/

  enterShell = ''
    go mod download
    gomod2nix generate
  '';

  # https://devenv.sh/tests/
  enterTest = ''
    go test ./...
  '';

  # https://devenv.sh/services/
  # services.postgres.enable = true;

  # https://devenv.sh/languages/
  languages.go.enable = true;

  # https://devenv.sh/pre-commit-hooks/
  # pre-commit.hooks.shellcheck.enable = true;

  # https://devenv.sh/processes/
  # processes.ping.exec = "ping example.com";

  # See full reference at https://devenv.sh/reference/options/
}
