package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/jaimi-io/clobvm/actions"
	"github.com/jaimi-io/clobvm/auth"
	"github.com/jaimi-io/clobvm/cmd/clob-cli/consts"
	"github.com/jaimi-io/clobvm/genesis"
	crpc "github.com/jaimi-io/clobvm/rpc"
	trpc "github.com/jaimi-io/clobvm/rpc"
	"github.com/jaimi-io/hypersdk/chain"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/jaimi-io/hypersdk/rpc"
	"github.com/jaimi-io/hypersdk/utils"
	hutils "github.com/jaimi-io/hypersdk/utils"
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

func promptToken(label string) (ids.ID, error) {
	promptText := promptui.Prompt{
		Label: fmt.Sprintf("%s tokenID", label),
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

func promptInt(
	label string,
) (int, error) {
	promptText := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("Amount cannot be empty")
			}
			amount, err := strconv.Atoi(input)
			if err != nil {
				return err
			}
			if amount <= 0 {
				return fmt.Errorf("%d must be > 0", amount)
			}
			return nil
		},
	}
	rawAmount, err := promptText.Run()
	if err != nil {
		return 0, err
	}
	rawAmount = strings.TrimSpace(rawAmount)
	return strconv.Atoi(rawAmount)
}

func getTokens() (ids.ID, ids.ID) {
	avaxID, _ := ids.FromString("VmwmdfVNQLiP1zJWmhaHipksKBAHmDZH5rZvdfCQfQ9peNx8a")
	usdcID, _ := ids.FromString("eaX7nEYVKiiFLEvRYQWmHixL9nwC1jFxsa1R75ipEchWBMKiG")
	return avaxID, usdcID
}

func submitDummy(
	ctx context.Context,
	cli *rpc.JSONRPCClient,
	tcli *trpc.JSONRPCClient,
	dest crypto.PublicKey,
	factory chain.AuthFactory,
) error {
	var (
		logEmitted bool
		txsSent    uint64
	)
	for ctx.Err() == nil {
		_, h, t, err := cli.Accepted(ctx)
		if err != nil {
			return err
		}
		dummyBlockAgeThreshold := int64(25)
		dummyHeightThreshold   := uint64(3)
		underHeight := h < dummyHeightThreshold
		if underHeight || time.Now().Unix()-t > dummyBlockAgeThreshold {
			if underHeight && !logEmitted {
				hutils.Outf(
					"{{yellow}}waiting for snowman++ activation (needed for AWM)...{{/}}\n",
				)
				logEmitted = true
			}
			parser, err := tcli.Parser(ctx)
			if err != nil {
				return err
			}
			avaxID, _ := getTokens()
			submit, _, _, err := cli.GenerateTransaction(ctx, parser, nil, &actions.Transfer{
				To:    dest,
				TokenID: avaxID,
				Amount: txsSent + 1, // prevent duplicate txs
			}, factory)
			if err != nil {
				return err
			}
			if err := submit(ctx); err != nil {
				return err
			}
			// if _, err := tcli.WaitForTransaction(ctx, tx.ID()); err != nil {
			// 	return err
			// }
			txsSent++
			time.Sleep(750 * time.Millisecond)
			continue
		}
		if logEmitted {
			hutils.Outf("{{yellow}}snowman++ activated{{/}}\n")
		}
		return nil
	}
	return ctx.Err()
}