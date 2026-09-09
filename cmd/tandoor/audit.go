// cmd/tandoor/cmd_audit.go

package main

import "github.com/urfave/cli/v3"

// GetAuditCommand returns the top-level `audit` command group.
func GetAuditCommand() *cli.Command {
	return &cli.Command{
		Name:        "audit",
		Usage:       "Food audit and remediation commands",
		Description: "Audit foods for naming issues, duplicates and connectors. Use inspect/suggest/fix workflow.",
		Commands: []*cli.Command{
			auditFoodsCommand(),
			auditFoodCommand(),
			auditDuplicatesCommand(),
			auditConnectorsCommand(),
		},
	}
}
