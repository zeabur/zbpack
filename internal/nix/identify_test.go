package nix_test

import (
	"testing"

	"github.com/zeabur/zbpack/internal/nix"
)

func TestFindPossibleNixDockerPackage(t *testing.T) {
	t.Parallel()

	testmap := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "realcase-bloop",
			content:  bloopNix,
			expected: "docker",
		},
		{
			name:     "realcase-wfmash",
			content:  wfmashNix,
			expected: "dockerImage",
		},
		{
			name:     "realcase-bonfire-app",
			content:  bonfireAppNix,
			expected: "container",
		},
		{
			name:     "realcase-blink",
			content:  blinkNix,
			expected: "dockerImage",
		},
		{
			name:     "realcase-huffman",
			content:  huffmanNix,
			expected: "image",
		},
		{
			name:     "realcase-quartz",
			content:  quartzNix,
			expected: "Docker",
		},
		{
			name:     "realcase-task-scheduler",
			content:  taskSchedulerNix,
			expected: "docker",
		},
		{
			name:     "snipeet-docker",
			content:  `docker = pkgs.dockerTools.buildImage {`,
			expected: "docker",
		},
		{
			name:     "snipeet-docker-image",
			content:  `dockerImage = pkgs.dockerTools.buildImage {`,
			expected: "dockerImage",
		},
		{
			name:     "snipeet-docker-image-full",
			content:  `packages.${system}.docker-image = pkgs.dockerTools.buildImage {`,
			expected: "docker-image",
		},
		{
			name:     "snipeet-docker-image-full-commented",
			content:  `#  packages.${system}.dockerImage = pkgs.dockerTools.buildImage {`,
			expected: "",
		},
		{
			name:     "snipeet-docker-image-full-commented-nfc",
			content:  `   # packages.${system}.dockerImage = pkgs.dockerTools.buildImage {`,
			expected: "",
		},
		{
			name:     "snipeet-docker-image-incomplete",
			content:  `packages.${system}.dockerImage`,
			expected: "",
		},
	}

	for _, test := range testmap {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := nix.FindPossibleNixDockerPackage(test.content); got != test.expected {
				t.Errorf("expected %q, got %q", test.expected, got)
			}
		})
	}
}

