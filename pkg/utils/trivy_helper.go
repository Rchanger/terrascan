package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"time"

	"github.com/accurics/terrascan/pkg/iac-providers/output"
	"github.com/aquasecurity/fanal/analyzer"
	"github.com/aquasecurity/fanal/analyzer/config"
	aquaArtifactimage "github.com/aquasecurity/fanal/artifact/image"
	"github.com/aquasecurity/fanal/image"
	"github.com/aquasecurity/trivy/pkg/cache"
	"github.com/aquasecurity/trivy/pkg/rpc/client"
	"github.com/aquasecurity/trivy/pkg/scanner"
	"github.com/aquasecurity/trivy/pkg/types"
	"go.uber.org/zap"
)

var (
	serverURL      = "http://localhost:8080"
	vulnType       = []string{"os", "library"}
	securityChecks = []string{"vuln"}
	trivyCliFlags  = []string{"-q", "-f", "json"}
	trivyBinary    = "trivy"
)

// ScanDockerImageWithTrivyServer creates trivy protobuf client and run scan on provided image
func ScanDockerImageWithTrivyServer(ctx context.Context, imageName string) (output.Results, error) {
	vulnerabilities := output.Results{}
	sc, cleanUp, err := initializeDockerScanner(ctx, imageName, client.CustomHeaders{}, client.RemoteURL(serverURL), time.Second*5000)
	if err != nil {
		zap.S().Error("could not initialize scanner: %v", err)
		return vulnerabilities, err
	}

	defer cleanUp()

	results, err := sc.ScanArtifact(ctx, types.ScanOptions{
		VulnType:            vulnType,
		SecurityChecks:      securityChecks,
		ScanRemovedPackages: false,
		ListAllPackages:     false,
	})

	if err != nil {
		zap.S().Error("error scaninng image %s: %v", imageName, err)
		return vulnerabilities, err
	}

	byteData, err := json.Marshal(results)
	if err != nil {
		zap.S().Error("error marshaling string: %v", results)
		return vulnerabilities, err
	}

	err = json.Unmarshal(byteData, &vulnerabilities)
	if err != nil {
		zap.S().Error("error unmarshaling string: %v", err)
		return vulnerabilities, err
	}
	if err != nil {
		zap.S().Error("could not scan image: %v", err)
		return vulnerabilities, err
	}
	return vulnerabilities, nil
}

//initializeDockerScanner creates client scanner
func initializeDockerScanner(ctx context.Context, imageName string, customHeaders client.CustomHeaders, url client.RemoteURL, timeout time.Duration) (scanner.Scanner, func(), error) {
	protoClient := client.NewProtobufClient(url)
	clientScanner := client.NewScanner(customHeaders, protoClient)
	artifactCache := cache.NewRemoteCache(cache.RemoteURL(url), nil)
	dockerOption, err := types.GetDockerOption(timeout)
	if err != nil {
		return scanner.Scanner{}, nil, err
	}

	image, cleanup, err := image.NewDockerImage(ctx, imageName, dockerOption)
	if err != nil {
		return scanner.Scanner{}, nil, err
	}

	artifact, err := aquaArtifactimage.NewArtifact(image, artifactCache, []analyzer.Type{}, config.ScannerOption{})
	if err != nil {
		cleanup()
		return scanner.Scanner{}, nil, err
	}

	scanner2 := scanner.NewScanner(clientScanner, artifact)
	return scanner2, func() {
		cleanup()
	}, nil
}

// ScanDockerImageWithTrivyCLI
func ScanDockerImageWithTrivyCLI(imageName string) (output.Results, error) {
	results := output.Results{}
	args := append(trivyCliFlags, imageName)

	cmd := exec.Command(trivyBinary, args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		zap.S().Error("error calling trivy for scanning image %s : %v", imageName, zap.Error(err))
		return results, err
	}

	err = json.Unmarshal(out.Bytes(), &results)
	if err != nil {
		zap.S().Error("error unmarshalling trivy result for image: %s", imageName, zap.Error(err))
		return results, err
	}
	return results, nil
}
