package v6

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v6/shared"
	"github.com/SermoDigital/jose/jwt"
)

//go:generate counterfeiter . OauthTokenActor

type OauthTokenActor interface {
	RefreshAccessToken(refreshToken string) (string, error)
	ParseAccessToken(accessToken string) (jwt.JWT, error)
}

type OauthTokenCommand struct {
	usage           interface{} `usage:"CF_NAME oauth-token"`
	relatedCommands interface{} `related_commands:"curl"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       OauthTokenActor
}

func (cmd *OauthTokenCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd OauthTokenCommand) Execute(_ []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	if cmd.Config.UAAGrantType() == string(constant.GrantTypeClientCredentials) && cmd.Config.UAAOAuthClientSecret() == "" {
		token, err := cmd.Actor.ParseAccessToken(cmd.Config.AccessToken())
		if err != nil {
			return errors.New(cmd.UI.TranslateText("Access token is invalid."))
		}

		expiration, success := token.Claims().Expiration()
		if !success {
			return errors.New(cmd.UI.TranslateText("Access token is missing expiration claim."))
		}

		if expiration.Before(time.Now()) {
			return errors.New(cmd.UI.TranslateText("Access token has expired."))
		}

		cmd.UI.DisplayText(cmd.Config.AccessToken())
		return nil
	}

	accessToken, err := cmd.Actor.RefreshAccessToken(cmd.Config.RefreshToken())
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(accessToken)
	return nil
}