// Expected: "docker"
// https://github.com/BloopAI/bloop/blob/1c110d173d2f1c9750d84db2e0864695fa910070/flake.nix#L176
const bloopNix = `{
  description = "bloop";

  # nixConfig = {
  #   extra-substituters = "https://bloopai.cachix.org";
  #   extra-trusted-public-keys =
  #     "bloopai.cachix.org-1:uSHFor+Jd3znikUnLc58xnHBXTcuIBSjdJxV5rLIMJU=";
  # };

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    nixpkgs2305.url = "github:nixos/nixpkgs/nixos-23.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, nixpkgs2305, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        pkgsStable = import nixpkgs2305 { inherit system; };
        pkgsStatic = pkgs.pkgsStatic;
        lib = pkgs.lib;

        llvm = pkgs.llvmPackages_14;
        clang = llvm.clang;
        libclang = llvm.libclang;
        stdenv =
          if pkgs.stdenv.isLinux then
            pkgs.stdenvAdapters.useMoldLinker llvm.stdenv
          else
            llvm.stdenv;

        mkShell =
          if stdenv.isLinux then
            pkgs.mkShell.override { inherit stdenv; }
          else
            pkgs.mkShell;

        rustPlatform = pkgs.makeRustPlatform {
          cargo = pkgs.cargo;
          rustc = pkgs.rustc;
        };

        runtimeDeps = with pkgs;
          ([ openssl.out rocksdb git zlib nsync ]
            ++ lib.optionals stdenv.isDarwin [
            darwin.apple_sdk.frameworks.Foundation
            darwin.apple_sdk.frameworks.CoreFoundation
            darwin.apple_sdk.frameworks.Security
          ]);

        buildDeps = with pkgs;
          ([
            stdenv.cc.cc.lib
            glib.dev
            pkg-config
            openssl.out
            openssl.dev
            llvm.bintools

            protobuf
          ] ++ lib.optionals stdenv.isDarwin [
            darwin.apple_sdk.frameworks.Foundation
            darwin.apple_sdk.frameworks.CoreFoundation
            darwin.apple_sdk.frameworks.Security
          ]);

        guiDeps = with pkgs;
          [ nodePackages.npm nodejs ] ++ (lib.optionals stdenv.isLinux [
            gdk-pixbuf
            gdk-pixbuf.dev
            zlib.dev
            dbus.dev
            libsoup.dev
            gtk3.dev
            webkitgtk
            dmidecode
            appimage-run
            appimagekit
            gdk-pixbuf
          ] ++ lib.optionals stdenv.isDarwin [
            darwin.apple_sdk.frameworks.MetalKit
            darwin.apple_sdk.frameworks.MetalPerformanceShaders
            darwin.apple_sdk.frameworks.Carbon
            darwin.apple_sdk.frameworks.WebKit
            darwin.apple_sdk.frameworks.AppKit
          ]);

        envVars = {
          LIBCLANG_PATH = "${libclang.lib}/lib";
          ROCKSDB_LIB_DIR = "${pkgs.rocksdb}/lib";
          ROCKSDB_INCLUDE_DIR = "${pkgs.rocksdb}/include";
          OPENSSL_LIB_DIR = "${pkgs.openssl.out}/lib";
          OPENSSL_INCLUDE_DIR = "${pkgs.openssl.dev}/include";
          OPENSSL_NO_VENDOR = "1";
        } // lib.optionalAttrs stdenv.isLinux {
          RUSTFLAGS = "-C link-arg=-fuse-ld=mold";
        };

        bleep =
          (rustPlatform.buildRustPackage.override { inherit stdenv; } rec {
            meta = with pkgs.lib; {
              description = "Search code. Fast.";
              homepage = "https://bloop.ai";
              license = licenses.asl20;
              platforms = platforms.all;
            };

            name = "bleep";
            pname = name;
            src = pkgs.lib.sources.cleanSource ./.;

            cargoLock = {
              lockFile = ./Cargo.lock;
              outputHashes = {
                "hyperpolyglot-0.1.7" =
                  "sha256-JY75NB6sPxN0p/hksnBbat4S2EYFi2nExYoVHpYoib8=";
                "tree-sitter-cpp-0.20.0" =
                  "sha256-h6mJdmQzJlxYIcY+d5IiaFghraUgBGZwqFPKwB3E4pQ=";
                "tree-sitter-go-0.19.1" =
                  "sha256-f885YTswEDH/QfRPUxcLp/1E2zXLKl25R9IyTGKb1eM=";
                "tree-sitter-java-0.20.0" =
                  "sha256-gQzoWGV9wYiLibMFkLoY2sdEJg+ae9NnHt/GFfFzP8U=";
                "ort-1.14.8" =
                  "sha256-6YAhbrgI95WwRV0ngS0yaYlxfDGUFXYU0/oGf6vs68M=";
                "comrak-0.18.0" =
                  "sha256-UWY00jF2aKAG3Oz0P1UWF/7TiTIrCUGHwfjW+O1ok7Q=";
                "tree-sitter-php-0.19.1" =
                  "sha256-oHUfcuqtFFl+70/uJjE74J1JVV93G9++UaEIntOH5tM=";
                "esaxx-rs-0.1.8" =
                  "sha256-rPNNSn829eOo/glgmHPqnoylZmDLlaI5vKMRtfTikGs=";
              };
            };

            buildNoDefaultFeatures = true;
            checkNoDefaultFeatures = true;
            cargoTestFlags = "-p ${name}";
            cargoBuildFlags = "-p ${name}";

            nativeCheckInputs = buildDeps;
            nativeBuildInputs = buildDeps;
            checkInputs = runtimeDeps;
            buildInputs = runtimeDeps;
          }).overrideAttrs (old: envVars);

        frontend = (pkgs.buildNpmPackage rec {
          meta = with pkgs.lib; {
            description = "Search code. Fast.";
            homepage = "https://bloop.ai";
            license = licenses.asl20;
            platforms = platforms.all;
          };

          name = "bleep-frontend";
          pname = name;
          src = pkgs.lib.sources.cleanSource ./.;

          # The prepack script runs the build script, which we'd rather do in the build phase.
          npmPackFlags = [ "--ignore-scripts" ];
          npmDepsHash = "sha256-YvmdThbqlmQ9MXL+a7eyXJ33sQNStQah9MUW2zhc/Uc=";
          makeCacheWritable = true;
          npmBuildScript = "build-web";
          installPhase = ''
            mkdir -p $out
            cp -r client/dist $out/dist
          '';
        });

      in
      {
        packages = {
          default = bleep;

          frontend = frontend;
          bleep = bleep;
          docker = pkgs.dockerTools.buildImage {
            name = "bleep";
            config = { Cmd = [ "${bleep}/bin/bleep" ]; };
            extraCommands = ''
              ln -s ${bleep}/bin/bleep /bleep
            '';

          };
        };

        devShells = {
          default = (mkShell {
            buildInputs = buildDeps ++ runtimeDeps ++ guiDeps ++ (with pkgs; [
              git-lfs
              rustfmt
              clippy
              rust-analyzer
              cargo
              rustc
              cargo-watch
            ]);

            src = pkgs.lib.sources.cleanSource ./.;

            BLOOP_LOG = "bleep=debug";
          }).overrideAttrs (old: envVars);
        };

        formatter = pkgs.nixfmt;
      });
}`

