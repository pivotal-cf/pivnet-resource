package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
)

type UserGroupsCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1"`
}

type UserGroupCommand struct {
	UserGroupID int `long:"user-group-id" description:"User group ID e.g. 1234" required:"true"`
}

type CreateUserGroupCommand struct {
	Name        string   `long:"name" description:"Name e.g. all_users" required:"true"`
	Description string   `long:"description" description:"Description e.g. 'All users in the world'" required:"true"`
	Members     []string `long:"member" description:"Email addresses of members to be added"`
}

type UpdateUserGroupCommand struct {
	UserGroupID int     `long:"user-group-id" description:"User group ID e.g. 1234" required:"true"`
	Name        *string `long:"name" description:"Name e.g. all_users"`
	Description *string `long:"description" description:"Description e.g. 'All users in the world'"`
}

type AddUserGroupCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1" required:"true"`
	UserGroupID    int    `long:"user-group-id" description:"User Group ID e.g. 1234" required:"true"`
}

type DeleteUserGroupCommand struct {
	UserGroupID int `long:"user-group-id" description:"User group ID e.g. 1234" required:"true"`
}

func (command *UserGroupCommand) Execute([]string) error {
	client := NewClient()

	userGroup, err := client.UserGroups.Get(
		command.UserGroupID,
	)
	if err != nil {
		return err
	}

	return printUserGroup(userGroup)
}

func (command *UserGroupsCommand) Execute([]string) error {
	client := NewClient()

	if command.ProductSlug == "" && command.ReleaseVersion == "" {
		var err error
		userGroups, err := client.UserGroups.List()
		if err != nil {
			return err
		}

		return printUserGroups(userGroups)
	}

	if command.ProductSlug == "" || command.ReleaseVersion == "" {
		return fmt.Errorf("Both or neither of product slug and release version must be provided")
	}

	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return err
	}

	var release pivnet.Release
	for _, r := range releases {
		if r.Version == command.ReleaseVersion {
			release = r
			break
		}
	}

	if release.Version != command.ReleaseVersion {
		return fmt.Errorf("release not found")
	}

	userGroups, err := client.UserGroups.ListForRelease(command.ProductSlug, release.ID)
	if err != nil {
		return err
	}

	return printUserGroups(userGroups)
}

func printUserGroups(userGroups []pivnet.UserGroup) error {
	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Name", "Description"})

		for _, u := range userGroups {
			table.Append([]string{
				strconv.Itoa(u.ID),
				u.Name,
				u.Description,
			})
		}
		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(userGroups)
	case PrintAsYAML:
		return printYAML(userGroups)
	}

	return nil
}

func (command *CreateUserGroupCommand) Execute([]string) error {
	client := NewClient()

	userGroup, err := client.UserGroups.Create(command.Name, command.Description, command.Members)
	if err != nil {
		return err
	}

	return printUserGroup(userGroup)
}

func printUserGroup(userGroup pivnet.UserGroup) error {
	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Name", "Description", "Members"})

		table.Append([]string{
			strconv.Itoa(userGroup.ID),
			userGroup.Name,
			userGroup.Description,
			strings.Join(userGroup.Members, ", "),
		})

		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(userGroup)
	case PrintAsYAML:
		return printYAML(userGroup)
	}

	return nil
}

func (command *DeleteUserGroupCommand) Execute([]string) error {
	client := NewClient()

	err := client.UserGroups.Delete(command.UserGroupID)
	if err != nil {
		return err
	}

	if Pivnet.Format == PrintAsTable {
		_, err = fmt.Fprintf(
			OutputWriter,
			"user group %d deleted successfully\n",
			command.UserGroupID,
		)
	}

	return nil
}

func (command *UpdateUserGroupCommand) Execute([]string) error {
	client := NewClient()

	userGroup, err := client.UserGroups.Get(command.UserGroupID)
	if err != nil {
		return err
	}

	if command.Name != nil {
		userGroup.Name = *command.Name
	}

	if command.Description != nil {
		userGroup.Description = *command.Description
	}

	updated, err := client.UserGroups.Update(userGroup)
	if err != nil {
		return err
	}

	return printUserGroup(updated)
}

func (command *AddUserGroupCommand) Execute([]string) error {
	client := NewClient()

	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return err
	}

	var release pivnet.Release
	for _, r := range releases {
		if r.Version == command.ReleaseVersion {
			release = r
			break
		}
	}

	if release.Version != command.ReleaseVersion {
		return fmt.Errorf("release not found")
	}

	err = client.UserGroups.AddToRelease(
		command.ProductSlug,
		release.ID,
		command.UserGroupID,
	)
	if err != nil {
		return err
	}

	if Pivnet.Format == PrintAsTable {
		_, err = fmt.Fprintf(
			OutputWriter,
			"user group %d added successfully to %s/%s\n",
			command.UserGroupID,
			command.ProductSlug,
			command.ReleaseVersion,
		)
	}

	return nil
}
