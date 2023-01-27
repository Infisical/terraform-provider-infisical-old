default: install

generate:
	go generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name infisical
	go generate oapi-codegen --package=client -generate=client,types -o ./client/infisical.gen.go https://raw.githubusercontent.com/Infisical/infisical/main/docs/spec.yaml

install:
	go install .

test:
	go test -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 10m -v ./...