// Expected: "dockerImage"
// https://github.com/waveygang/wfmash/blob/13882fe1788d0861ebcff3ff356bce266484c9c3/flake.nix#L61
const wfmashNix = `{
  description = "A flake for building wfmash and a Docker image for it";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
  in {
    packages.${system}.wfmash = pkgs.stdenv.mkDerivation rec {
      pname = "wfmash";
      version = "0.14.0";

      src = pkgs.fetchFromGitHub {
        owner = "waveygang";
        repo = "wfmash";
        rev = "7376468d0d1f67ad58ca3fc5d07e888745b22c06";
        sha256 = "sha256-0vThKJflmasPoVhGdFwm1l9VUoxZkEwZruYhlzF1Ehw=";
      };

      nativeBuildInputs = [ pkgs.cmake pkgs.makeWrapper ];

      buildInputs = [
        pkgs.gsl
        pkgs.gmp
        pkgs.jemalloc
        pkgs.htslib
        pkgs.git
        pkgs.zlib
        pkgs.pkg-config
      ];

      # Define custom attributes
      enableOptimizations = true;
      reproducibleBuild = false;

      # Use custom attributes to set compiler flags
      CFLAGS = if enableOptimizations then "-Ofast -march=x86-64-v3" else "";
      CXXFLAGS = if enableOptimizations then "-Ofast -march=x86-64-v3" else "";

      postPatch = ''
        mkdir -p include
        echo "#define WFMASH_GIT_VERSION \"${version}\"" > include/wfmash_git_version.hpp
      '';

      postInstall = ''
        wrapProgram $out/bin/wfmash --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.gsl pkgs.gmp ]}
      '';

      meta = with pkgs.lib; {
        description = "Base-accurate DNA sequence alignments using WFA and mashmap2";
        homepage = "https://github.com/ekg/wfmash";
        license = licenses.mit;
        platforms = platforms.linux;
        maintainers = [ maintainers.bzizou ];
      };
    };

    dockerImage = pkgs.dockerTools.buildImage {
      name = "wfmash-docker";
      tag = "latest";
      copyToRoot = [ self.packages.${system}.wfmash ];
      config = {
        Entrypoint = [ "${self.packages.${system}.wfmash}/bin/wfmash" ];
      };
    };
  };
}`

