package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/auth"
	"github.com/jaimi-io/clobvm/genesis"
	crpc "github.com/jaimi-io/clobvm/rpc"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/jaimi-io/hypersdk/rpc"
	"github.com/jaimi-io/hypersdk/utils"
	"github.com/manifoldco/promptui"
)

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

func defaultActor() (ids.ID, crypto.PrivateKey, *auth.EIP712Factory, *rpc.JSONRPCClient, *crpc.JSONRPCClient, error) {
	// priv := crypto.PrivateKey(
	// 	[crypto.PrivateKeyLen]byte{
	// 		32, 241, 118, 222, 210, 13, 164, 128, 3, 18,
	// 		109, 215, 176, 215, 168, 171, 194, 181, 4, 11,
	// 		253, 199, 173, 240, 107, 148, 127, 190, 48, 164,
	// 		12, 48, 115, 50, 124, 153, 59, 53, 196, 150, 168,
	// 		143, 151, 235, 222, 128, 136, 161, 9, 40, 139, 85,
	// 		182, 153, 68, 135, 62, 166, 45, 235, 251, 246, 69, 8,
	// 	},
	// )
	priv, err := crypto.LoadKey("/home/jaimip/clobvm/test.pk")
	if err != nil {
		return ids.Empty, crypto.PrivateKey{}, nil, nil, nil, err
	}
	chainID, err := promptID("chainID")
	if err != nil {
		return ids.Empty, crypto.PrivateKey{}, nil, nil, nil, err
	}
	uri, err := promptString("uri")
	if err != nil {
		return ids.Empty, crypto.PrivateKey{}, nil, nil, nil, err
	}
	return chainID, priv, auth.NewEIP712Factory(
			priv,
		), rpc.NewJSONRPCClient(
			uri,
		), crpc.NewRPCClient(
			uri,
			chainID,
			genesis.New(),
		), nil
}

func promptToken(label string, allowNative bool) (ids.ID, error) {
	text := fmt.Sprintf("%s", label)
	promptText := promptui.Prompt{
		Label: text,
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

