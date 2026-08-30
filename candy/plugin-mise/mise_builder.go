package mise

// mise_builder.go — the builder:mise BUILD-TIME leg. OpResolve renders the
// multi-stage build that provisions the mise dev-tool manager (jdx/mise) into
// the image: a distro-aware FROM…AS stage installs the pinned mise release from
// the upstream tarball, stages MISE_DATA_DIR + the system config, and the
// CopyArtifacts deliver the binary, the shims, and /etc/mise/config.toml into
// the final image (post-main-FROM). The InlineFragment sets the runtime env.
//
// The external_builder: selection path passes MINIMAL OpResolve input (only the
// candy name + BuildEnv.Distros), so the stage picks its base from the target
// distro and provisions mise itself; the candy's `mise:` verb steps then install
// the tools in the main image. When the P3 detection path lands, the host
// computes the FULL BuilderResolveInput (CopySrc/Manifest/UID/GID/Home/cache
// mounts) and this same render switches to full mode: the candy's mise.toml is
// COPY'd into the stage and `mise install` runs there, so the tools land in the
// stage and are copied into the image.

import (
	"encoding/json"
	"fmt"
	"strings"

	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// miseVersion is the pinned mise release the builder stage installs. Bump
// deliberately (mise releases weekly); the tarball naming is Go-style arch
// (x64/arm64), mapped from uname at build time.
const miseVersion = "v2026.8.14"

// distroBases maps the charly distro families to a base image the stage can
// build on (curl+tar installable).
var distroBases = map[string]string{
	"fedora": "quay.io/fedora/fedora:43",
	"debian": "debian:13",
	"ubuntu": "ubuntu:24.04",
	"arch":   "archlinux:latest",
	"alpine": "alpine:3.21",
}

// distroPkgInstall maps a distro family to the package-manager line that
// installs curl + tar + gzip in the stage.
var distroPkgInstall = map[string]string{
	"fedora": "dnf -y install --setopt=install_weak_deps=False curl tar gzip",
	"debian": "apt-get update && apt-get -y install --no-install-recommends curl tar gzip",
	"ubuntu": "apt-get update && apt-get -y install --no-install-recommends curl tar gzip",
	"arch":   "pacman -Syu --noconfirm curl tar gzip",
	"alpine": "apk add --no-cache curl tar gzip",
}

// invokeResolve handles the builder OpResolve: it decodes the host-computed
// BuilderResolveInput (minimal on the external_builder path, full on the
// detection path) + the BuildEnv, renders the mise multi-stage build, and
// returns the BuilderResolveReply.
func invokeResolve(req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	var in spec.BuilderResolveInput
	if len(req.GetParamsJson()) > 0 {
		if err := json.Unmarshal(req.GetParamsJson(), &in); err != nil {
			return nil, fmt.Errorf("builder %q: decode resolve input: %w", word, err)
		}
	}
	var env spec.BuildEnv
	if len(req.GetEnvJson()) > 0 {
		_ = json.Unmarshal(req.GetEnvJson(), &env)
	}
	reply := renderMiseStage(in, env)
	j, err := json.Marshal(reply)
	if err != nil {
		return nil, err
	}
	return &pb.InvokeReply{ResultJson: j}, nil
}

// renderMiseStage renders the mise multi-stage build from the host-computed
// input. FULL mode (CopySrc + Manifest present — the P3 detection path) copies
// the candy's mise.toml into the stage and runs `mise install` there, so the
// tools land in the stage and are copied into the image. MINIMAL mode (the P1
// external_builder path) provisions mise itself and leaves tool installs to the
// candy's `mise:` verb steps in the main image.
func renderMiseStage(in spec.BuilderResolveInput, env spec.BuildEnv) spec.BuilderResolveReply {
	distro := "fedora"
	if len(env.Distros) > 0 && env.Distros[0] != "" {
		distro = env.Distros[0]
	}
	base := distroBases[distro]
	if base == "" {
		base = distroBases["fedora"]
	}
	pkg := distroPkgInstall[distro]
	if pkg == "" {
		pkg = distroPkgInstall["fedora"]
	}
	stageName := "mise-build"
	if in.StageName != "" {
		stageName = in.StageName
	}

	var stage strings.Builder
	stage.WriteString("FROM " + base + " AS " + stageName + "\n")
	stage.WriteString("USER root\n")
	stage.WriteString("RUN " + pkg + "\n")
	stage.WriteString("ENV MISE_DATA_DIR=/mise MISE_YES=1\n")
	stage.WriteString("RUN A=$(uname -m | sed 's/x86_64/x64/;s/aarch64/arm64/') && \\n")
	stage.WriteString("    curl -sL -o /tmp/mise.tar.gz https://github.com/jdx/mise/releases/download/" + miseVersion + "/mise-" + miseVersion + "-linux-$A.tar.gz && \\n")
	stage.WriteString("    tar xzf /tmp/mise.tar.gz -C /usr/local --strip-components=1 && rm /tmp/mise.tar.gz\n")

	if in.Manifest != "" && in.CopySrc != "" {
		// FULL mode: the candy's config drives the stage install.
		stage.WriteString("COPY --chown=" + fmt.Sprintf("%d:%d", in.UID, in.GID) + " " + in.CopySrc + "/" + in.Manifest + " mise.toml\n")
		stage.WriteString("RUN " + in.CacheMountsOwned + "mise install && mise reshim && mkdir -p /mise-out && cp -a /mise/. /mise-out/ && cp mise.toml /mise-out/config.toml\n")
	} else {
		// MINIMAL mode: provision mise + an empty system config; the candy's
		// `mise:` verb steps install the tools in the main image.
		stage.WriteString("RUN mkdir -p /mise-out && cp -a /mise/. /mise-out/ && printf '[tools]\\n' > /mise-out/config.toml\n")
	}

	artifacts := []string{
		"COPY --from=" + stageName + " /usr/local/bin/mise /usr/local/bin/mise",
		"COPY --from=" + stageName + " /mise-out/ /usr/local/share/mise/",
		"COPY --from=" + stageName + " /mise-out/config.toml /etc/mise/config.toml",
	}
	inline := "ENV MISE_DATA_DIR=/usr/local/share/mise\nENV PATH=/usr/local/share/mise/shims:$PATH\n"
	return spec.BuilderResolveReply{
		Stage:          stage.String(),
		CopyArtifacts:  artifacts,
		InlineFragment: inline,
	}
}