// expected: "container"
// https://github.com/bonfire-networks/bonfire-app/blob/ccf5b0aa51e117411cf512dadbd555081d4e135f/flake.nix#L83
const bonfireAppNix = `{
  description = "Bonfire self contained build";

  inputs = {
    nixpkgs = { url = "github:NixOS/nixpkgs/nixpkgs-unstable"; };
    flake-utils = { url = "github:numtide/flake-utils"; };
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    let
      # props to hold settings to apply on this file like name and version
      props = import ./props.nix;
      # set elixir nix version
      elixir_nix_version = elixir_version:
        builtins.replaceStrings [ "." ] [ "_" ] "elixir_${elixir_version}";
      erlang_nix_version = erlang_version: "erlangR${erlang_version}";
    in
    flake-utils.lib.eachSystem flake-utils.lib.defaultSystems (system:
      let
        inherit (nixpkgs.lib) optional;
        pkgs = import nixpkgs { inherit system; };

        # project name for mix release
        pname = props.app_name;
        # project version for mix release
        version = props.app_version;

        # use ~r/erlangR[1-9]+/ for specific erlang release version
        beamPackages = pkgs.beam.packagesWith
          pkgs.beam.interpreters.${erlang_nix_version props.erlang_release};
        # all elixir and erlange packages
        erlang = beamPackages.erlang;
        # use ~r/elixir_1_[1-9]+/ major elixir version
        elixir = beamPackages.${elixir_nix_version props.elixir_release};
        elixir-ls = beamPackages.elixir_ls.overrideAttrs
          (oldAttrs: rec { elixir = elixir; });
        hex = beamPackages.hex;

        # use rebar from nix instead of fetch externally
        rebar3 = beamPackages.rebar3;
        rebar = beamPackages.rebar;

        locality = pkgs.glibcLocales;

        # needed to set libs for mix2nix
        lib = pkgs.lib;
        mix2nix = pkgs.mix2nix;

        installHook = { release }: ''
          export APP_VERSION="${version}"
          export APP_NAME="${pname}"
          export ELIXIR_RELEASE="${props.elixir_release}"
          runHook preInstall
          mix release --no-deps-check --path "$out" ${release}
          runHook postInstall
        '';

        # src of the project
        src = ./.;
        # mix2nix dependencies
        mixNixDeps = import ./deps.nix { inherit lib beamPackages; };

        # mix release definition
        release-prod = beamPackages.mixRelease {
          inherit src pname version mixNixDeps elixir;
          mixEnv = "prod";

          installPhase = installHook { release = "prod"; };
        };

        release-dev = beamPackages.mixRelease {
          inherit src pname version mixNixDeps elixir;
          mixEnv = "dev";
          enableDebugInfo = true;
          installPhase = installHook { release = "dev"; };
        };
      in
      rec {
        # packages to build
        packages = {
          prod = release-prod;
          dev = release-dev;
          container = pkgs.dockerTools.buildImage {
            name = pname;
            tag = packages.prod.version;
            # required extra packages to make release work
            contents =
              [ packages.prod pkgs.coreutils pkgs.gnused pkgs.gnugrep ];
            created = "now";
            config.Entrypoint = [ "${packages.prod}/bin/prod" ];
            config.Cmd = [ "version" ];
          };
          oci = pkgs.ociTools.buildContainer {
            args = [ "${packages.prod}/bin/prod" ];
          };
          default = packages.prod;
        };

        # apps to run with nix run
        apps = {
          prod = flake-utils.lib.mkApp {
            name = pname;
            drv = packages.prod;
            exePath = "/bin/prod";
          };
          dev = flake-utils.lib.mkApp {
            name = "${pname}-dev";
            drv = packages.dev;
            exePath = "/bin/dev";
          };
          default = apps.prod;
        };

        # Module for deployment
        nixosModules.bonfire = import ./nix/module.nix;
        nixosModule = nixosModules.bonfire;

        devShells.default = pkgs.mkShell {

          shellHook = ''
            export APP_VERSION="${version}"
            export APP_NAME="${pname}"
            export ELIXIR_MAJOR_RELEASE="${props.elixir_release}"
            export MIX_HOME="$PWD/.cache/mix";
            export HEX_HOME="$PWD/.cache/hex";
            export MIX_PATH="${hex}/lib/erlang/lib/hex/ebin"
            export PATH="$MIX_PATH/bin:$HEX_HOME/bin:$PATH"
            mix local.rebar --if-missing rebar3 ${rebar3}/bin/rebar3;
            mix local.rebar --if-missing rebar ${rebar}/bin/rebar;

            export PGDATA=$PWD/db
            export PGHOST=$PWD/db
            export PGUSERNAME=${props.PGUSERNAME}
            export PGPASS=${props.PGPASS}
            export PGDATABASE=${props.PGDATABASE}
            export POSTGRES_USER=${props.PGUSERNAME}
            export POSTGRES_PASSWORD=${props.PGPASS}
            export POSTGRES_DB=${props.PGDATABASE}
            if [[ ! -d $PGDATA ]]; then
              mkdir $PGDATA
              # comment out if not using CoW fs
              chattr +C $PGDATA
              initdb -D $PGDATA
            fi
          '';

          buildInputs = [
            elixir
            erlang
            mix2nix
            locality
            rebar3
            rebar
            pkgs.yarn
            pkgs.cargo
            pkgs.rustc
            (pkgs.postgresql_12.withPackages (p: [ p.postgis ]))
          ] ++ optional pkgs.stdenv.isLinux
            pkgs.libnotify # For ExUnit Notifier on Linux.
          ++ optional pkgs.stdenv.isLinux
            pkgs.meilisearch # For meilisearch when running linux only
          ++ optional pkgs.stdenv.isLinux
            pkgs.inotify-tools; # For file_system on Linux.
        };
      });
}`

