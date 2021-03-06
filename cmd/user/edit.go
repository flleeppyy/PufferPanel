/*
 Copyright 2020 Padduck, LLC
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

package user

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/jinzhu/gorm"
	"github.com/pufferpanel/pufferpanel/v2"
	"github.com/pufferpanel/pufferpanel/v2/database"
	"github.com/pufferpanel/pufferpanel/v2/services"
	"github.com/spf13/cobra"
)

var EditUserCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a user",
	Run:   editUser,
	Args:  cobra.NoArgs,
}

func editUser(cmd *cobra.Command, args []string) {
	err := pufferpanel.LoadConfig("")
	if err != nil {
		fmt.Printf("Error loading config: %s", err.Error())
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		fmt.Printf("Error connecting to database: %s", err.Error())
		return
	}
	defer database.Close()

	var username string
	_ = survey.AskOne(&survey.Input{
		Message: "Username:",
	}, &username, survey.WithValidator(survey.Required))

	us := &services.User{DB: db}

	user, err := us.Get(username)
	if err != nil && gorm.IsRecordNotFoundError(err) {
		fmt.Printf("No user with username '%s'\n", username)
		return
	} else if err != nil {
		fmt.Printf("Error getting user: %s\n", err.Error())
		return
	}

	action := ""
	_ = survey.AskOne(&survey.Select{
		Message: "Select option to edit",
		Options: []string{"Username", "Email", "Password", "Change Admin Status"},
	}, &action)

	switch action {
	case "Username":
		{
			prompt := ""
			_ = survey.AskOne(&survey.Input{
				Message: "New Username:",
			}, &prompt, survey.WithValidator(survey.Required))
			user.Username = prompt

			err = us.Update(user)
			if err != nil {
				fmt.Printf("Error updating username: %s\n", err.Error())
			}
		}
	case "Email":
		{
			prompt := ""
			_ = survey.AskOne(&survey.Input{
				Message: "New Email:",
			}, &prompt, survey.WithValidator(survey.Required))
			user.Email = prompt

			err = us.Update(user)
			if err != nil {
				fmt.Printf("Error updating email: %s\n", err.Error())
			}
		}
	case "Password":
		{
			prompt := ""
			_ = survey.AskOne(&survey.Password{
				Message: "New Password:",
			}, &prompt, survey.WithValidator(validatePassword))

			err = user.SetPassword(prompt)
			if err != nil {
				fmt.Printf("Error updating password: %s\n", err.Error())
			}

			err = us.Update(user)
			if err != nil {
				fmt.Printf("Error updating password: %s\n", err.Error())
			}
		}
	case "Change Admin Status":
		{
			prompt := false
			_ = survey.AskOne(&survey.Confirm{
				Message: "Set Admin Status: ",
			}, &prompt)

			ps := &services.Permission{DB: db}
			perms, err := ps.GetForUserAndServer(user.ID, nil)
			if err != nil {
				fmt.Printf("Error updating permissions: %s\n", err.Error())
				return
			}

			perms.Admin = prompt

			err = ps.UpdatePermissions(perms)
			if err != nil {
				fmt.Printf("Error updating password: %s\n", err.Error())
			}
		}
	}
}
