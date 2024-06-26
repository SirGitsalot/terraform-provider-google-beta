// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package networksecurity_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-provider-google-beta/google-beta/acctest"
	"github.com/hashicorp/terraform-provider-google-beta/google-beta/envvar"
	"github.com/hashicorp/terraform-provider-google-beta/google-beta/tpgresource"
	transport_tpg "github.com/hashicorp/terraform-provider-google-beta/google-beta/transport"
)

func TestAccNetworkSecurityFirewallEndpointAssociations_basic(t *testing.T) {
	acctest.SkipIfVcr(t)
	t.Parallel()

	orgId := envvar.GetTestOrgFromEnv(t)
	randomSuffix := acctest.RandString(t, 10)

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderBetaFactories(t),
		CheckDestroy:             testAccCheckNetworkSecurityFirewallEndpointDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkSecurityFirewallEndpointAssociation_basic(randomSuffix, orgId),
			},
			{
				ResourceName:            "google_network_security_firewall_endpoint_association.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"labels", "terraform_labels"},
			},
			{
				Config: testAccNetworkSecurityFirewallEndpointAssociation_update(randomSuffix, orgId),
			},
			{
				ResourceName:            "google_network_security_firewall_endpoint_association.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"labels", "terraform_labels"},
			},
		},
	})
}

func testAccNetworkSecurityFirewallEndpointAssociation_basic(randomSuffix string, orgId string) string {
	return fmt.Sprintf(`
resource "google_compute_network" "foobar" {
	provider                = google-beta
    name                    = "tf-test-my-vpc%s"
    auto_create_subnetworks = false
}

resource "google_network_security_firewall_endpoint" "foobar" {
    provider = google-beta
    name     = "tf-test-my-firewall-endpoint%s"
    parent   = "organizations/%s"
    location = "us-central1-a"
}

# TODO: add tlsInspectionPolicy once resource is ready
resource "google_network_security_firewall_endpoint_association" "foobar" {
    provider          = google-beta
    name              = "tf-test-my-firewall-endpoint%s"
    parent            = "organizations/%s"
    location          = "us-central1-a"
    firewall_endpoint = google_network_security_firewall_endpoint.foobar.id
    network           = google_compute_network.foobar.id

    labels = {
        foo = "bar"
    }
}
`, randomSuffix, randomSuffix, orgId, randomSuffix, orgId)
}

func testAccNetworkSecurityFirewallEndpointAssociation_update(randomSuffix string, orgId string) string {
	return fmt.Sprintf(`
resource "google_compute_network" "foobar" {
	provider                = google-beta
    name                    = "tf-test-my-vpc%s"
    auto_create_subnetworks = false
}

resource "google_network_security_firewall_endpoint" "foobar" {
    provider = google-beta
    name     = "tf-test-my-firewall-endpoint%s"
    parent   = "organizations/%s"
    location = "us-central1-a"
}

# TODO: add tlsInspectionPolicy once resource is ready
resource "google_network_security_firewall_endpoint_association" "foobar" {
    provider          = google-beta
    name              = "tf-test-my-firewall-endpoint%s"
    parent            = "organizations/%s"
    location          = "us-central1-a"
    firewall_endpoint = google_network_security_firewall_endpoint.foobar.id
    network           = google_compute_network.foobar.id

    labels = {
        foo = "bar-updated"
    }
}
`, randomSuffix, randomSuffix, orgId, randomSuffix, orgId)
}

func testAccCheckNetworkSecurityFirewallEndpointAssociationDestroyProducer(t *testing.T) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "google_network_security_firewall_endpoint_association" {
				continue
			}
			if strings.HasPrefix(name, "data.") {
				continue
			}

			config := acctest.GoogleProviderConfig(t)

			url, err := tpgresource.ReplaceVarsForTest(config, rs, "{{NetworkSecurityBasePath}}{{parent}}/locations/{{location}}/firewallEndpointAssociations/{{name}}")
			if err != nil {
				return err
			}

			billingProject := ""

			if config.BillingProject != "" {
				billingProject = config.BillingProject
			}

			_, err = transport_tpg.SendRequest(transport_tpg.SendRequestOptions{
				Config:    config,
				Method:    "GET",
				Project:   billingProject,
				RawURL:    url,
				UserAgent: config.UserAgent,
			})
			if err == nil {
				return fmt.Errorf("NetworkSecurityFirewallEndpointAssociation still exists at %s", url)
			}
		}

		return nil
	}
}