// https://github.com/GaloyMoney/blink/blob/17ef7b3ecba422d9fbcf1ebfb3f3a8c800f7bf58/flake.nix#L308
// Expected: "dockerImage"
const blinkNix = `{
  description = "Galoy dev environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    concourse-shared.url = "github:galoymoney/concourse-shared";

    rust-overlay = {
      url = "github:oxalica/rust-overlay";
      inputs = {
        nixpkgs.follows = "nixpkgs";
        flake-utils.follows = "flake-utils";
      };
    };
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    concourse-shared,
    rust-overlay,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      overlays = [
        (self: super: {
          nodejs = super.nodejs_20;
          pnpm = super.nodePackages.pnpm;
        })
        (import rust-overlay)
      ];
      pkgs = import nixpkgs {inherit overlays system;};
      rustVersion = pkgs.rust-bin.fromRustupToolchainFile ./rust-toolchain.toml;
      rust-toolchain = rustVersion.override {
        extensions = ["rust-analyzer" "rust-src"];
      };

      buck2NativeBuildInputs = with pkgs; [
        buck2
        protobuf
        nodejs
        pnpm
        python3
        ripgrep
        cacert
        clang
        lld
        rust-toolchain
      ];

      nativeBuildInputs = with pkgs;
        [
          envsubst
          nodejs
          tilt
          typescript
          bats
          postgresql
          alejandra
          gnumake
          docker
          docker-compose
          shellcheck
          shfmt
          vendir
          jq
          ytt
          sqlx-cli
          cargo-nextest
          cargo-audit
          cargo-watch
          reindeer
          gitMinimal
          grpcurl
          buf
          netcat
        ]
        ++ buck2NativeBuildInputs
        ++ lib.optionals pkgs.stdenv.isLinux [
          xvfb-run
          cypress
        ];

      buck2BuildInputs = with pkgs;
        []
        ++ lib.optionals pkgs.stdenv.isDarwin [
          darwin.apple_sdk.frameworks.SystemConfiguration
        ];

      buck2Version = pkgs.buck2.version;
      postPatch = with pkgs; ''
        rg -l '#!(/usr/bin/env|/bin/bash|/bin/sh)' prelude toolchains \
          | while read -r file; do
            patchShebangs --build "$file"
          done

        rg -l '(/usr/bin/env|/bin/bash)' prelude toolchains \
          | while read -r file; do
            substituteInPlace "$file" \
              --replace /usr/bin/env "${coreutils}/bin/env" \
              --replace /bin/bash "${bash}/bin/bash"
          done
      '';

      tscDerivation = {
        pkgName,
        pathPrefix ? "core",
      }:
        pkgs.stdenv.mkDerivation {
          bin_target = pkgName;
          deps_target = "prod_build";

          name = pkgName;
          buck2_target = "//${pathPrefix}/${pkgName}";
          __impure = true;
          src = ./.;
          nativeBuildInputs = buck2NativeBuildInputs;
          inherit postPatch;

          buildPhase = ''
            export HOME="$(dirname $(pwd))/home"
            buck2 build "$buck2_target" --verbose 8

            deps_result=$(buck2 build --show-simple-output "$buck2_target:$deps_target" 2> /dev/null)
            bin_result=$(buck2 build --show-simple-output "$buck2_target:$bin_target" 2> /dev/null)

            mkdir -p build/$name-$system/bin

            echo "$(pwd)/$deps_result" > build/$name-$system/buck2-deps-path

            cp -rpv $deps_result build/$name-$system/lib
            cp -rpv $bin_result build/$name-$system/bin/
          '';

          installPhase = ''
            mkdir -pv "$out"
            cp -rpv "build/$name-$system/lib" "$out/"
            cp -rpv "build/$name-$system/bin" "$out/"

            substituteInPlace "$out/bin/run" \
              --replace "#!${pkgs.coreutils}/bin/env sh" "#!${pkgs.bash}/bin/sh" \
              --replace "$(cat build/$name-$system/buck2-deps-path)" "$out/lib" \
              --replace "exec node" "exec ${pkgs.nodejs}/bin/node"
          '';
        };

      nextDerivation = {
        pkgName,
        pathPrefix ? "apps",
      }:
        pkgs.stdenv.mkDerivation {
          bin_target = pkgName;
          name = pkgName;
          buck2_target = "//${pathPrefix}/${pkgName}";
          __impure = true;
          src = ./.;
          nativeBuildInputs = buck2NativeBuildInputs;
          inherit postPatch;

          buildPhase = ''
            export HOME="$(dirname $(pwd))/home"
            mkdir -p build

            buck2 build "$buck2_target" --verbose 8

            result=$(buck2 build --show-simple-output "$buck2_target" 2> /dev/null)

            mkdir -p "build/$name-$system"
            cp -rpv "$result" "build/$name-$system/"
          '';

          installPhase = ''
            mkdir -pv "$out"
            cp -rpv build/$name-$system/app/* "$out/"

            # Need to escape this shell variable which should not be
            # iterpreted in Nix as a variable nor a shell variable when run
            # but rather a literal string which happens to be a shell
            # variable. Nuclear arms race of quoting and escaping special
            # characters to make this work...
            substituteInPlace "$out/bin/run" \
              --replace "#!${pkgs.coreutils}/bin/env sh" "#!${pkgs.bash}/bin/sh" \
              --replace "\''${0%/*}/../lib/" "$out/lib/" \
              --replace "exec node" "exec ${pkgs.nodejs}/bin/node"
          '';
        };

      npmDerivation = {
        pkgName,
        binTarget,
        npmBinTarget,
        pathPrefix ? "core",
      }:
        pkgs.stdenv.mkDerivation {
          name = "${binTarget}-${pkgName}";
          buck2_target = "//${pathPrefix}/${pkgName}";
          bin_target = binTarget;
          npm_bin_target = npmBinTarget;
          __impure = true;
          src = ./.;
          nativeBuildInputs = buck2NativeBuildInputs;
          inherit postPatch;

          buildPhase = ''
            export HOME="$(dirname $(pwd))/home"
            buck2 build "$buck2_target" --verbose 8

            npm_bin_result=$(buck2 build --show-simple-output "$buck2_target:$npm_bin_target" 2> /dev/null)
            bin_result=$(buck2 build --show-simple-output "$buck2_target:$bin_target" 2> /dev/null)
            bin_parent_path=$(dirname $(dirname $bin_result))
            node_modules_result=$(buck2 build --show-simple-output "$buck2_target:node_modules" 2> /dev/null)

            mkdir -p build/$name-$system/bin

            echo "$(pwd)/$npm_bin_result" > build/$name-$system/buck2-npm-bin-path
            echo "$(pwd)/$node_modules_result" > build/$name-$system/buck2-node-modules-path
            echo "$(pwd)/$bin_parent_path/__workspace" > build/$name-$system/buck2-deps-path

            mv $bin_parent_path/__workspace build/$name-$system/lib
            cp -rpv $bin_result build/$name-$system/bin/
            cp -rpv $npm_bin_result build/$name-$system/bin/
          '';

          installPhase = ''
            mkdir -pv "$out"
            cp -rpv "build/$name-$system/lib" "$out/"
            cp -rpv "build/$name-$system/bin" "$out/"

            npm_bin=$(cat build/$name-$system/buck2-npm-bin-path)

            npm_bin_parent_path=$(dirname $npm_bin)
            substituteInPlace "$out/bin/run" \
              --replace "#!${pkgs.coreutils}/bin/env sh" "#!${pkgs.bash}/bin/sh" \
              --replace "$npm_bin_parent_path" "$out/bin" \
              --replace "$(cat build/$name-$system/buck2-deps-path)" "$out/lib"

            npm_bin_name=$(basename $npm_bin)
            substituteInPlace "$out/bin/$npm_bin_name" \
              --replace "#!${pkgs.coreutils}/bin/env sh" "#!${pkgs.bash}/bin/sh" \
              --replace "$(cat build/$name-$system/buck2-node-modules-path)" "$out/lib"

            npm_bin_source_file=$(cat "$out/bin/$npm_bin_name" | grep "exec" | awk '{print $2}')
            substituteInPlace "$npm_bin_source_file" \
              --replace "$(cat build/$name-$system/buck2-node-modules-path)" "$out/lib" \
              --replace "exec node" "exec ${pkgs.nodejs}/bin/node" \
              --replace " sed " " ${pkgs.gnused}/bin/sed " \
              --replace "dirname" "${pkgs.coreutils}/bin/dirname" \
              --replace "uname" "${pkgs.coreutils}/bin/uname"
          '';
        };

      rustDerivation = {
        pkgName,
        pathPrefix ? "core",
      }:
        pkgs.stdenv.mkDerivation {
          bin_target = pkgName;

          name = pkgName;
          buck2_target = "//${pathPrefix}/${pkgName}";
          src = ./.;
          nativeBuildInputs = buck2NativeBuildInputs;
          inherit postPatch;

          buildPhase = ''
            export HOME="$(dirname $(pwd))/home"
            buck2 build "$buck2_target" --verbose 8

            result=$(buck2 build --show-simple-output "$buck2_target:$bin_target" 2> /dev/null)

            mkdir -p build/$name-$system/bin
            cp -rpv $result build/$name-$system/bin/
          '';

          installPhase = ''
            mkdir -pv "$out"
            cp -rpv "build/$name-$system/bin" "$out/"
          '';
        };

      gh-token = concourse-shared.packages.${system}.gh-token;
    in
      with pkgs; {
        packages = {
          api = tscDerivation {pkgName = "api";};
          api-trigger = tscDerivation {pkgName = "api-trigger";};
          api-ws-server = tscDerivation {pkgName = "api-ws-server";};
          api-exporter = tscDerivation {pkgName = "api-exporter";};
          api-cron = tscDerivation {pkgName = "api-cron";};

          consent = nextDerivation {pkgName = "consent";};
          dashboard = nextDerivation {pkgName = "dashboard";};
          pay = nextDerivation {pkgName = "pay";};
          admin-panel = nextDerivation {pkgName = "admin-panel";};
          map = nextDerivation {pkgName = "map";};
          voucher = nextDerivation {pkgName = "voucher";};

          migrate-mongo = npmDerivation {
            pkgName = "api";
            binTarget = "migrate-mongo-up";
            npmBinTarget = "migrate_mongo_bin";
          };

          api-keys = rustDerivation {pkgName = "api-keys";};
          notifications = rustDerivation {pkgName = "notifications";};

          dockerImage = dockerTools.buildImage {
            name = "galoy-dev";
            tag = "latest";

            # Optional base image to bring in extra binaries for debugging etc.
            fromImage = dockerTools.pullImage {
              imageName = "ubuntu";
              imageDigest = "sha256:496a9a44971eb4ac7aa9a218867b7eec98bdef452246c037aa206c841b653e08";
              sha256 = "sha256-LYdoE40tYih0XXJoJ8/b1e/IAkO94Jrs2C8oXWTeUTg=";
              finalImageTag = "mantic-20240122";
              finalImageName = "ubuntu";
            };

            config = {
              Cmd = ["bash"];
              Env = [
                "GIT_SSL_CAINFO=${cacert}/etc/ssl/certs/ca-bundle.crt"
                "SSL_CERT_FILE=${cacert}/etc/ssl/certs/ca-bundle.crt"
              ];
            };

            copyToRoot = buildEnv {
              name = "image-root";
              paths =
                nativeBuildInputs
                ++ [
                  bash
                  yq-go
                  google-cloud-sdk
                  gh
                  gh-token
                  openssh
                  rsync
                  git-cliff
                  unixtools.xxd
                ];

              pathsToLink = ["/bin"];
            };
          };
        };

        devShells.default = mkShell {
          inherit nativeBuildInputs;
          buildInputs = buck2BuildInputs;

          BUCK2_VERSION = buck2Version;
          COMPOSE_PROJECT_NAME = "galoy-dev";
        };

        formatter = alejandra;
      });
}`

