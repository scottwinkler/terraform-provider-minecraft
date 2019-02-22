go build
mv terraform-provider-minecraft ./test/entity
rm -rf ./test/entity/.terraform
rm ./test/entity/terraform.tfstate
cd test/entity
terraform init
terraform apply -auto-approve