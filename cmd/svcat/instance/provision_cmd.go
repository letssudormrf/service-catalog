/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package instance

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/parameters"
	"github.com/spf13/cobra"
)

type provisonCmd struct {
	*command.Context
	ns           string
	instanceName string
	className    string
	planName     string
	rawParams    []string
	jsonParams   string
	params       interface{}
	rawSecrets   []string
	secrets      map[string]string
}

// NewProvisionCmd builds a "svcat provision" command
func NewProvisionCmd(cxt *command.Context) *cobra.Command {
	provisionCmd := &provisonCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:   "provision NAME --plan PLAN --class CLASS",
		Short: "Create a new instance of a service",
		Example: `
  svcat provision wordpress-mysql-instance --class mysqldb --plan free -p location=eastus -p sslEnforcement=disabled
  svcat provision wordpress-mysql-instance --class mysqldb --plan free -s mysecret[dbparams]
  svcat provision secure-instance --class mysqldb --plan secureDB --params-json '{
    "encrypt" : true,
    "firewallRules" : [
        {
            "name": "AllowSome",
            "startIPAddress": "75.70.113.50",
            "endIPAddress" : "75.70.113.131"
        },
        {
            "name": "AllowMore",
            "startIPAddress": "13.54.0.0",
            "endIPAddress" : "13.56.0.0"
        }
    ]
}
'
`,
		PreRunE: command.PreRunE(provisionCmd),
		RunE:    command.RunE(provisionCmd),
	}
	cmd.Flags().StringVarP(&provisionCmd.ns, "namespace", "n", "default",
		"The namespace in which to create the instance")
	cmd.Flags().StringVar(&provisionCmd.className, "class", "",
		"The class name (Required)")
	cmd.MarkFlagRequired("class")
	cmd.Flags().StringVar(&provisionCmd.planName, "plan", "",
		"The plan name (Required)")
	cmd.MarkFlagRequired("plan")
	cmd.Flags().StringSliceVarP(&provisionCmd.rawParams, "param", "p", nil,
		"Additional parameter to use when provisioning the service, format: NAME=VALUE. Cannot be combined with --params-json")
	cmd.Flags().StringSliceVarP(&provisionCmd.rawSecrets, "secret", "s", nil,
		"Additional parameter, whose value is stored in a secret, to use when provisioning the service, format: SECRET[KEY]")
	cmd.Flags().StringVar(&provisionCmd.jsonParams, "params-json", "",
		"Additional parameters to use when provisioning the service, provided as a JSON object. Cannot be combined with --param")
	return cmd
}

func (c *provisonCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("an instance name is required")
	}
	c.instanceName = args[0]

	var err error

	if c.jsonParams != "" && len(c.rawParams) > 0 {
		return fmt.Errorf("--params-json cannot be used with --param")
	}

	if c.jsonParams != "" {
		c.params, err = parameters.ParseVariableJSON(c.jsonParams)
		if err != nil {
			return fmt.Errorf("invalid --params value (%s)", err)
		}
	} else {
		c.params, err = parameters.ParseVariableAssignments(c.rawParams)
		if err != nil {
			return fmt.Errorf("invalid --param value (%s)", err)
		}
	}

	c.secrets, err = parameters.ParseKeyMaps(c.rawSecrets)
	if err != nil {
		return fmt.Errorf("invalid --secret value (%s)", err)
	}

	return nil
}

func (c *provisonCmd) Run() error {
	return c.Provision()
}

func (c *provisonCmd) Provision() error {
	instance, err := c.App.Provision(c.ns, c.instanceName, c.className, c.planName, c.params, c.secrets)
	if err != nil {
		return err
	}

	output.WriteInstanceDetails(c.Output, instance)

	return nil
}