// Expected: "image"
// https://github.com/V-Mann-Nick/huffman/blob/8efa0520c3e4e1e9151e79ce0fb78447a7bb94a3/flake.nix#L42
const huffmanNix = `{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    fenix = {
      url = "github:nix-community/fenix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    fenix,
    nixpkgs,
    ...
  }: let
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};
    rust = fenix.packages.${system};
    toolchain = rust.stable.withComponents [
      "rustc"
      "rust-std"
      "cargo"
      "rust-docs"
      "rustfmt"
      "clippy"
      "rust-src"
      "rust-analyzer"
    ];
    rustPlatform = pkgs.makeRustPlatform {
      cargo = toolchain;
      rustc = toolchain;
    };
    cargoToml = builtins.fromTOML (builtins.readFile ./Cargo.toml);
    pname = cargoToml.package.name;
    version = cargoToml.package.version;
  in {
    packages.${system} = rec {
      default = rustPlatform.buildRustPackage {
        inherit pname version;
        src = ./.;
        cargoLock.lockFile = ./Cargo.lock;
      };
      image = pkgs.dockerTools.buildImage {
        name = default.pname;
        config.Entrypoint = ["${default}/bin/huf"];
      };
    };
    devShells.${system}.default = pkgs.mkShell {
      name = pname;
      buildInputs = [toolchain];
      RUST_SRC_PATH = "${toolchain}/lib/rustlib/src/rust/library";
    };
    formatter.${system} = pkgs.alejandra;
  };
}`

