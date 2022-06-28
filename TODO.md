# TODO List

- [X] Implement support for VLAN class parameters for version >= 7.2
- [X] Implement support for DNS views
- [-] Upgrade SDK to V2 (https://www.terraform.io/plugin/sdkv2/guides/v2-upgrade-guide)
  - [X] Migrate to the new provider
  - [ ] Migrate all Create/Reade/Delete Functions to their context aware equivalent (See IPSpace resource create for example)
  - [ ] Migrate all SchemaValidateFunc to SchemaValidateDiagFunc - https://www.terraform.io/plugin/sdkv2/guides/v2-upgrade-guide#deprecation-of-helper-schema-schemavalidatefunc
  - [ ] Remove all ExistsFunc that are now deprecated - https://www.terraform.io/plugin/sdkv2/guides/v2-upgrade-guide#deprecation-of-helper-schema-existsfunc
  - [ ] Check all deprecated validation function - https://www.terraform.io/plugin/sdkv2/guides/v2-upgrade-guide#removal-of-deprecated-validation-functions
  - [ ] Implement support for diagnostics
  - [ ] Implement resource-Level and field-Level descriptions
  - [ ] Leverage new validation from schema.Schema.Computed - https://www.terraform.io/plugin/sdkv2/guides/v2-upgrade-guide#stronger-validation-for-helper-schema-schema-computed-fields
- [ ] Implement context https://www.terraform.io/plugin/sdkv2/guides/v2-upgrade-guide#more-support-for-context-context
- [ ] Implement binary generation for https://www.terraform.io/registry/providers/os-arch
- [ ] Implement a new releaser https://goreleaser.com/install/
- [ ] Implement support for RPZ Zone and RPZ rules
- [ ] Implement support for DHCP resources
- [ ] Implement support for Subnet/VLAN relationship
- [ ] Implement support for SOLIDserver resources covering (NTP/SNMP/Admin & ipmadmin Passwords/Certificat SSL/Services)
- [ ] Increase test coverage based on https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html

# Useful Links

* https://godoc.org/github.com/hashicorp/terraform/helper/validation#pkg-index
* https://tutorialedge.net/golang/intro-testing-in-go/
* https://www.terraform.io/docs/registry/providers/publishing.html
* https://www.terraform.io/docs/registry/providers/docs.html