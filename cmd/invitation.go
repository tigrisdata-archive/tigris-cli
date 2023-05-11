// Copyright 2022-2023 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var (
	status string
	code   string
	emails []string
	email  string
	roles  []string

	sender string
)

var listInvitationsCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists invitations",
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.ManagementGet().ListInvitations(ctx, status)
			if err != nil {
				return util.Error(err, "list invitations")
			}

			err = util.PrettyJSON(resp)
			util.Fatal(err, "list invitations")

			return nil
		})
	},
}

var createInvitationCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates invitation(s)",
	Long:  "Creates invitations with provided email(s) and role(s)",
	Example: fmt.Sprintf(`%[1]s invitation create --email welcome@example.com --role reader`,
		rootCmd.Root().Name()),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			var invs []*driver.InvitationInfo

			for k, v := range emails {
				role := ""
				if k < len(roles) {
					role = roles[k]
				}
				invs = append(invs, &driver.InvitationInfo{
					Email:                v,
					Role:                 role,
					InvitationSentByName: sender,
				})
			}

			err := client.ManagementGet().CreateInvitations(ctx, invs)

			return util.Error(err, "create invitations")
		})
	},
}

var deleteInvitationCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an invitation",
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			for _, v := range emails {
				err := client.ManagementGet().DeleteInvitations(ctx, v, status)
				return util.Error(err, "delete invitations")
			}

			return nil
		})
	},
}

var verifyInvitationCmd = &cobra.Command{
	Use:   "verify {email} [code]",
	Short: "Verify invitation",
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			err := client.ManagementGet().VerifyInvitation(ctx, email, code)
			util.Fatal(err, "verify invitation")

			util.Stdoutf("Successfully verified\n")

			return nil
		})
	},
}

var invitationCmd = &cobra.Command{
	Use:     "invitation",
	Aliases: []string{"invitations"},
	Short:   "Invitation management commands",
	Args:    cobra.MinimumNArgs(1),
}

func init() {
	createInvitationCmd.Flags().StringSliceVarP(&emails, "email", "e", nil, "Email of the user to invite")
	createInvitationCmd.Flags().StringSliceVarP(&roles, "role", "r", nil, "Role of the user to invite")
	createInvitationCmd.Flags().StringVarP(&sender, "sender", "s", "", "Sender's name")

	listInvitationsCmd.Flags().StringVar(&status, "status", "s", "Check invitations with this status")

	deleteInvitationCmd.Flags().StringSliceVarP(&emails, "email", "e", nil, "Email of the user to delete invitation for")
	deleteInvitationCmd.Flags().StringVar(&status, "status", "", "Delete invitations with this status")

	verifyInvitationCmd.Flags().StringVarP(&email, "email", "e", "", "Email of the user to verify invitation for")
	verifyInvitationCmd.Flags().StringVarP(&code, "code", "c", "", "Invitation code to verify")

	invitationCmd.AddCommand(deleteInvitationCmd)
	invitationCmd.AddCommand(createInvitationCmd)
	invitationCmd.AddCommand(listInvitationsCmd)
	invitationCmd.AddCommand(verifyInvitationCmd)

	rootCmd.AddCommand(invitationCmd)
}