// Expected: "Docker"
// https://github.com/clr-cera/Quartz/blob/92fa1f3810d1fc08c1b2a4889b77e1bc640130a1/flake.nix#L44
const quartzNix = `{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        inherit (nixpkgs.lib) getExe;

        pkgs = import nixpkgs {
          inherit system;
        };

        my-python-packages = ps:
          with ps; [
            dnspython
            build
            setuptools
          ];
      in rec
      {
        packages = rec {
          default = Quartz;

          Quartz = pkgs.python3Packages.buildPythonApplication {
            format = "pyproject";
            name = "Quartz";
            pname = "Quartz";
            src = ./.;
            propagatedBuildInputs = [(pkgs.python3.withPackages my-python-packages)];
          };

          Server = pkgs.writeShellScriptBin "Quartz-Server" ''
            python ${Quartz}/lib/python3.10/site-packages/Quartz/server.py
          '';

          Docker = pkgs.dockerTools.buildImage {
            name = "quartz-docker";
            tag = "latest";
            copyToRoot = pkgs.buildEnv {
              name = "image-root";
              paths = [pkgs.coreutils pkgs.bash Quartz Server (pkgs.python3.withPackages my-python-packages)];
              pathsToLink = ["/bin"];
            };

            config = {
              Cmd = ["/bin/Quartz-Server"];
            };
          };
        };
        apps = {
          default = {
            type = "app";
            program = "${packages.Quartz}/bin/Quartz";
          };

          Server = {
            type = "app";
            program = getExe packages.Server;
          };
        };
      }
    );
}`

