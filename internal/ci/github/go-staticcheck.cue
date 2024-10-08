package github

workflows: "go-staticcheck": {
	name: "go-staticcheck"
	on: ["push", "pull_request"]
	jobs: ci: {
		name:      "Run CI"
		"runs-on": "ubuntu-latest"
		steps: [{
			uses: "actions/checkout@v4"
			with: "fetch-depth": 1
		}, {
			name: "Setup Go"
			uses: "actions/setup-go@v5"
			with: "go-version-file": "go.mod"
		}, {
			run: "go version"
		}, {
			name: "Install Node"
			uses: "actions/setup-node@v4"
			with: "node-version": 18
		}, {
			run: "npm --version"
		}, {
			name: "Install Just"
			uses: "extractions/setup-just@v1"
		}, {
			name: "Install build dependencies"
			run:  "just install-templ install-node-packages"
		}, {
			name: "Generate dependencies"
			run:  "just build-tw-css build-templ"
		}, {
			uses: "dominikh/staticcheck-action@v1"
			with: version: "latest"
		}]
	}
}
