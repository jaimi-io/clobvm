package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/auth"
	"github.com/jaimi-io/clobvm/cmd/clob-cli/consts"
	"github.com/jaimi-io/clobvm/genesis"
	crpc "github.com/jaimi-io/clobvm/rpc"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/jaimi-io/hypersdk/rpc"
	"github.com/jaimi-io/hypersdk/utils"
	"github.com/manifoldco/promptui"
)

func promptChainID() (ids.ID, error) {
	promptText := promptui.Prompt{
		Label: "chainID",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("ID cannot be empty")
			}
			_, err := ids.FromString(input)
			return err
		},
	}
	bytes, err := os.ReadFile("/home/jaimip/clobvm/.uri")
	var rawID string
	if err == nil && len(bytes) > 0 {
		rawID = strings.Split(string(bytes), "/")[5]
	} else {
		rawID, err = promptText.Run()
	}
	if err != nil {
		return ids.Empty, err
	}
	rawID = strings.TrimSpace(rawID)
	id, err := ids.FromString(rawID)
	if err != nil {
		return ids.Empty, err
	}
	return id, nil
}

func promptString(label string) (string, error) {
	promptText := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("String cannot be empty")
			}
			return nil
		},
	}
	text, err := promptText.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), err
}

func promptURI() (string, error) {
	bytes, err := os.ReadFile("/home/jaimip/clobvm/.uri")
	var uri string
	if err == nil && len(bytes) > 0 {
		uri = strings.TrimSpace(string(bytes))
	} else {
		uri, err = promptString("uri")
	}
	return uri, err
}

func defaultActor() (ids.ID, crypto.PrivateKey, *auth.EIP712Factory, *rpc.JSONRPCClient, *crpc.JSONRPCClient, error) {
	return consts.ChainID, consts.PrivKey, auth.NewEIP712Factory(
			consts.PrivKey,
		), rpc.NewJSONRPCClient(
			consts.URI,
		), crpc.NewRPCClient(
			consts.URI,
			consts.ChainID,
			genesis.New(),
		), nil
}

func promptToken() (ids.ID, error) {
	promptText := promptui.Prompt{
		Label: "tokenID",
		Validate: func(input string) error {
			c := make([]byte, 32)
			copy(c, []byte(input))
			_, err := ids.ToID(c)
			return err
		},
	}
	token, err := promptText.Run()
	if err != nil {
		return ids.Empty, err
	}
	token = strings.TrimSpace(token)
	c := make([]byte, 32)
	copy(c, []byte(token))
	tokenID, err := ids.ToID(c)
	return tokenID, nil
}

func promptAddress(label string) (crypto.PublicKey, error) {
	promptText := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("Address cannot be empty")
			}
			_, err := crypto.ParseAddress("clob", input)
			return err
		},
	}
	recipient, err := promptText.Run()
	if err != nil {
		return crypto.EmptyPublicKey, err
	}
	recipient = strings.TrimSpace(recipient)
	return crypto.ParseAddress("clob", recipient)
}

func promptAmount(
	label string,
) (uint64, error) {
	promptText := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("Amount cannot be empty")
			}
			_, err := strconv.ParseUint(input, 10, 64)
			if err != nil {
				return err
			}
			return nil
		},
	}
	rawAmount, err := promptText.Run()
	if err != nil {
		return 0, err
	}
	rawAmount = strings.TrimSpace(rawAmount)
	return strconv.ParseUint(rawAmount, 10, 64)
}

func promptContinue() (bool, error) {
	promptText := promptui.Prompt{
		Label: "continue (y/n)",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("Continue cannot be empty")
			}
			lower := strings.ToLower(input)
			if lower == "y" || lower == "n" {
				return nil
			}
			return errors.New("Invalid input")
		},
	}
	rawContinue, err := promptText.Run()
	if err != nil {
		return false, err
	}
	cont := strings.ToLower(rawContinue)
	if cont == "n" {
		utils.Outf("{{red}}exiting...{{/}}\n")
		return false, nil
	}
	return true, nil
}

func promptBool(label string) (bool, error) {
	promptText := promptui.Prompt{
		Label: fmt.Sprintf("%s (y/n)", label),
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("Continue cannot be empty")
			}
			lower := strings.ToLower(input)
			if lower == "y" || lower == "n" {
				return nil
			}
			return errors.New("Invalid input")
		},
	}
	rawContinue, err := promptText.Run()
	if err != nil {
		return false, err
	}
	cont := strings.ToLower(rawContinue)
	if cont == "n" {
		return false, nil
	}
	return true, nil
}

func promptID(label string) (ids.ID, error) {
	promptText := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("ID cannot be empty")
			}
			_, err := ids.FromString(input)
			return err
		},
	}
	rawID, err := promptText.Run()
	if err != nil {
		return ids.Empty, err
	}
	rawID = strings.TrimSpace(rawID)
	id, err := ids.FromString(rawID)
	if err != nil {
		return ids.Empty, err
	}
	return id, nil
}