// expected: "docker"
// https://github.com/mrkirby153/task-scheduler/blob/38bd43db43af6fac939499a9088b4b6033b60980/flake.nix#L27
const taskSchedulerNix = `{
  description = "A task scheduler";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
      buildInputs = with pkgs; [
        elixir
      ];
    in rec {
      packages.default = pkgs.beamPackages.mixRelease {
        version = "v0.1.0";
        name = "task-scheduler";
        pname = "task-scheduler";
        src = ./.;
        mixNixDeps = with pkgs; import ./mix_deps.nix { inherit lib beamPackages; };
      };
      packages.docker = pkgs.dockerTools.buildImage {
        name = "task-scheduler";
        tag = "latest";
        copyToRoot = pkgs.buildEnv {
          name = "image-root";
          paths = [ packages.default pkgs.busybox ];
          pathsToLink = [ "/bin" ];
        };
        runAsRoot = ''
        #!${pkgs.runtimeShell}
        mkdir -p /data
        '';
        config = {
          Cmd = [ "${packages.default}/bin/task_scheduler" "start" ];
          WorkingDir = "/data";
          Env = [ "LOCALE=en_US.UTF-8" ];
        };
      };
      devShells = {
        default = pkgs.mkShell {
          buildInputs = buildInputs ++ [
            pkgs.mix2nix
          ];
        };
      };
    });
}`
