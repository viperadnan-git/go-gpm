module github.com/viperadnan-git/go-gpm/cmd/gpcli

go 1.25

require (
	github.com/creativeprojects/go-selfupdate v1.5.2
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/urfave/cli/v3 v3.6.2
	github.com/viperadnan-git/go-gpm v1.1.0
)

require (
	code.gitea.io/sdk/gitea v0.23.2 // indirect
	github.com/42wim/httpsig v1.2.3 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/davidmz/go-pageant v1.0.2 // indirect
	github.com/go-fed/httpsig v1.1.0 // indirect
	github.com/google/go-github/v30 v30.1.0 // indirect
	github.com/google/go-github/v74 v74.0.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/xanzy/go-gitlab v0.115.0 // indirect
	gitlab.com/gitlab-org/api/client-go v1.34.0 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/oauth2 v0.35.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Use local library during development
replace github.com/viperadnan-git/go-gpm => ../..
